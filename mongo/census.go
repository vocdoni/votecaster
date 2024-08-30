package mongo

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.vocdoni.io/dvote/types"
)

// AddCensus creates a new census document in the database.
func (ms *MongoStorage) AddCensus(censusID types.HexBytes, userFID uint64) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	census := Census{
		CensusID:  censusID.String(),
		CreatedBy: userFID,
	}
	_, err := ms.census.InsertOne(context.Background(), census)
	if err != nil {
		return fmt.Errorf("cannot insert census: %w", err)
	}

	return nil
}

// AddParticipantsToCensus updates a census document with participants and their associated values.
func (ms *MongoStorage) AddParticipantsToCensus(censusID types.HexBytes, participants map[string]struct {
	Weight        *big.Int
	Participation uint32
},
	fromTotalAddresses uint32, censusURI string,
) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	participantsString := map[string]string{}
	totalParticipants := uint32(0)
	totalWeight := new(big.Int)
	for k, v := range participants {
		participantsString[k] = fmt.Sprintf("%s:%d", v.Weight.String(), v.Participation)
		totalParticipants += v.Participation
		totalWeight.Add(totalWeight, v.Weight)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	update := bson.M{
		"$set": bson.M{
			"participants":          participantsString,
			"fromTotalAddresses":    fromTotalAddresses,
			"fromTotalParticipants": totalParticipants,
			"totalWeight":           totalWeight.String(),
			"url":                   censusURI,
		},
	}

	_, err := ms.census.UpdateOne(ctx, bson.M{"_id": censusID.String()}, update)
	if err != nil {
		return fmt.Errorf("cannot update census: %w", err)
	}

	return nil
}

// Census retrieves a census document based on its ID.
func (ms *MongoStorage) Census(censusID types.HexBytes) (Census, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var census Census
	err := ms.census.FindOne(ctx, bson.M{"_id": censusID.String()}).Decode(&census)
	if err != nil {
		return Census{}, fmt.Errorf("cannot find census: %w", err)
	}

	return census, nil
}

// ParticipantParticipation retrieves the participation value for a given participant in a census.
func (ms *MongoStorage) ParticipantParticipation(censusRoot types.HexBytes, fid uint64) (uint32, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := ms.userData(fid)
	if err != nil {
		return 0, fmt.Errorf("could not get user %d from database", fid)
	}

	var census Census
	err = ms.census.FindOne(ctx, bson.M{"root": censusRoot.String()}).Decode(&census)
	if err != nil {
		return 0, fmt.Errorf("cannot find Census with root: %w", err)
	}
	participant, ok := census.Participants[user.Username]
	if !ok {
		// just return 1 as fallback
		return 1, nil
	}
	weightAndParticipation := strings.Split(participant, ":")
	if len(weightAndParticipation) != 2 {
		return 1, nil
	}
	participation, err := strconv.ParseUint(weightAndParticipation[1], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("cannot parse participation: %w", err)
	}
	return uint32(participation), nil
}

// ParticipantsByWeight retrieves the top N participants by weight in a census.
func (ms *MongoStorage) ParticipantsByWeight(electionID types.HexBytes, n int) (map[string]*big.Int, error) {
	census, err := ms.censusFromElection(electionID)
	if err != nil {
		return nil, err
	}
	keys := []string{}
	for k := range census.Participants {
		keys = append(keys, strings.Split(k, ":")[0])
	}
	// sort the keys by the value of the participants, descending
	sort.SliceStable(keys, func(i, j int) bool {
		iWeight, _ := new(big.Int).SetString(census.Participants[keys[i]], 10)
		jWeight, _ := new(big.Int).SetString(census.Participants[keys[j]], 10)
		return iWeight.Cmp(jWeight) > 0
	})
	if n > len(keys) {
		n = len(keys)
	}
	users := map[string]*big.Int{}
	for _, username := range keys[:n] {
		users[username], _ = new(big.Int).SetString(census.Participants[username], 10)
	}
	return users, nil
}

// SetRootForCensus updates the root for a given census document.
// If the census does not exist, it returns nil without error.
func (ms *MongoStorage) SetRootForCensus(censusID, root types.HexBytes) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{"root": root.String()}}
	_, err := ms.census.UpdateOne(ctx, bson.M{"_id": censusID.String()}, update)
	if err != nil {
		return fmt.Errorf("cannot update Census root: %w", err)
	}

	return nil
}

// CensusFromRoot retrieves a Census document by its root.
func (ms *MongoStorage) CensusFromRoot(root types.HexBytes) (*Census, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var census Census
	err := ms.census.FindOne(ctx, bson.M{"root": root.String()}).Decode(&census)
	if err != nil {
		return nil, fmt.Errorf("cannot find Census with root: %w", err)
	}

	return &census, nil
}

// CensusFromRoot retrieves a Census document by its root.
func (ms *MongoStorage) CensusFromElection(electionID types.HexBytes) (*Census, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	return ms.censusFromElection(electionID)
}

// censusFromRoot retrieves a Census document by its root. It does not adquire the keysLock.
func (ms *MongoStorage) censusFromElection(electionID types.HexBytes) (*Census, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var census Census
	err := ms.census.FindOne(ctx, bson.M{"electionId": electionID.String()}).Decode(&census)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrElectionUnknown
		}
		return nil, fmt.Errorf("cannot find Census with electionID: %w", err)
	}

	return &census, nil
}

// SetElectionIdForCensusRoot updates the electionId for a given census document.
// It will only be set if the electionId field is empty.
// If the census does not exist, it returns nil without error.
func (ms *MongoStorage) SetElectionIdForCensusRoot(root, electionID types.HexBytes) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare the filter to find the first census document with the given root and an empty electionId
	filter := bson.M{
		"root":       root.String(),
		"electionId": "", // Match only documents where electionId is an empty string
	}

	// Prepare the update document to set the new ElectionID
	update := bson.M{"$set": bson.M{"electionId": electionID.String()}}

	// Execute the update operation, limiting to the first match
	_, err := ms.census.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("cannot update ElectionID for Census with root %s: %w", root.String(), err)
	}

	return nil
}

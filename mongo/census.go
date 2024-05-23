package mongo

import (
	"context"
	"fmt"
	"math/big"
	"sort"
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
func (ms *MongoStorage) AddParticipantsToCensus(censusID types.HexBytes, participants map[string]*big.Int,
	fromTotalAddresses uint32, totalWeight *big.Int, censusURI string,
) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	participantsString := map[string]string{}
	for k, v := range participants {
		participantsString[k] = v.String()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	update := bson.M{
		"$set": bson.M{
			"participants":       participantsString,
			"fromTotalAddresses": fromTotalAddresses,
			"totalWeight":        totalWeight.String(),
			"url":                censusURI,
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

// ParticipantsByWeight retrieves the top N participants by weight in a census.
func (ms *MongoStorage) ParticipantsByWeight(censusID types.HexBytes, n int) (map[string]*big.Int, error) {
	census, err := ms.censusFromElection(censusID)
	if err != nil {
		return nil, err
	}
	keys := []string{}
	for k := range census.Participants {
		keys = append(keys, k)
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

// SetElectionIdForCensusRoot updates the ElectionID for a given census document by its root.
// If the root is not found, it returns nil, indicating no error occurred.
func (ms *MongoStorage) SetElectionIdForCensusRoot(root, electionID types.HexBytes) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare the update document to set the new ElectionID
	update := bson.M{"$set": bson.M{"electionId": electionID.String()}}

	// Execute the update operation
	result, err := ms.census.UpdateOne(ctx, bson.M{"root": root.String()}, update)
	if err != nil {
		return fmt.Errorf("cannot update ElectionID for Census with root %s: %w", root.String(), err)
	}

	// If the root is not found, MatchedCount will be 0. We treat this as a no-op success.
	if result.MatchedCount == 0 {
		return nil
	}

	return nil
}

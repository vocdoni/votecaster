package mongo

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
func (ms *MongoStorage) AddParticipantsToCensus(censusID types.HexBytes, participants map[string]string,
	fromTotalAddresses uint32, totalWeight *big.Int, tokenDecimals uint32, censusURI string,
) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"participants":       participants,
			"fromTotalAddresses": fromTotalAddresses,
			"tokenDecimals":      tokenDecimals,
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

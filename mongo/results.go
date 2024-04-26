package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

// AddFinalResults adds the final results of an election in PNG format.
// It performs and upsert operation, so it will update the results if they already exist.
func (ms *MongoStorage) AddFinalResults(electionID types.HexBytes, finalPNG []byte, choices, votes []string) error {
	results := &Results{
		ElectionID: electionID.String(),
		FinalPNG:   finalPNG,
		Choices:    choices,
		Votes:      votes,
		Finalized:  true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.ReplaceOptions{}
	opts.Upsert = new(bool)
	*opts.Upsert = true
	_, err := ms.results.ReplaceOne(ctx, bson.M{"_id": results.ElectionID}, results, &opts)
	if err != nil {
		return fmt.Errorf("cannot update object: %w", err)
	}
	log.Debugw("stored PNG results", "electionID", electionID.String())
	return nil
}

// Results retrieves the final results of an election.
func (ms *MongoStorage) Results(electionID types.HexBytes) (*Results, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var results Results
	err := ms.results.FindOne(ctx, bson.M{"_id": electionID.String()}).Decode(&results)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrElectionUnknown
		}
		return nil, fmt.Errorf("error retrieving results for electionID %s: %w", electionID.String(), err)
	}

	return &results, nil
}

// FinalResultsPNG returns the final results of an election in PNG format.
// It returns nil if the results image is not found.
func (ms *MongoStorage) FinalResultsPNG(electionID types.HexBytes) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := ms.results.FindOne(ctx, bson.M{"_id": electionID.String()})
	if result == nil {
		return nil
	}
	var results Results
	if err := result.Decode(&results); err != nil {
		return nil
	}
	return results.FinalPNG
}

// ElectionsWithoutResults returns a list of election IDs where results are not finalized.
func (ms *MongoStorage) ElectionsWithoutResults() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create the filter for finding results where "finalized" is false
	filter := bson.M{"finalized": false}

	// Define the options to only retrieve the "electionId" field
	opts := options.Find().SetProjection(bson.M{"_id": 1})

	// Execute the find operation
	cursor, err := ms.results.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find results: %w", err)
	}
	defer cursor.Close(ctx)

	var electionIDs []string
	for cursor.Next(ctx) {
		var result struct {
			ElectionID string `bson:"_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode result: %w", err)
		}
		electionIDs = append(electionIDs, result.ElectionID)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return electionIDs, nil
}

// SetPartialResults sets or updates the choices and votes for an election result only if it is not finalized.
// It performs an upsert operation, so it will create the result entry if it does not exist and is not finalized.
func (ms *MongoStorage) SetPartialResults(electionID types.HexBytes, choices, votes []string) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare the update fields
	update := bson.M{
		"$set": bson.M{
			"title": choices,
			"votes": votes,
		},
	}

	// Prepare the filter with an additional check for not finalized
	filter := bson.M{
		"_id":       electionID.String(),
		"finalized": bson.M{"$ne": true},
	}

	// Options to enable upsert, i.e., insert if not exists
	opts := options.UpdateOptions{}
	opts.Upsert = new(bool)
	*opts.Upsert = true

	// Execute the update operation
	_, err := ms.results.UpdateOne(ctx, filter, update, &opts)
	if err != nil {
		return fmt.Errorf("cannot update choices and votes: %w", err)
	}
	return nil
}

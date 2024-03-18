package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

// AddFinalResults adds the final results of an election in PNG format.
// It performs and upsert operation, so it will update the results if they already exist.
func (ms *MongoStorage) AddFinalResults(electionID types.HexBytes, finalPNG []byte) error {
	results := &Results{
		ElectionID: electionID.String(),
		FinalPNG:   finalPNG,
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

// ElectionsWithoutResults returns a list of election IDs that do not have a corresponding entry in the results collection.
func (ms *MongoStorage) ElectionsWithoutResults(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// Define the aggregation pipeline
	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "results",
				"localField":   "_id",
				"foreignField": "_id",
				"as":           "result",
			},
		},
		{
			"$match": bson.M{"result": bson.M{"$size": 0}}, // Filter elections without results
		},
		{
			"$project": bson.M{"_id": 1}, // Only return the election IDs
		},
	}

	cursor, err := ms.elections.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var elections []struct {
		ID string `bson:"_id"`
	}
	if err := cursor.All(ctx, &elections); err != nil {
		return nil, err
	}

	var electionIDs []string
	for _, e := range elections {
		electionIDs = append(electionIDs, e.ID)
	}

	return electionIDs, nil
}

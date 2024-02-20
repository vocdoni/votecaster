package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

// AddFinalResults adds the final results of an election in PNG format.
func (ms *MongoStorage) AddFinalResults(electionID types.HexBytes, finalPNG []byte) error {
	results := &Results{
		ElectionID: electionID.String(),
		FinalPNG:   finalPNG,
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	_, err := ms.results.InsertOne(ctx2, results)
	log.Debugw("stored PNG results", "electionID", electionID.String())
	return err
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

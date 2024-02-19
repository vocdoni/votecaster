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
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	// Check if the election results already exist, if so, just return
	// Since we don't adquire the lock until the call of this function,
	// it's possible that the election results are already stored.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := ms.results.FindOne(ctx, bson.M{"_id": electionID.String()})
	if result != nil {
		return nil
	}

	results := &Results{
		ElectionID: electionID.String(),
		FinalPNG:   finalPNG,
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	_, err := ms.results.InsertOne(ctx2, results)
	return err
}

// FinalResultsPNG returns the final results of an election in PNG format.
// It returns nil if the results image is not found.
func (ms *MongoStorage) FinalResultsPNG(electionID types.HexBytes) []byte {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := ms.results.FindOne(ctx, bson.M{"_id": electionID.String()})
	if result == nil {
		return nil
	}
	var results Results
	if err := result.Decode(&results); err != nil {
		log.Warn(err)
		return nil
	}
	return results.FinalPNG
}

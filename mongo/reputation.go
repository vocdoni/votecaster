package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TotalVotesForUserElections calculates the total number of votes casted on elections created by the user.
func (ms *MongoStorage) TotalVotesForUserElections(userID uint64) (uint64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use aggregation to calculate the total votes
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"userId": userID}}},
		bson.D{{Key: "$group", Value: bson.M{"_id": "$userId", "totalVotes": bson.M{"$sum": "$castedVotes"}}}},
	}

	cursor, err := ms.elections.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			TotalVotes uint64 `bson:"totalVotes"`
		}
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.TotalVotes, nil
	}

	// If there's no result, it means no elections or votes were found.
	return 0, nil
}

// CommunitiesCountForUser calculates the number of communities where the
// user is an admin.
func (ms *MongoStorage) CommunitiesCountForUser(userID uint64) (uint64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"owners": userID}}},
		bson.D{{Key: "$count", Value: "count"}},
	}

	cursor, err := ms.communities.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			Count uint64 `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.Count, nil
	}

	return 0, nil
}

// SetReputationForUser updates the reputation for a given user ID.
func (ms *MongoStorage) SetReputationForUser(userID uint64, reputation uint32) error {
	return ms.updateUserAccessProfile(userID, bson.M{"$set": bson.M{"reputation": reputation}})
}

func (ms *MongoStorage) DetailedUserReputation(userID uint64) (*UserReputation, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	return ms.userReputation(userID)
}

func (ms *MongoStorage) SetDetailedReputationForUser(userID uint64, reputation *UserReputation) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	reputation.UserID = userID
	return ms.updateUserReputation(reputation)
}

func (ms *MongoStorage) userReputation(userID uint64) (*UserReputation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var profile UserReputation
	if err := ms.reputations.FindOne(ctx, bson.M{"_id": userID}).Decode(&profile); err != nil {
		return nil, ErrUserUnknown
	}
	return &profile, nil
}

func (ms *MongoStorage) updateUserReputation(reputation *UserReputation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateDoc, err := dynamicUpdateDocument(reputation, nil)
	if err != nil {
		return err
	}

	opts := options.Update().SetUpsert(true)
	_, err = ms.reputations.UpdateOne(ctx, bson.M{"_id": reputation.UserID}, updateDoc, opts)
	return err
}

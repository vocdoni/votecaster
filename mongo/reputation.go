package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.vocdoni.io/dvote/log"
)

// ReputationsIterator iterates over available reputations and sends them to
// the provided channel.
func (ms *MongoStorage) Reputations() ([]*Reputation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Executing the find operation with the specified filter and options
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	cur, err := ms.reputations.Find(ctx, bson.M{
		"$or": []bson.M{
			{"participation": bson.M{"$gt": 0}},
			{"censusSize": bson.M{"$gt": 0}},
			{"totalReputation": bson.M{"$gt": 0}},
		},
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var reputations []*Reputation
	for cur.Next(ctx) {
		reputation := &Reputation{}
		if err := cur.Decode(reputation); err != nil {
			log.Warn(err)
			continue
		}
		reputations = append(reputations, reputation)
	}
	return reputations, nil
}

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

// DetailedUserReputation method return the reputation of a user based on the
// user ID. It returns the detailed reputation information and values from the
// database.
func (ms *MongoStorage) DetailedUserReputation(userID uint64) (*Reputation, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	return ms.userReputation(userID)
}

// DetailedCommunityReputation method return the reputation of a community based
// on the community ID. It returns the detailed reputation information and
// values from the database.
func (ms *MongoStorage) DetailedCommunityReputation(communityID string) (*Reputation, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	return ms.communityReputation(communityID)
}

// SetDetailedReputationForUser method updates the detailed reputation for a
// given user ID. It overwrites the previous reputation values with the provided
// values, if some values are not provided, they will keep the previous values
// if they exist.
func (ms *MongoStorage) SetDetailedReputationForUser(userID uint64, reputation *Reputation) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	reputation.UserID = userID
	return ms.updateReputation(reputation)
}

// SetDetailedReputationForCommunity method updates the detailed reputation for
// a given community ID. It overwrites the previous reputation values with the
// provided values, if some values are not provided, they will keep the previous
// values if they exist.
func (ms *MongoStorage) SetDetailedReputationForCommunity(communityID string, reputation *Reputation) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	reputation.CommunityID = communityID
	return ms.updateReputation(reputation)
}

// CommunitiesReputationByParticipationAndCensusSize method returns the top 10
// communities reputations sorted by participation and census size.
func (ms *MongoStorage) CommunitiesReputationByParticipationAndCensusSize() ([]*Reputation, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	// get the top 10 communities sorted by participation and census size
	limit := int64(10)
	opts := options.FindOptions{Limit: &limit}
	opts.SetSort(bson.D{{Key: "participation", Value: -1}, {Key: "censusSize", Value: -1}})
	filter := bson.M{"communityID": bson.M{"$exists": true}}
	// find the communities reputation data
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.reputations.Find(ctx, filter, &opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	// iterate over the results to get user or community data
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	var reputations []*Reputation
	for cur.Next(ctx2) {
		var rep Reputation
		if err := cur.Decode(&rep); err != nil {
			return nil, err
		}
		reputations = append(reputations, &rep)
	}
	return reputations, nil
}

func (ms *MongoStorage) userReputation(userID uint64) (*Reputation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var profile Reputation
	if err := ms.reputations.FindOne(ctx, bson.M{"userID": userID}).Decode(&profile); err != nil {
		return nil, ErrUserUnknown
	}
	return &profile, nil
}

func (ms *MongoStorage) communityReputation(communityID string) (*Reputation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var profile Reputation
	if err := ms.reputations.FindOne(ctx, bson.M{"communityID": communityID}).Decode(&profile); err != nil {
		return nil, fmt.Errorf("community '%s' reputation not found: %w", communityID, err)
	}
	return &profile, nil
}

func (ms *MongoStorage) updateReputation(reputation *Reputation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateDoc, err := dynamicUpdateDocument(reputation, nil)
	if err != nil {
		return err
	}
	filter := bson.M{"userID": reputation.UserID}
	if reputation.CommunityID != "" {
		filter = bson.M{"communityID": reputation.CommunityID}
	}
	opts := options.Update().SetUpsert(true)
	_, err = ms.reputations.UpdateOne(ctx, filter, updateDoc, opts)
	return err
}

package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.vocdoni.io/dvote/log"
)

// UsersByElectionNumber returns the list of users ordered by the number of elections they have created.
func (ms *MongoStorage) UsersByElectionNumber() ([]UserRanking, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	limit := int64(10)
	opts := options.FindOptions{Limit: &limit}
	opts.SetSort(bson.M{"electionCount": -1})
	opts.SetProjection(bson.M{"_id": true, "username": true, "displayname": true, "electionCount": true})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.users.Find(ctx, bson.M{}, &opts)
	if err != nil {
		return nil, err
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	var ranking []UserRanking
	for cur.Next(ctx2) {
		user := User{}
		err := cur.Decode(&user)
		if err != nil {
			log.Warn(err)
		}
		ranking = append(ranking, UserRanking{
			FID:         user.UserID,
			Username:    user.Username,
			Displayname: user.Displayname,
			Count:       user.ElectionCount,
		})
	}
	return ranking, nil
}

// UsersByVoteNumber returns the list of users ordered by the number of votes they have casted.
func (ms *MongoStorage) UsersByVoteNumber() ([]UserRanking, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	limit := int64(10)
	opts := options.FindOptions{Limit: &limit}
	opts.SetSort(bson.M{"castedVotes": -1})
	opts.SetProjection(bson.M{"_id": true, "username": true, "displayname": true, "castedVotes": true})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.users.Find(ctx, bson.M{}, &opts)
	if err != nil {
		return nil, err
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	var ranking []UserRanking
	for cur.Next(ctx2) {
		user := User{}
		err := cur.Decode(&user)
		if err != nil {
			log.Warn(err)
		}
		ranking = append(ranking, UserRanking{
			FID:         user.UserID,
			Username:    user.Username,
			Displayname: user.Displayname,
			Count:       user.CastedVotes,
		})
	}
	return ranking, nil
}

// ElectionsByVoteNumber returns the list of elections ordered by the number of votes casted within the last 60 days.
func (ms *MongoStorage) ElectionsByVoteNumber() ([]*Election, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	limit := int64(10)
	opts := options.FindOptions{Limit: &limit}
	opts.SetSort(bson.M{"castedVotes": -1})
	opts.SetProjection(bson.M{"_id": true, "castedVotes": true, "userId": true, "question": true})

	// Calculate the date 60 days ago
	timeLimit := time.Now().AddDate(0, 0, -60)

	// Create the filter for elections within the last 60 days
	filter := bson.M{
		"createdTime": bson.M{"$gte": timeLimit},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.elections.Find(ctx, filter, &opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var ranking []*Election
	for cur.Next(ctx) {
		election := &Election{}
		if err := cur.Decode(election); err != nil {
			log.Warn(err)
			continue
		}
		if election.CastedVotes == 0 {
			continue
		}
		ranking = append(ranking, election)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return ranking, nil
}

// UserByReputation returns the list of users ordered by their reputation score.
func (ms *MongoStorage) UserByReputation() ([]UserRanking, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	// first get users fid from userAccessProfiles collection sorted by the
	// the reputation score.
	limit := int64(10)
	opts := options.FindOptions{Limit: &limit}
	opts.SetSort(bson.M{"reputation": -1})
	opts.SetProjection(bson.M{"fid": "$_id", "count": "$reputation"})
	// get top 10 fids sorted by reputation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.userAccessProfiles.Find(ctx, bson.M{}, &opts)
	if err != nil {
		return nil, err
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()
	var ranking []UserRanking
	for cur.Next(ctx2) {
		var user UserRanking
		if err := cur.Decode(&user); err != nil {
			log.Warn(err)
			continue
		}
		userData, err := ms.User(user.FID)
		if err != nil {
			log.Warn(err)
			continue
		}
		user.Username = userData.Username
		user.Displayname = userData.Displayname
		ranking = append(ranking, user)
	}
	return ranking, nil
}

// LastCreatedElections returns the last created elections.
func (ms *MongoStorage) LastCreatedElections(count int) ([]*Election, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the last N created elections, ordered by CreatedTime descending
	opts := options.Find().SetSort(bson.D{{Key: "createdTime", Value: -1}}).SetLimit(int64(count))
	cursor, err := ms.elections.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve elections: %w", err)
	}
	defer cursor.Close(ctx)

	var elections []*Election
	for cursor.Next(ctx) {
		var election Election
		if err := cursor.Decode(&election); err != nil {
			return nil, fmt.Errorf("failed to decode election: %w", err)
		}
		elections = append(elections, &election)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return elections, nil
}

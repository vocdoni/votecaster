package mongo

import (
	"context"
	"encoding/hex"
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
	opts.SetProjection(bson.M{"_id": true, "username": true, "electionCount": true})
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
	opts.SetProjection(bson.M{"_id": true, "username": true, "castedVotes": true})
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

// ElectionsByVoteNumber returns the list elections ordered by the number of votes casted.
func (ms *MongoStorage) ElectionsByVoteNumber() ([]ElectionRanking, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	limit := int64(10)
	opts := options.FindOptions{Limit: &limit}
	opts.SetSort(bson.M{"castedVotes": -1})
	opts.SetProjection(bson.M{"_id": true, "castedVotes": true, "userId": true})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.elections.Find(ctx, bson.M{}, &opts)
	if err != nil {
		return nil, err
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	var ranking []ElectionRanking
	for cur.Next(ctx2) {
		election := Election{}
		err := cur.Decode(&election)
		if err != nil {
			log.Warn(err)
		}
		if election.CastedVotes == 0 {
			continue
		}
		eID, err := hex.DecodeString(election.ElectionID)
		if err != nil {
			log.Warn(err)
			continue
		}

		title := election.Question
		if title == "" {
			// if election question is not stored in the database, try to get it from the API
			// THIS CODE CAN BE REMOVE AT SOME POINT, WHEN DB IS POPULATED
			electionInfo, err := ms.election(eID)
			if err != nil {
				log.Warn(err)
			} else {
				if electionInfo != nil && electionInfo.Metadata != nil {
					title = electionInfo.Metadata.Title["default"]
				}
			}
		}

		username := ""
		user, err := ms.userData(election.UserID)
		if err != nil {
			log.Warn(err)
		} else {
			username = user.Username
		}

		ranking = append(ranking, ElectionRanking{
			ElectionID:           election.ElectionID,
			VoteCount:            election.CastedVotes,
			CreatedByFID:         election.UserID,
			CreatedByDisplayname: user.Displayname,
			Title:                title,
			CreatedByUsername:    username,
		})

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

		// if election question is not stored in the database, try to get it from the API
		// THIS CODE CAN BE REMOVE AT SOME POINT, WHEN DB IS POPULATED
		if election.Question == "" {
			eID, err := hex.DecodeString(election.ElectionID)
			if err != nil {
				log.Warn(err)
				continue
			}
			electionInfo, err := ms.election(eID)
			if err != nil {
				log.Warn(err)
			} else {
				if electionInfo != nil && electionInfo.Metadata != nil {
					election.Question = electionInfo.Metadata.Title["default"]
				}
			}
		}

		elections = append(elections, &election)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return elections, nil
}

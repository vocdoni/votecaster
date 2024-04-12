package mongo

import (
	"context"
	"encoding/hex"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
			Count:       float64(user.ElectionCount),
		})
	}
	return ranking, nil
}

// UsersByVoteNumber returns the list of users ordered by the number of votes they have casted.
func (ms *MongoStorage) UsersByVoteNumber() ([]UserRanking, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	// aggregation pipeline to get the users sorted by their reputation, the
	// reputation is stored in the userAccessProfiles collection and is
	// linked to the users collection by the _id field.
	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "userAccessProfiles"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "accessProfile"},
		}}},
		{{Key: "$unwind", Value: "$accessProfile"}},
		{{Key: "$project", Value: bson.D{
			{Key: "fid", Value: "$_id"},
			{Key: "username", Value: 1},
			{Key: "displayname", Value: 1},
			{Key: "count", Value: "$accessProfile.reputation"},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.users.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	var ranking []UserRanking
	for cur.Next(ctx2) {
		var user UserRanking
		err := cur.Decode(&user)
		if err != nil {
			log.Warn(err)
		}
		ranking = append(ranking, user)
	}
	return ranking, nil
}

// UsersByVoteNumber returns the list of top 10 users ordered by the ratio
// between the number of votes casted and the number of elections created by
// them.
func (ms *MongoStorage) UsersByPollsVotesRatio() ([]UserRanking, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	// aggregation pipeline to get the users sorted by their ratio between
	// casted votes and created elections.
	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "elections"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "userId"},
			{Key: "as", Value: "createdElections"},
		}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "totalElections", Value: bson.D{{Key: "$size", Value: "$createdElections"}}},
			{Key: "totalVoters", Value: bson.D{{Key: "$sum", Value: "$createdElections.castedVotes"}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "fid", Value: "$_id"},
			{Key: "username", Value: 1},
			{Key: "displayname", Value: 1},
			{Key: "count", Value: bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$eq", Value: bson.A{"$totalElections", 0}}},
				0.0, // Avoid division by zero
				bson.D{{Key: "$divide", Value: bson.A{"$totalVoters", "$totalElections"}}},
			}}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.users.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	var ranking []UserRanking
	for cur.Next(ctx2) {
		var user UserRanking
		err := cur.Decode(&user)
		if err != nil {
			log.Warn(err)
		}
		ranking = append(ranking, user)
	}
	for _, user := range ranking {
		log.Infof("User %s: %f", user.Username, user.Count)
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
		user, err := ms.getUserData(election.UserID)
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
			Count:       float64(user.CastedVotes),
		})
	}
	return ranking, nil
}

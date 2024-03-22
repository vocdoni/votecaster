package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserData holds the user's data for calculating reputation.
type UserData struct {
	FollowersCount                uint64 `json:"followersCount"`
	ElectionsCreated              uint64 `json:"electionsCreated"`
	CastedVotes                   uint64 `json:"castedVotes"`
	VotesCastedOnCreatedElections uint64 `json:"participationAchievement"`
}

// calculateReputation calculates the user's reputation based on predefined criteria.
func calculateReputation(user *UserData) uint32 {
	reputation := 0.0

	// Calculate FollowersCount score (up to 10 points, max 20000 followers)
	reputation += float64(user.FollowersCount) / 2000

	// Calculate ElectionsCreated score (up to 10 points, max 100 elections)
	reputation += float64(user.ElectionsCreated) / 10

	// Calculate CastedVotes score (up to 30 points, max 120 votes)
	reputation += float64(user.CastedVotes) / 4

	// Calculate VotesCastedOnCreatedElections score (up to 50 points, max 1000 votes)
	reputation += float64(user.VotesCastedOnCreatedElections) / 20

	// Ensure the reputation does not exceed 100
	if reputation > 100 {
		reputation = 100
	}

	return uint32(reputation)
}

// UpdateAndGetReputationForUser updates the user's reputation based on their activities and returns the new reputation.
func (ms *MongoStorage) UpdateAndGetReputationForUser(userID uint64) (uint32, *UserData, error) {
	// Fetch the user data
	user, err := ms.User(userID)
	if err != nil {
		return 0, nil, ErrUserUnknown
	}

	// Fetch the total votes cast on elections created by the user
	totalVotes, err := ms.getTotalVotesForUserElections(userID)
	if err != nil {
		return 0, nil, fmt.Errorf("error fetching total votes for user elections: %w", err)
	}

	userData := UserData{
		FollowersCount:                user.Followers,
		ElectionsCreated:              user.ElectionCount,
		CastedVotes:                   user.CastedVotes,
		VotesCastedOnCreatedElections: totalVotes,
	}

	// Calculate the new reputation
	newReputation := calculateReputation(&userData)

	// Update the user's reputation in the database
	if err := ms.SetReputationForUser(userID, newReputation); err != nil {
		return 0, nil, fmt.Errorf("error updating user reputation: %w", err)
	}

	return newReputation, &userData, nil
}

// getTotalVotesForUserElections calculates the total number of votes casted on elections created by the user.
func (ms *MongoStorage) getTotalVotesForUserElections(userID uint64) (uint64, error) {
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

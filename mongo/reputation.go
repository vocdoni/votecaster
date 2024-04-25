package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	maxFollowersReputation = 10
	maxElectionsReputation = 10
	maxVotesReputation     = 25
	maxCastedReputation    = 45
	maxCommunityReputation = 10
	maxReputation          = 100
)

// UserData holds the user's data for calculating reputation.
type UserData struct {
	FollowersCount                uint64 `json:"followersCount"`
	ElectionsCreated              uint64 `json:"electionsCreated"`
	CastedVotes                   uint64 `json:"castedVotes"`
	VotesCastedOnCreatedElections uint64 `json:"participationAchievement"`
	CommunitiesCount              uint64 `json:"communitiesCount"`
}

// calculateReputation calculates the user's reputation based on predefined criteria.
func calculateReputation(user *UserData) uint32 {
	reputation := 0.0
	// Calculate FollowersCount score (up to 10 points, max 20000 followers)
	if followersRep := float64(user.FollowersCount) / 2000; followersRep <= maxFollowersReputation {
		reputation += followersRep
	} else {
		reputation += maxFollowersReputation
	}
	// Calculate ElectionsCreated score (up to 10 points, max 100 elections)
	if electionsRep := float64(user.ElectionsCreated) / 10; electionsRep <= maxElectionsReputation {
		reputation += electionsRep
	} else {
		reputation += maxElectionsReputation
	}
	// Calculate CastedVotes score (up to 30 points, max 120 votes)
	if votesRep := float64(user.CastedVotes) / 4; votesRep <= maxVotesReputation {
		reputation += votesRep
	} else {
		reputation += maxVotesReputation
	}
	// Calculate VotesCastedOnCreatedElections score (up to 50 points, max 1000 votes)
	if castedRep := float64(user.VotesCastedOnCreatedElections) / 20; castedRep <= maxCastedReputation {
		reputation += castedRep
	} else {
		reputation += maxCastedReputation
	}
	// Calculate CommunitiesCount score (up to 10 points, max 5 communities)
	if comRep := float64(user.CommunitiesCount) * 2; comRep <= maxCommunityReputation {
		reputation += comRep
	} else {
		reputation += maxCommunityReputation
	}
	// Ensure the reputation does not exceed 100
	if reputation > maxReputation {
		reputation = maxReputation
	}
	return uint32(reputation)
}

// UpdateAndGetReputationForUser updates the user's reputation based on their activities and returns the new reputation.
func (ms *MongoStorage) UpdateAndGetReputationForUser(userID uint64) (uint32, *UserData, error) {
	// Fetch the user data
	user, err := ms.User(userID)
	if err != nil {
		if errors.Is(err, ErrUserUnknown) {
			// If the user is not found, create a new user with blank data
			if err := ms.AddUser(userID, "", "", []string{}, []string{}, "", 0); err != nil {
				return 0, nil, fmt.Errorf("error adding user: %w", err)
			}
			if err := ms.SetReputationForUser(userID, 0); err != nil {
				return 0, nil, fmt.Errorf("error setting user reputation: %w", err)
			}
			return 0, &UserData{}, nil
		}
		return 0, nil, fmt.Errorf("error fetching user: %w", err)
	}
	// Fetch the total votes cast on elections created by the user
	totalVotes, err := ms.getTotalVotesForUserElections(userID)
	if err != nil {
		return 0, nil, fmt.Errorf("error fetching total votes for user elections: %w", err)
	}
	// Fetch the number of communities where the user is an admin
	communitiesCount, err := ms.getCommunitiesCountForUser(userID)
	if err != nil {
		return 0, nil, fmt.Errorf("error fetching communities count for user: %w", err)
	}
	userData := UserData{
		FollowersCount:                user.Followers,
		ElectionsCreated:              user.ElectionCount,
		CastedVotes:                   user.CastedVotes,
		VotesCastedOnCreatedElections: totalVotes,
		CommunitiesCount:              communitiesCount,
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

// getCommunitiesCountForUser calculates the number of communities where the
// user is an admin.
func (ms *MongoStorage) getCommunitiesCountForUser(userID uint64) (uint64, error) {
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

package mongo

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

func (ms *MongoStorage) AddElection(
	electionID types.HexBytes,
	userFID uint64,
	source string,
	question string,
	usersCount, usersCountInitial, tokenDecimals uint32) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	election := Election{
		UserID:                userFID,
		ElectionID:            electionID.String(),
		CreatedTime:           time.Now(),
		Source:                source,
		FarcasterUserCount:    usersCount,
		InitialAddressesCount: usersCountInitial,
		Question:              question,
		ElectionMeta: ElectionMeta{
			CensusERC20TokenDecimals: tokenDecimals,
		},
	}
	log.Infow("added new election", "electionID", electionID.String(), "userID", userFID, "question", question)
	return ms.addElection(&election)
}

// ElectionsByUser returns all the elections created by the user with the FID
// provided
func (ms *MongoStorage) ElectionsByUser(userFID uint64) ([]ElectionRanking, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := ms.elections.Find(ctx, bson.M{"userId": userFID})
	if err != nil {
		log.Warn(err)
		return nil, ErrElectionUnknown
	}
	defer cursor.Close(ctx)
	elections := []ElectionRanking{}
	for cursor.Next(ctx) {
		election := Election{}
		if err := cursor.Decode(&election); err != nil {
			log.Warn(err)
			continue
		}
		bElectionID, err := hex.DecodeString(election.ElectionID)
		if err != nil {
			log.Warn(err)
			continue
		}
		info, err := ms.election(bElectionID)
		if err != nil {
			log.Warn(err)
			continue
		}
		if info == nil || info.Metadata == nil || info.Metadata.Title == nil {
			log.Warn("no title found in election metadata")
			continue
		}
		user, err := ms.getUserData(election.UserID)
		if err != nil {
			log.Warn(err)
			continue
		}
		elections = append(elections, ElectionRanking{
			ElectionID:        election.ElectionID,
			Title:             info.Metadata.Title["default"],
			VoteCount:         election.CastedVotes,
			CreatedByFID:      election.UserID,
			CreatedByUsername: user.Username,
		})
	}
	return elections, nil
}

func (ms *MongoStorage) addElection(election *Election) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.elections.InsertOne(ctx, election)
	return err
}

func (ms *MongoStorage) Election(electionID types.HexBytes) (*Election, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	election, err := ms.getElection(electionID)
	if err != nil {
		return nil, err
	}
	return election, nil
}

func (ms *MongoStorage) getElection(electionID types.HexBytes) (*Election, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := ms.elections.FindOne(ctx, bson.M{"_id": electionID.String()})
	var election Election
	if err := result.Decode(&election); err != nil {
		log.Warn(err)
		return nil, ErrElectionUnknown
	}
	return &election, nil
}

// updateElection makes a conditional update on the election, updating only non-zero fields
func (ms *MongoStorage) updateElection(election *Election) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateDoc, err := dynamicUpdateDocument(election, nil)
	if err != nil {
		return fmt.Errorf("failed to create update document: %w", err)
	}
	log.Debugw("update election", "updateDoc", updateDoc)
	opts := options.Update().SetUpsert(true) // Ensures the document is created if it does not exist
	_, err = ms.elections.UpdateOne(ctx, bson.M{"_id": election.ElectionID}, updateDoc, opts)
	if err != nil {
		return fmt.Errorf("cannot update election: %w", err)
	}
	return nil
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

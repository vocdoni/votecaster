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
	usersCount, usersCountInitial, tokenDecimals uint32,
	endTime time.Time,
	community *ElectionCommunity,
) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	election := Election{
		UserID:                userFID,
		ElectionID:            electionID.String(),
		CreatedTime:           time.Now(),
		EndTime:               endTime,
		Source:                source,
		FarcasterUserCount:    usersCount,
		InitialAddressesCount: usersCountInitial,
		Question:              question,
		ElectionMeta: ElectionMeta{
			CensusERC20TokenDecimals: tokenDecimals,
		},
		Community: community,
	}
	log.Infow("added new election", "electionID", electionID.String(), "userID", userFID, "question", question)
	return ms.addElection(&election)
}

// ElectionsByUser returns all the elections created by the user with the FID
// provided, sorted by CreatedTime in descending order.
func (ms *MongoStorage) ElectionsByUser(userFID uint64, count int64) ([]ElectionRanking, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Specify the sorting order for the query
	opts := options.Find().SetSort(bson.D{{Key: "createdTime", Value: -1}}).SetLimit(count)

	cursor, err := ms.elections.Find(ctx, bson.M{"userId": userFID}, opts)
	if err != nil {
		log.Warn(err)
		return nil, ErrElectionUnknown
	}
	defer cursor.Close(ctx)

	var elections []ElectionRanking
	for cursor.Next(ctx) {
		var election Election
		if err := cursor.Decode(&election); err != nil {
			log.Warn(err)
			continue
		}

		user, err := ms.userData(election.UserID)
		if err != nil {
			log.Warn(err)
			continue
		}

		// Fall back to the election title if no question is stored in the database
		question := election.Question
		if question == "" {
			eid, err := hex.DecodeString(election.ElectionID)
			if err != nil {
				log.Warnf("invalid election ID: %v", err)
				continue
			}
			e, err := ms.election(eid)
			if err != nil {
				log.Warnf("failed to get election: %v", err)
				continue
			}
			if e == nil || e.Metadata == nil || e.Metadata.Title == nil {
				log.Warnw("missing election question, from vocdoni API", "electionID", election.ElectionID)
				continue
			}
			question = e.Metadata.Title["default"]
		}

		elections = append(elections, ElectionRanking{
			ElectionID:           election.ElectionID,
			Title:                question,
			VoteCount:            election.CastedVotes,
			CreatedByFID:         election.UserID,
			CreatedByUsername:    user.Username,
			CreatedByDisplayname: user.Displayname,
		})
	}
	return elections, nil
}

// ElectionsByCommunity returns all the elections created by the community with the ID.
func (ms *MongoStorage) ElectionsByCommunity(communityID uint64) ([]*Election, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Specify the sorting order for the query
	opts := options.Find().SetSort(bson.D{{Key: "createdTime", Value: -1}})

	cursor, err := ms.elections.Find(ctx, bson.M{"community.id": communityID}, opts)
	if err != nil {
		log.Warn(err)
		return nil, fmt.Errorf("failed to find elections by community ID: %w", err)
	}
	defer cursor.Close(ctx)

	var elections []*Election
	for cursor.Next(ctx) {
		var election Election
		if err := cursor.Decode(&election); err != nil {
			log.Warn("failed to decode election: ", err)
			continue
		}
		elections = append(elections, &election)
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

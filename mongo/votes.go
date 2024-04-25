package mongo

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/vocdoni/vote-frame/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

func (ms *MongoStorage) IncreaseVoteCount(userFID uint64, electionID types.HexBytes, weight *big.Int) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	log.Debugw("increase vote count", "userID", userFID, "electionID", electionID.String(), "weight", weight.String())

	user, err := ms.userData(userFID)
	if err != nil {
		return err
	}
	user.CastedVotes++
	if err := ms.updateUser(user); err != nil {
		return err
	}

	election, err := ms.getElection(electionID)
	if err != nil {
		if errors.Is(err, ErrElectionUnknown) {
			log.Warnw("creating fallback election", "electionID", electionID.String(), "userFID", userFID)
			election = &Election{
				UserID:       userFID,
				CastedVotes:  0,
				CastedWeight: new(big.Int).SetUint64(0).String(),
				ElectionID:   electionID.String(),
				CreatedTime:  time.Now(),
			}
			if err := ms.addElection(election); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	election.CastedVotes++
	election.LastVoteTime = time.Now()
	accCastedWeight, _ := new(big.Int).SetString(election.CastedWeight, 10)
	if accCastedWeight == nil {
		accCastedWeight = new(big.Int).SetUint64(0)
	}
	election.CastedWeight = new(big.Int).Add(accCastedWeight, helpers.TruncateDecimals(weight, election.CensusERC20TokenDecimals)).String()

	if err := ms.updateElection(election); err != nil {
		return err
	}

	return ms.addVoterToElection(election, userFID)
}

// VotersOfElection returns the list of voters of an election (usernames).
func (ms *MongoStorage) VotersOfElection(electionID types.HexBytes) ([]*User, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	voters, err := ms.votersOfElection(electionID)
	if err != nil {
		return nil, err
	}
	var users []*User
	for _, voter := range voters.Voters {
		u, err := ms.userData(voter)
		if err != nil {
			log.Warnw("failed to get user", "userID", voter, "err", err)
			continue
		}
		users = append(users, u)
	}
	return users, nil
}

func (ms *MongoStorage) votersOfElection(electionID types.HexBytes) (*VotersOfElection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := ms.voters.FindOne(ctx, bson.M{"_id": electionID.String()})
	var voters VotersOfElection
	if err := result.Decode(&voters); err != nil {
		log.Warn(err)
		return nil, ErrElectionUnknown
	}
	return &voters, nil
}

// addVoterToElection adds a voter to the list of voters of an election and updates the turnout.
func (ms *MongoStorage) addVoterToElection(election *Election, userFID uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Debugw("add voter to election", "userID", userFID, "electionID", election.ElectionID)
	eid, err := hex.DecodeString(election.ElectionID)
	if err != nil {
		return fmt.Errorf("invalid election ID: %w", err)
	}
	voters, err := ms.votersOfElection(eid)
	if err != nil {
		if errors.Is(err, ErrElectionUnknown) {
			return ms.createVotersList(ctx, eid, userFID)
		} else {
			return fmt.Errorf("error retrieving voters: %w", err)
		}
	}

	if ms.isUserVoter(voters, userFID) {
		return nil // Voter already in list, no update needed
	}

	return ms.updateVotersList(ctx, eid, voters, userFID)
}

func (ms *MongoStorage) createVotersList(ctx context.Context, electionID types.HexBytes, userFID uint64) error {
	voters := &VotersOfElection{
		ElectionID: electionID.String(),
		Voters:     []uint64{userFID},
	}
	_, err := ms.voters.InsertOne(ctx, voters)
	if err != nil {
		return fmt.Errorf("failed to insert new voters list: %w", err)
	}
	return nil
}

func (ms *MongoStorage) isUserVoter(voters *VotersOfElection, userFID uint64) bool {
	for _, v := range voters.Voters {
		if v == userFID {
			return true
		}
	}
	return false
}

func (ms *MongoStorage) updateVotersList(ctx context.Context, electionID types.HexBytes, voters *VotersOfElection, userFID uint64) error {
	voters.Voters = append(voters.Voters, userFID)
	_, err := ms.voters.ReplaceOne(ctx, bson.M{"_id": electionID.String()}, voters)
	if err != nil {
		return fmt.Errorf("failed to update voters list: %w", err)
	}
	return nil
}

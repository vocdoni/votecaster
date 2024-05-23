package mongo

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

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
	election.CastedWeight = new(big.Int).Add(accCastedWeight, weight).String()

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

// RemindersOfElection returns the list of remindable voters of an election and
// the number of already reminded voters.
func (ms *MongoStorage) RemindersOfElection(electionID types.HexBytes) (map[uint64]string, uint32, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	voters, err := ms.votersOfElection(electionID)
	if err != nil {
		return nil, 0, err
	}
	return voters.RemindableVoters, uint32(len(voters.AlreadyReminded)), nil
}

// RemindersSent updates the list of remindable voters and the list of already
// reminded voters of an election. It receives a map of user fids and usernames
// of the last reminders sent, and updates the lists accordingly, by removing
// the users from the remindable list and adding them to the already reminded
// list.
func (ms *MongoStorage) RemindersSent(electionID types.HexBytes, reminders map[uint64]string) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	voters, err := ms.votersOfElection(electionID)
	if err != nil {
		if errors.Is(err, ErrElectionUnknown) {
			return ms.createVotersList(ctx, electionID, 0)
		} else {
			return fmt.Errorf("error retrieving voters: %w", err)
		}
	}
	// get the current reminders from the database
	remindable := map[uint64]string{}
	for userFid, username := range voters.RemindableVoters {
		remindable[userFid] = username
	}
	alreadyReminded := map[uint64]string{}
	for userFid, username := range voters.AlreadyReminded {
		alreadyReminded[userFid] = username
	}
	// update the already reminded and remindable lists
	for userFid, username := range reminders {
		alreadyReminded[userFid] = username
		delete(remindable, userFid)
	}
	ctx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	_, err = ms.voters.UpdateOne(ctx, bson.M{"_id": electionID.String()}, bson.M{"$set": bson.M{
		"already_reminded":  alreadyReminded,
		"remindable_voters": remindable,
	}})
	if err != nil {
		return fmt.Errorf("failed to update reminders: %w", err)
	}
	return nil
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
	// create a new voters list with the user as the only voter if userFID is not 0
	users := []uint64{}
	if userFID != 0 {
		users = append(users, userFID)
	}
	voters := &VotersOfElection{
		ElectionID:       electionID.String(),
		Voters:           users,
		RemindableVoters: map[uint64]string{},
		AlreadyReminded:  map[uint64]string{},
	}
	if _, err := ms.voters.InsertOne(ctx, voters); err != nil {
		return fmt.Errorf("failed to insert new voters list: %w", err)
	}
	return nil
}

// populateRemindableVoters creates the list of remindable voters for an election.
// It gets the list of participants from the census and creates the list of
// remindable voters from the participants usernames and fids.
func (ms *MongoStorage) populateRemindableVoters(electionID types.HexBytes) error {
	log.Debugw("populating remindable voters", "electionID", electionID.String())
	// get the list of users that can be reminded (all participants)
	census, err := ms.censusFromElection(electionID)
	if err != nil {
		return fmt.Errorf("failed to get census: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// check if the election exists in the voters database, if not create it
	if _, err := ms.votersOfElection(electionID); err != nil {
		if err == ErrElectionUnknown {
			if err := ms.createVotersList(ctx, electionID, 0); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("error retrieving voters: %w", err)
		}
	}
	// create the remindable voters list getting the user fids from the
	// participants usernames
	remindableVoters := map[uint64]string{}
	for username := range census.Participants {
		user, err := ms.userDataByUsername(username)
		if err != nil {
			return fmt.Errorf("failed to get user by username: %w", err)
		}
		remindableVoters[user.UserID] = username
	}
	_, err = ms.voters.UpdateOne(ctx, bson.M{"_id": electionID.String()}, bson.M{"$set": bson.M{"remindable_voters": remindableVoters}})
	if err != nil {
		return fmt.Errorf("failed to update reminders: %w", err)
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
	// remove voter from remindable list
	delete(voters.RemindableVoters, userFID)
	if _, err := ms.voters.ReplaceOne(ctx, bson.M{"_id": electionID.String()}, voters); err != nil {
		return fmt.Errorf("failed to update voters list: %w", err)
	}
	return nil
}

package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

func (ms *MongoStorage) IncreaseVoteCount(userFID uint64, electionID types.HexBytes) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	log.Debugw("increase vote count", "userID", userFID, "electionID", electionID.String())

	user, err := ms.getUserData(userFID)
	if err != nil {
		return err
	}
	user.CastedVotes++

	election, err := ms.getElection(electionID)
	if err != nil {
		if errors.Is(err, ErrElectionUnknown) {
			log.Warnw("creating fallback election", "electionID", electionID.String(), "userFID", userFID)
			election = &Election{
				UserID:      userFID,
				CastedVotes: 0,
				ElectionID:  electionID.String(),
				CreatedTime: time.Now(),
				Source:      ElectionSourceUnknown,
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

	if err := ms.updateUser(user); err != nil {
		return err
	}
	if err := ms.addVoterToElection(electionID, userFID); err != nil {
		return err
	}
	return ms.updateElection(election)
}

func (ms *MongoStorage) addVoterToElection(electionID types.HexBytes, userFID uint64) error {
	log.Debugw("add voter to election", "userID", userFID, "electionID", electionID.String())
	voters, err := ms.getVotersOfElection(electionID)
	if err != nil {
		if errors.Is(err, ErrElectionUnknown) {
			voters = &VotersOfElection{
				ElectionID: electionID.String(),
				Voters:     []uint64{userFID},
			}
			_, err := ms.voters.InsertOne(context.Background(), voters)
			return err
		}
		return err
	}

	voters.Voters = append(voters.Voters, userFID)
	_, err = ms.voters.ReplaceOne(context.Background(), bson.M{"_id": electionID.String()}, voters)
	return err
}

// VotersOfElection returns the list of voters of an election (usernames).
func (ms *MongoStorage) VotersOfElection(electionID types.HexBytes) ([]string, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	voters, err := ms.getVotersOfElection(electionID)
	if err != nil {
		return nil, err
	}
	var usernames []string
	for _, voter := range voters.Voters {
		u, err := ms.getUserData(voter)
		if err != nil {
			log.Warnw("failed to get user", "userID", voter, "err", err)
			continue
		}
		usernames = append(usernames, u.Username)
	}
	return usernames, nil
}

func (ms *MongoStorage) getVotersOfElection(electionID types.HexBytes) (*VotersOfElection, error) {
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

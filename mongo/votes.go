package mongo

import (
	"errors"
	"time"

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
	return ms.updateElection(election)
}

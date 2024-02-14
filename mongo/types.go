package mongo

import (
	"fmt"
	"time"

	"go.vocdoni.io/dvote/types"
)

var (
	ErrUserUnknown     = fmt.Errorf("user unknown")
	ErrElectionUnknown = fmt.Errorf("electionID unknown")
)

// Users is the list of users.
type Users struct {
	Users []uint64 `json:"users"`
}

// User represents a farcaster user.
type User struct {
	UserID        uint64   `json:"userID,omitempty" bson:"_id"`
	ElectionCount uint64   `json:"electionCount" bson:"electionCount"`
	CastedVotes   uint64   `json:"castedVotes" bson:"castedVotes"`
	Username      string   `json:"username" bson:"username"`
	Addresses     []string `json:"addresses" bson:"addresses"`
}

// Election represents an election and its details owned by a user.
type Election struct {
	ElectionID   types.HexBytes `json:"electionId" bson:"_id"`
	UserID       uint64         `json:"userId" bson:"userid"`
	CastedVotes  uint64         `json:"castedVotes" bson:"votes"`
	LastVoteTime time.Time      `json:"lastVoteTime" bson:"lastvotetime"`
	CreatedTime  time.Time      `json:"createdTime" bson:"createdtime"`
}

// UserCollection is a dataset containing several users (used for dump and import).
type UserCollection struct {
	Users []User `json:"users" bson:"users"`
}

// ElectionCollection is a dataset containing several elections (used for dump and import).
type ElectionCollection struct {
	Elections []Election `json:"elections" bson:"elections"`
}

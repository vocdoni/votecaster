package mongo

import (
	"fmt"
	"time"
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
	ElectionID   string    `json:"electionId" bson:"_id"`
	UserID       uint64    `json:"userId" bson:"userId"`
	CastedVotes  uint64    `json:"castedVotes" bson:"castedVotes"`
	LastVoteTime time.Time `json:"lastVoteTime" bson:"lastVoteTime"`
	CreatedTime  time.Time `json:"createdTime" bson:"createdTime"`
}

// Results represents the final results of an election.
type Results struct {
	ElectionID string `json:"electionId" bson:"_id"`
	FinalPNG   []byte `json:"finalPNG" bson:"finalPNG"`
}

// VotersOfElection represents the list of voters of an election.
type VotersOfElection struct {
	ElectionID string   `json:"electionId" bson:"_id"`
	Voters     []uint64 `json:"voters" bson:"voters"`
}

// Collection is a dataset containing several users, elections and results (used for dump and import).
type Collection struct {
	UserCollection
	ElectionCollection
	ResultsCollection
}

// UserCollection is a dataset containing several users (used for dump and import).
type UserCollection struct {
	Users []User `json:"users" bson:"users"`
}

// ElectionCollection is a dataset containing several elections (used for dump and import).
type ElectionCollection struct {
	Elections []Election `json:"elections" bson:"elections"`
}

// ResultsCollection is a dataset containing several election results (used for dump and import).
type ResultsCollection struct {
	Results []Results `json:"results" bson:"results"`
}

// UserRanking is a user ranking entry.
type UserRanking struct {
	FID      uint64 `json:"fid" bson:"fid"`
	Username string `json:"username" bson:"username"`
	Count    uint64 `json:"count" bson:"count"`
}

// ElectionRanking is an election ranking entry.
type ElectionRanking struct {
	ElectionID        string `json:"electionId" bson:"_id"`
	VoteCount         uint64 `json:"voteCount" bson:"voteCount"`
	CreatedByFID      uint64 `json:"createdByFID" bson:"createdByFID"`
	CreatedByUsername string `json:"createdByUsername" bson:"createdByUsername"`
	Title             string `json:"title" bson:"title"`
}

package main

import "time"

const (
	maxElectionDuration = 24 * time.Hour * 15
)

// FarcasterProfile is the profile of a farcaster user.
type FarcasterProfile struct {
	Bio           string   `json:"bio"`
	Custody       string   `json:"custody"`
	DisplayName   string   `json:"displayName"`
	FID           uint64   `json:"fid"`
	Avatar        string   `json:"pfpUrl"`
	Username      string   `json:"username"`
	Verifications []string `json:"verifications"`
}

// ElectionCreateRequest is the request received by the farcaster auth, when creating an election.
type ElectionCreateRequest struct {
	ElectionDescription
	Profile     *FarcasterProfile `json:"profile,omitempty"`
	Census      *CensusInfo       `json:"census,omitempty"`
	NotifyUsers bool              `json:"notifyUsers"`
}

// ElectionDescription defines the parameters for a new election.
type ElectionDescription struct {
	Question          string        `json:"question"`
	Options           []string      `json:"options"`
	Duration          time.Duration `json:"duration"`
	Overwrite         bool          `json:"overwrite"`
	UsersCount        uint32        `json:"usersCount"`
	UsersCountInitial uint32        `json:"usersCountInitial"`
}

// CensusToken defines the parameters for a census token
type CensusToken struct {
	Address    string `json:"address"`
	Blockchain string `json:"blockchain"`
}

// CensusTokensRequest wraps a token census creation request
type CensusTokensRequest struct {
	Tokens []*CensusToken `json:"tokens"`
}

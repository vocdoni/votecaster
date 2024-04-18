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
	Profile          *FarcasterProfile `json:"profile,omitempty"`
	Census           *CensusInfo       `json:"census,omitempty"`
	NotifyUsers      bool              `json:"notifyUsers"`
	NotificationText string            `json:"notificationText"`
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

// Channel defines the attributes of a channel
type Channel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Followers   int    `json:"followerCount"`
	ImageURL    string `json:"image"`
	URL         string `json:"url"`
}

// User defines the attributes of a farcaster user in the farcaster.vote databse
type User struct {
	FID         uint64 `json:"fid"`
	Username    string `json:"username"`
	DisplayName string `json:"displayname"`
	Avatar      string `json:"pfpUrl"`
}

// ChannelList defines the list of channels
type ChannelList struct {
	Channels []*Channel `json:"channels"`
}

// CensusAddress defines the parameters of a address for a community census, is
// also used to check if a source of a future census is valid
type CensusAddress struct {
	Address    string `json:"address"`
	Blockchain string `json:"blockchain"`
}

// Community defines the attributes of a community, including the admins
// (FarcasterProfile), the census addresses (CensusAddress) and the channels
// (Channel)
type Community struct {
	ID              uint64           `json:"id"`
	Name            string           `json:"name"`
	LogoURL         string           `json:"logoURL"`
	Admins          []*User          `json:"admins"`
	Notifications   bool             `json:"notifications"`
	CensusName      string           `json:"censusName"`
	CensusType      string           `json:"censusType"`
	CensusAddresses []*CensusAddress `json:"censusAddresses,omitempty"`
	CensusChannel   *Channel         `json:"censusChannel,omitempty"`
	Channels        []string        `json:"channels"`
}

// CommunityList defines the list of communities
type CommunityList struct {
	Communities []*Community `json:"communities"`
}

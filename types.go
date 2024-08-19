package main

import (
	"time"

	"github.com/vocdoni/vote-frame/mongo"
)

const (
	maxElectionDuration = 24 * time.Hour * 15
	minPaginatedItems   = int64(1)
	maxPaginatedItems   = int64(100)
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

// WarpcastAPIKey is the user API key for the warpcast API service.
type WarpcastAPIKey struct {
	APIKey string `json:"apiKey"`
}

// ElectionCreateRequest is the request received by the farcaster auth, when creating an election.
type ElectionCreateRequest struct {
	ElectionDescription
	Profile          *FarcasterProfile `json:"profile,omitempty"`
	Census           *CensusInfo       `json:"census,omitempty"`
	NotifyUsers      bool              `json:"notifyUsers"`
	NotificationText string            `json:"notificationText"`
	CommunityID      *string           `json:"community,omitempty"`
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

// ElectionInfo defines the full details for an election, used by the API.
type ElectionInfo struct {
	CreatedTime             time.Time                `json:"createdTime"`
	ElectionID              string                   `json:"electionId"`
	LastVoteTime            time.Time                `json:"lastVoteTime"`
	EndTime                 time.Time                `json:"endTime"`
	Question                string                   `json:"question"`
	CastedVotes             uint64                   `json:"voteCount"`
	CastedWeight            string                   `json:"castedWeight,omitempty"`
	CensusParticipantsCount uint64                   `json:"censusParticipantsCount"`
	Turnout                 float32                  `json:"turnout"`
	FID                     uint64                   `json:"createdByFID,omitempty"`
	Username                string                   `json:"createdByUsername,omitempty"`
	Displayname             string                   `json:"createdByDisplayname,omitempty"`
	TotalWeight             string                   `json:"totalWeight,omitempty"`
	Participants            []uint64                 `json:"participants,omitempty"`
	Choices                 []string                 `json:"options,omitempty"`
	Votes                   []string                 `json:"tally,omitempty"`
	Finalized               bool                     `json:"finalized"`
	Community               *mongo.ElectionCommunity `json:"community,omitempty"`
}

// RankedElection defines the attributes of a ranked election
type RankedElection struct {
	CreatedTime             time.Time  `json:"createdTime"`
	ElectionID              string     `json:"electionId"`
	LastVoteTime            time.Time  `json:"lastVoteTime"`
	Question                string     `json:"title"`
	CastedVotes             uint64     `json:"voteCount"`
	CensusParticipantsCount uint64     `json:"censusParticipantsCount"`
	Username                string     `json:"createdByUsername"`
	Displayname             string     `json:"createdByDisplayname"`
	Community               *Community `json:"community,omitempty"`
}

// RankedElections defines the list of ranked elections and the pagination info
type RankedElections struct {
	Elections  []*RankedElection `json:"polls"`
	Pagination *Pagination       `json:"pagination,omitempty"`
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
	DisplayName string `json:"displayName"`
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
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	LogoURL         string           `json:"logoURL"`
	GroupChatURL    string           `json:"groupChat"`
	Admins          []*User          `json:"admins,omitempty"`
	Notifications   bool             `json:"notifications"`
	CensusType      string           `json:"censusType,omitempty"`
	CensusAddresses []*CensusAddress `json:"censusAddresses,omitempty"`
	CensusChannel   *Channel         `json:"censusChannel,omitempty"`
	UserRef         *User            `json:"userRef,omitempty"`
	Channels        []string         `json:"channels,omitempty"`
	Disabled        bool             `json:"disabled"`
}

// CommunityList defines the list of communities
type CommunityList struct {
	Communities []*Community `json:"communities"`
	Pagination  *Pagination  `json:"pagination,omitempty"`
}

// Pagination defines the pagination of a list
type Pagination struct {
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
	Total  int64 `json:"total"`
}

// ElectionVotersUsernames defines the usernames of the voters and the remaining
// users to vote in an election
type ElectionVotersUsernames struct {
	Usernames []string `json:"usernames"`
}

// DirectNotification defines the required parameters to send a notification
// via direct message
type DirectNotification struct {
	ElectionID string   `json:"electionId"`
	Content    string   `json:"content"`
	FIDs       []uint64 `json:"fids"`
}

// Rminders defines the data related to a election reminders, such as the
// list of remindable voters and the number of reminders that have been already
// sent
type Reminders struct {
	RemindableVoters       map[uint64]string `json:"remindableVoters"`
	RemindableVotersWeight map[uint64]string `json:"votersWeight"`
	AlreadySent            uint64            `json:"alreadySent"`
	MaxReminders           uint64            `json:"maxReminders"`
}

const (
	// IndividualRemindersType is the type of reminder that is sent to each user
	// individually
	IndividualRemindersType = "individual"
	// RankedRemindersType is the type of reminder that is sent to a ranked list
	// of users
	RankedRemindersType = "ranked"
)

// ReminderRequest defines the parameters to send a reminder, including the
// type of reminder, the content of the reminder, the users to send the reminder
// to (for individual reminders) and the number of users to send the reminder
// to (for ranked reminders)
type ReminderRequest struct {
	Type          string            `json:"type"`          // 'individual' or 'ranked'
	Content       string            `json:"content"`       // content of the reminder
	Users         map[uint64]string `json:"users"`         // map of userFID to username for 'individual' type
	NumberOfUsers int               `json:"numberOfUsers"` // number of users to send reminders to for 'ranked' type
}

// ReminderResponse defines the response of a reminder request, including the
// queue ID of the background process that will send the reminders. It allows to
// check the status of the process.
type ReminderResponse struct {
	QueueID string `json:"queueId"`
}

// RemindersStatus defines the status of a reminders process, including the
// number of reminders that have been already sent, the total number of
// reminders to send and the list of users that have failed to receive the
// reminder (with the error message)
type RemindersStatus struct {
	Completed   bool              `json:"completed"`
	ElectionID  string            `json:"electionId"`
	AlreadySent int               `json:"alreadySent"`
	Total       int               `json:"total"`
	Fails       map[string]string `json:"fails,omitempty"`
}

// AnnouncementRequest defines the parameters to send an announcement, including
// the content of the announcement and the users to send the announcement to.
type AnnouncementRequest struct {
	Content string            `json:"content"` // content of the reminder
	Users   map[uint64]string `json:"users"`   // map of userFID to username
}

// AnnouncementResponse defines the response of an announcement request,
// including the queue ID of the background process that will send the
// announcement. It allows to check the status of the process.
type AnnouncementResponse struct {
	QueuedID string `json:"queuedId"`
}

// AnnouncementStatus defines the status of an announcement process, including
// the number of announcements that have been already sent, the total number of
// announcements to send and the list of users that have failed to receive the
// announcement (with the error message). It also includes the error message in
// case of an global error and a flag to indicate if the process has been
// completed.
type AnnouncementStatus struct {
	CommunityID string            `json:"communityId"`
	Completed   bool              `json:"completed"`
	AlreadySent int               `json:"alreadySent"`
	Total       int               `json:"total"`
	Fails       map[string]string `json:"fails,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// ComposerActionResponse is the response of the composer endpoint, which is a
// redirection to the composer app to be used to create a new election from the
// cast form in warpcast.
type ComposerActionResponse struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// ComposerActionMetadata is the metadata of the composer action, which is used
// to show the action in the warpcast composer selector.
type ComposerActionMetadata struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	Action      struct {
		Type string `json:"type"`
	} `json:"action"`
}

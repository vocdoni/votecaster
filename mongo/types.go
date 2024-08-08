package mongo

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrUserUnknown     = fmt.Errorf("user unknown")
	ErrAvatarUnknown   = fmt.Errorf("avatar unknown")
	ErrElectionUnknown = fmt.Errorf("electionID unknown")
	ErrNoResults       = fmt.Errorf("no results found")
)

// Users is the list of users.
type Users struct {
	Users []uint64 `json:"users"`
}

// User represents a farcaster user.
type User struct {
	UserID         uint64    `json:"userID,omitempty" bson:"_id"`
	ElectionCount  uint64    `json:"electionCount" bson:"electionCount"`
	CastedVotes    uint64    `json:"castedVotes" bson:"castedVotes"`
	Username       string    `json:"username" bson:"username"`
	Displayname    string    `json:"displayName" bson:"displayname"`
	CustodyAddress string    `json:"custodyAddress" bson:"custodyAddress"`
	Addresses      []string  `json:"addresses" bson:"addresses"`
	Signers        []string  `json:"signers" bson:"signers"`
	Followers      uint64    `json:"followers" bson:"followers"`
	LastUpdated    time.Time `json:"lastUpdated" bson:"lastUpdated"`
	Avatar         string    `json:"avatar" bson:"avatar"`
}

// UserAccessProfile holds the user's access profile data, used by our backend to determine the user's access level.
// It also holds the notification status.
type UserAccessProfile struct {
	UserID                  uint64   `json:"userID,omitempty" bson:"_id"`
	NotificationsAccepted   bool     `json:"notificationsAccepted" bson:"notificationsAccepted"`
	NotificationsRequested  bool     `json:"notificationsRequested" bson:"notificationsRequested"`
	Reputation              uint32   `json:"reputation" bson:"reputation"`
	AccessLevel             uint32   `json:"accessLevel" bson:"accessLevel"`
	WhiteListed             bool     `json:"whiteListed" bson:"whiteListed"`
	NotificationsMutedUsers []uint64 `json:"notificationsMutedUsers" bson:"notificationsMutedUsers"`
	WarpcastAPIKey          string   `json:"warpcastAPIKey" bson:"warpcastAPIKey"`
}

type Reputation struct {
	// ids
	CommunityID     string `json:"communityID" bson:"communityID"`
	UserID          uint64 `json:"userID" bson:"userID"`
	TotalReputation uint64 `json:"totalReputation" bson:"totalReputation"`
	TotalPoints     uint64 `json:"totalPoints" bson:"totalPoints"`
	// community
	Participation float64 `json:"participation" bson:"participation"`
	CensusSize    uint64  `json:"censusSize" bson:"censusSize"`
	// activity
	FollowersCount        uint64 `json:"followersCount" bson:"followersCount"`
	ElectionsCreatedCount uint64 `json:"electionsCreatedCount" bson:"electionsCreatedCount"`
	CastVotesCount        uint64 `json:"castVotesCount" bson:"castVotesCount"`
	ParticipationsCount   uint64 `json:"participationsCount" bson:"participationsCount"`
	CommunitiesCount      uint64 `json:"communitiesCount" bson:"communitiesCount"`
	// boosters
	HasVotecasterNFTPass           bool `json:"hasVotecasterNFTPass" bson:"hasVotecasterNFTPass"`
	HasVotecasterLaunchNFT         bool `json:"hasVotecasterLaunchNFT" bson:"hasVotecasterLaunchNFT"`
	IsVotecasterAlphafrensFollower bool `json:"isVotecasterAlphafrensFollower" bson:"isVotecasterAlphafrensFollower"`
	IsVotecasterFarcasterFollower  bool `json:"isVotecasterFarcasterFollower" bson:"isVotecasterFarcasterFollower"`
	IsVocdoniFarcasterFollower     bool `json:"isVocdoniFarcasterFollower" bson:"isVocdoniFarcasterFollower"`
	VotecasterAnnouncementRecasted bool `json:"votecasterAnnouncementRecasted" bson:"votecasterAnnouncementRecasted"`
	HasKIWI                        bool `json:"hasKIWI" bson:"hasKIWI"`
	HasDegenDAONFT                 bool `json:"hasDegenDAONFT" bson:"hasDegenDAONFT"`
	HasHaberdasheryNFT             bool `json:"hasHaberdasheryNFT" bson:"hasHaberdasheryNFT"`
	Has10kDegenAtLeast             bool `json:"has10kDegenAtLeast" bson:"has10kDegenAtLeast"`
	HasTokyoDAONFT                 bool `json:"hasTokyoDAONFT" bson:"hasTokyoDAONFT"`
	HasProxy                       bool `json:"hasProxy" bson:"hasProxy"`
	Has5ProxyAtLeast               bool `json:"has5ProxyAtLeast" bson:"has5ProxyAtLeast"`
	HasProxyStudioNFT              bool `json:"hasProxyStudioNFT" bson:"hasProxyStudioNFT"`
	HasNameDegen                   bool `json:"hasNameDegen" bson:"hasNameDegen"`
	HasFarcasterOGNFT              bool `json:"hasFarcasterOGNFT" bson:"hasFarcasterOGNFT"`
	HasMoxiePass                   bool `json:"hasMoxiePass" bson:"hasMoxiePass"`
}

// ElectionCommunity represents the community used to create an election.
type ElectionCommunity struct {
	ID   string `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
}

// Election represents an election and its details owned by a user.
type Election struct {
	ElectionID            string             `json:"electionId" bson:"_id"`
	UserID                uint64             `json:"userId" bson:"userId"`
	CastedVotes           uint64             `json:"castedVotes" bson:"castedVotes"`
	LastVoteTime          time.Time          `json:"lastVoteTime" bson:"lastVoteTime"`
	CreatedTime           time.Time          `json:"createdTime" bson:"createdTime"`
	EndTime               time.Time          `json:"endTime" bson:"endTime"`
	Source                string             `json:"source" bson:"source"`
	FarcasterUserCount    uint32             `json:"farcasterUserCount" bson:"farcasterUserCount"`
	InitialAddressesCount uint32             `json:"initialAddressesCount" bson:"initialAddressesCount"`
	Question              string             `json:"question" bson:"question"`
	Community             *ElectionCommunity `json:"community" bson:"community"`
	CastedWeight          string             `json:"castedWeight" bson:"castedWeight"`
}

// Census stores the census of an election ready to be used for voting on farcaster.
type Census struct {
	CensusID           string            `json:"censusId" bson:"_id"`
	Root               string            `json:"root" bson:"root"`
	ElectionID         string            `json:"electionId" bson:"electionId"`
	Participants       map[string]string `json:"participants" bson:"participants"`
	FromTotalAddresses uint32            `json:"fromTotalAddresses" bson:"fromTotalAddresses"`
	CreatedBy          uint64            `json:"createdBy" bson:"createdBy"`
	TotalWeight        string            `json:"totalWeight" bson:"totalWeight"`
	URL                string            `json:"url" bson:"url"`
}

// ElectionMeta stores non related election information that is useful
// for certain types of frame interactions
type ElectionMeta struct {
	// CensusERC20TokenDecimals is the number of decimals that a certain ERC20 token, that was used
	// for creating the census of the election, has.
	CensusERC20TokenDecimals uint32 `json:"censusERC20TokenDecimals" bson:"censusERC20TokenDecimals"`
}

// Results represents the final results of an election.
type Results struct {
	ElectionID string   `json:"electionId" bson:"_id"`
	FinalPNG   []byte   `json:"finalPNG" bson:"finalPNG"`
	Choices    []string `json:"title" bson:"title"`
	Votes      []string `json:"votes" bson:"votes"`
	Finalized  bool     `json:"finalized" bson:"finalized"`
}

// VotersOfElection represents the list of voters of an election. It includes
// the list of voters, the list of users that have already been reminded and
// the list of users that can be reminded about the election.
type VotersOfElection struct {
	ElectionID       string            `json:"electionId" bson:"_id"`
	Voters           []uint64          `json:"voters" bson:"voters"`
	AlreadyReminded  map[uint64]string `json:"already_reminded" bson:"already_reminded"`
	RemindableVoters map[uint64]string `json:"remindable_voters" bson:"remindable_voters"`
}

// Authentication represents the authentication data for a user.
type Authentication struct {
	UserID     uint64    `json:"userId" bson:"_id"`
	AuthTokens []string  `json:"authTokens" bson:"authTokens"`
	UpdatedAt  time.Time `json:"updatedAt" bson:"updatedAt"`
}

// NotificationType represents the type of notification to be sent to a user.
type NotificationType int

const (
	NotificationTypeNewElection NotificationType = iota
	// create more notification types here
)

// Notification represents a notification to be sent to a user.
type Notification struct {
	ID             int64            `json:"id" bson:"_id"`
	Type           NotificationType `json:"type" bson:"type"`
	UserID         uint64           `json:"userId" bson:"userId"`
	Username       string           `json:"username" bson:"username"`
	AuthorID       uint64           `json:"authorId" bson:"authorId"`
	AuthorUsername string           `json:"authorUsername" bson:"authorUsername"`
	CommunityID    string           `json:"communityId" bson:"communityId"`
	CommunityName  string           `json:"communityName" bson:"communityName"`
	ElectionID     string           `json:"electionId" bson:"electionId"`
	FrameUrl       string           `json:"frameUrl" bson:"frameUrl"`
	CustomText     string           `json:"customText" bson:"customText"`
	Deadline       time.Time        `json:"deadline" bson:"deadline"`
}

// Collection is a dataset containing several users, elections and results (used for dump and import).
type Collection struct {
	UserCollection
	ElectionCollection
	ResultsCollection
	VotersOfElectionCollection
	CensusCollection
	CommunitiesCollection
	AvatarsCollection
	UserAccessProfileCollection
	DelegationsCollection
	ReputationCollection
}

// UserCollection is a dataset containing several users (used for dump and import).
type UserCollection struct {
	Users []User `json:"users" bson:"users"`
}

// ElectionCollection is a dataset containing several elections (used for dump and import).
type ElectionCollection struct {
	Elections []Election `json:"elections" bson:"elections"`
}

// CensusCollection is a dataset containing several censuses (used for dump and import).
type CensusCollection struct {
	Censuses []Census `json:"censuses" bson:"censuses"`
}

// ResultsCollection is a dataset containing several election results (used for dump and import).
type ResultsCollection struct {
	Results []Results `json:"results" bson:"results"`
}

// VotersOfElectionCollection is a dataset containing several voters of elections (used for dump and import).
type VotersOfElectionCollection struct {
	VotersOfElection []VotersOfElection `json:"votersOfElection" bson:"votersOfElection"`
}

// CommunitiesCollection is a dataset containing several communities (used for dump and import).
type CommunitiesCollection struct {
	Communities []Community `json:"communities" bson:"communities"`
}

// AvatarsCollection is a dataset containing several avatars from users and communities (used for dump and import).
type AvatarsCollection struct {
	Avatars []Avatar `json:"avatars" bson:"avatars"`
}

// UserAccessProfileCollection is a dataset containing several user access profiles (used for dump and import).
type UserAccessProfileCollection struct {
	UserAccessProfiles []UserAccessProfile `json:"userAccessProfiles" bson:"userAccessProfiles"`
}

// DelegationsCollection is a dataset containing several delegations (used for dump and import).
type DelegationsCollection struct {
	Delegations []Delegation `json:"delegations" bson:"delegations"`
}

// ReputationCollection is a dataset containing several reputations (used for dump and import).
type ReputationCollection struct {
	Reputations []Reputation `json:"reputations" bson:"reputations"`
}

// UserRanking is a user ranking entry.
type UserRanking struct {
	FID         uint64 `json:"fid" bson:"fid"`
	Username    string `json:"username" bson:"username"`
	Displayname string `json:"displayName" bson:"displayname"`
	Count       uint64 `json:"count" bson:"count"`
}

// ElectionRanking is an election ranking entry.
type ElectionRanking struct {
	ElectionID           string `json:"electionId" bson:"_id"`
	VoteCount            uint64 `json:"voteCount" bson:"voteCount"`
	CreatedByFID         uint64 `json:"createdByFID" bson:"createdByFID"`
	CreatedByUsername    string `json:"createdByUsername" bson:"createdByUsername"`
	CreatedByDisplayname string `json:"createdByDisplayname" bson:"createdByDisplayname"`
	Title                string `json:"title" bson:"title"`
}

type ReputationRanking struct {
	UserID           uint64 `json:"userID" bson:"userID"`
	Username         string `json:"username" bson:"username"`
	UserDisplayname  string `json:"userDisplayname" bson:"userDisplayname"`
	CommunityID      string `json:"communityID" bson:"communityID"`
	CommunityName    string `json:"communityName" bson:"communityName"`
	CommunityCreator uint64 `json:"communityCreator" bson:"communityCreator"`
	TotalPoints      uint64 `json:"totalPoints" bson:"totalPoints"`
	ImageURL         string `json:"imageURL" bson:"imageURL"`
}

// Community represents a community entry.
type Community struct {
	ID            string          `json:"id" bson:"_id"`
	Name          string          `json:"name" bson:"name"`
	Channels      []string        `json:"channels" bson:"channels"`
	Census        CommunityCensus `json:"census" bson:"census"`
	ImageURL      string          `json:"imageURL" bson:"imageURL"`
	GroupChatURL  string          `json:"groupChatURL" bson:"groupChatURL"`
	Creator       uint64          `json:"creator" bson:"creator"`
	Admins        []uint64        `json:"owners" bson:"owners"`
	Notifications bool            `json:"notifications" bson:"notifications"`
	Disabled      bool            `json:"disabled" bson:"disabled"`
	Featured      bool            `json:"featured" bson:"featured"`
}

const (
	// TypeCommunityCensusChannel is the type for a community census that uses
	// a channel as source.
	TypeCommunityCensusChannel = "channel"
	// TypeCommunityCensusERC20 is the type for a community census that uses
	// ERC20 holders as source.
	TypeCommunityCensusERC20 = "erc20"
	// TypeCommunityCensusNFT is the type for a community census that uses
	// NFT holders as source.
	TypeCommunityCensusNFT = "nft"
	// TypeCommunityCensusFollowers is the type for a community census that uses
	// followers as source.
	TypeCommunityCensusFollowers = "followers"
)

// CommunityCensus represents the census of a community in the database. It
// includes the name, type, and the census addresses (CommunityCensusAddresses)
// or the census channel (depending on the type).
type CommunityCensus struct {
	Type      string                     `json:"type" bson:"type"`
	Addresses []CommunityCensusAddresses `json:"addresses" bson:"addresses"`
	Channel   string                     `json:"channel" bson:"channel"`
}

// CommunityCensusAddresses represents the addresses of a contract to be used to
// create the census of a community.
type CommunityCensusAddresses struct {
	Address    string `json:"address" bson:"address"`
	Blockchain string `json:"blockchain" bson:"blockchain"`
}

// Avatar represents an avatar image. Includes the avatar ID and the image data
// as a byte array.
type Avatar struct {
	ID          string    `json:"id" bson:"_id"`
	Data        []byte    `json:"data" bson:"data"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UserID      uint64    `json:"userId" bson:"userId"`
	CommunityID string    `json:"communityId" bson:"communityId"`
	ContentType string    `json:"contentType" bson:"contentType"`
}

// Delegation represents a delegation of votes from one user to another for a
// specific community.
type Delegation struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	From       uint64             `json:"from" bson:"from"`
	To         uint64             `json:"to" bson:"to"`
	CommuniyID string             `json:"communityId" bson:"communityId"`
}

// dynamicUpdateDocument creates a BSON update document from a struct, including only non-zero fields.
// It uses reflection to iterate over the struct fields and create the update document.
// The struct fields must have a bson tag to be included in the update document.
// The _id field is skipped.
func dynamicUpdateDocument(item interface{}, alwaysUpdateTags []string) (bson.M, error) {
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if !val.IsValid() || val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a valid struct")
	}

	update := bson.M{}
	typ := val.Type()

	// Create a map for quick lookup
	alwaysUpdateMap := make(map[string]bool, len(alwaysUpdateTags))
	for _, tag := range alwaysUpdateTags {
		alwaysUpdateMap[tag] = true
	}
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanInterface() {
			continue
		}
		fieldType := typ.Field(i)
		tag := fieldType.Tag.Get("bson")
		if tag == "" || tag == "-" || tag == "_id" {
			continue
		}
		// Check if the field should always be updated or is not the zero value
		_, alwaysUpdate := alwaysUpdateMap[tag]
		if alwaysUpdate || !reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			update[tag] = field.Interface()
		}
	}

	return bson.M{"$set": update}, nil
}

// paginatedObjects method returns the paginated list of objects from the given
// collection by the provided query. It returns the list of resulting objects,
// the total number of results, and an error if something goes wrong. It
// receives the query to filter the collections objects, the limit of results
// to return, and the offset (the number of objects to skip).
func paginatedObjects(collection *mongo.Collection, query bson.M, opts *options.FindOptions, limit, offset int64, results any) (int64, error) {
	// count total communities by query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	total, err := collection.CountDocuments(ctx, query)
	if err != nil {
		return 0, err
	}
	// get communities with pagination
	if opts == nil {
		opts = options.Find()
	}
	// limit the number of results if the limit is greater than -1
	if limit > -1 {
		opts = opts.SetLimit(limit)
	}
	// skip the number of results if the offset is greater than 0
	if offset == 0 {
		opts = opts.SetSkip(offset)
	}
	ctx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		if strings.Contains(err.Error(), "no documents in result") {
			return total, nil
		}
		return total, err
	}
	ctx, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel3()
	if err := cursor.All(ctx, results); err != nil {
		return total, err
	}
	return total, nil
}

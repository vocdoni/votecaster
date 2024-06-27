package reputation

import (
	"fmt"

	"github.com/vocdoni/vote-frame/mongo"
)

// Error definitions to be handled by the caller
var ErrNoReputationInfo = fmt.Errorf("no reputation information found about the user")

// ActivityReputation struct contains the reputation of a user regarding their
// activity, including the number of followers, the number of elections created,
// the number of casted votes, the number of votes casted on created elections,
// and the number of communities the user is part of as admin.
type ActivityReputation struct {
	FollowersCount                uint64 `json:"followersCount"`
	ElectionsCreated              uint64 `json:"electionsCreated"`
	CastedVotes                   uint64 `json:"castedVotes"`
	VotesCastedOnCreatedElections uint64 `json:"participationAchievement"`
	CommunitiesCount              uint64 `json:"communitiesCount"`
}

// Boosters struct contains the boosters of a user, including if the user has
// the Votecaster NFT pass, the Votecaster Launch NFT, the user is subscribed
// to Votecaster Alphafrens channel, the user follows Votecaster and Vocdoni
// profiles on Farcaster, the user has recasted the Votecaster Launch cast
// announcement, the user has KIWI, the user has the DegenDAO NFT, the user has
// at least 10k Degen, the user has the TokyoDAO NFT, the user has a Proxy, the
// user has at least 5 Proxies, and the user has the NameDegen NFT.
type Boosters struct {
	HasVotecasterNFTPass           bool `json:"hasVotecasterNFTPass"`
	HasVotecasterLaunchNFT         bool `json:"hasVotecasterLaunchNFT"`
	IsVotecasterAlphafrensFollower bool `json:"isVotecasterAlphafrensFollower"`
	IsVotecasterFarcasterFollower  bool `json:"isVotecasterFarcasterFollower"`
	IsVocdoniFarcasterFollower     bool `json:"isVocdoniFarcasterFollower"`
	VotecasterAnnouncementRecasted bool `json:"votecasterAnnouncementRecasted"`
	HasKIWI                        bool `json:"hasKIWI"`
	HasDegenDAONFT                 bool `json:"hasDegenDAONFT"`
	HasHaberdasheryNFT             bool `json:"hasHaberdasheryNFT"`
	Has10kDegenAtLeast             bool `json:"has10kDegenAtLeast"`
	HasTokyoDAONFT                 bool `json:"hasTokyoDAONFT"`
	HasProxy                       bool `json:"hasProxy"`
	Has5ProxyAtLeast               bool `json:"has5ProxyAtLeast"`
	HasProxyStudioNFT              bool `json:"hasProxyStudioNFT"`
	HasNameDegen                   bool `json:"hasNameDegen"`
}

// Reputation struct contains the reputation of a user, detailed by activity and
// boosters
type Reputation struct {
	*ActivityReputation `json:"activity"`
	*Boosters           `json:"boosters"`
	TotalReputation     uint32 `json:"totalReputation"`
}

// Calculator struct contains the database connection to calculate the
// reputation of a user, it uses the detailed reputation of a user from the
// database to calculate the reputation.
type Calculator struct {
	db *mongo.MongoStorage
}

func NewCalculator(db *mongo.MongoStorage) *Calculator {
	return &Calculator{db: db}
}

// UserReputation returns the reputation of a user based the user ID. It gets
// the detailed reputation information of the user from the database and
// calculates the resulting reputation value.
func (c *Calculator) UserReputation(userID uint64) (*Reputation, error) {
	dbRep, err := c.db.DetailedUserReputation(userID)
	if err != nil {
		return &Reputation{}, fmt.Errorf("%w: %w", ErrNoReputationInfo, err)
	}
	activityRep := &ActivityReputation{
		FollowersCount:                dbRep.FollowersCount,
		ElectionsCreated:              dbRep.ElectionsCreated,
		CastedVotes:                   dbRep.CastedVotes,
		VotesCastedOnCreatedElections: dbRep.VotesCastedOnCreatedElections,
		CommunitiesCount:              dbRep.CommunitiesCount,
	}
	boosters := &Boosters{
		HasVotecasterNFTPass:           dbRep.HasVotecasterNFTPass,
		HasVotecasterLaunchNFT:         dbRep.HasVotecasterLaunchNFT,
		IsVotecasterAlphafrensFollower: dbRep.IsVotecasterAlphafrensFollower,
		IsVotecasterFarcasterFollower:  dbRep.IsVotecasterFarcasterFollower,
		IsVocdoniFarcasterFollower:     dbRep.IsVocdoniFarcasterFollower,
		VotecasterAnnouncementRecasted: dbRep.VotecasterAnnouncementRecasted,
		HasKIWI:                        dbRep.HasKIWI,
		HasDegenDAONFT:                 dbRep.HasDegenDAONFT,
		HasHaberdasheryNFT:             dbRep.HasHaberdasheryNFT,
		Has10kDegenAtLeast:             dbRep.Has10kDegenAtLeast,
		HasTokyoDAONFT:                 dbRep.HasTokyoDAONFT,
		HasProxy:                       dbRep.HasProxy,
		Has5ProxyAtLeast:               dbRep.Has5ProxyAtLeast,
		HasProxyStudioNFT:              dbRep.HasProxyStudioNFT,
		HasNameDegen:                   dbRep.HasNameDegen,
	}
	return &Reputation{
		ActivityReputation: activityRep,
		Boosters:           boosters,
		TotalReputation:    totalReputation(activityRep, boosters),
	}, nil
}

package reputation

import "github.com/vocdoni/vote-frame/mongo"

type ActivityReputationCounts struct {
	FollowersCount        uint64 `json:"followersCount"`
	ElectionsCreatedCount uint64 `json:"createdElectionsCount"`
	CastVotesCount        uint64 `json:"castVotesCount"`
	ParticipationsCount   uint64 `json:"participationsCount"`
	CommunitiesCount      uint64 `json:"communitiesCount"`
}

type ActivityReputationPoints struct {
	FollowersPoints        uint64 `json:"followersPoints"`
	ElectionsCreatedPoints uint64 `json:"createdElectionsPoints"`
	CastVotesPoints        uint64 `json:"castVotesPoints"`
	ParticipationsPoints   uint64 `json:"participationsPoints"`
	CommunitiesPoints      uint64 `json:"communitiesPoints"`
}

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
	Has5ProxyAtLeast               bool `json:"has5ProxyAtLeast"`
	HasProxyStudioNFT              bool `json:"hasProxyStudioNFT"`
	HasNameDegen                   bool `json:"hasNameDegen"`
	HasFarcasterOGNFT              bool `json:"hasFarcasterOGNFT"`
}

type ReputationInfo map[string]uint64

type Reputation struct {
	*Boosters       `json:"boosters"`
	BoostersInfo    ReputationInfo            `json:"boostersInfo"`
	ActivityPoints  *ActivityReputationPoints `json:"activityPoints"`
	ActivityCounts  *ActivityReputationCounts `json:"activityCounts"`
	ActivityInfo    ReputationInfo            `json:"activityInfo"`
	TotalReputation uint64                    `json:"totalReputation"`
	TotalPoints     uint64                    `json:"totalPoints"`
}

func ReputationToAPIResponse(rep *mongo.Reputation) *Reputation {
	activityPoints := &ActivityReputationCounts{
		FollowersCount:        rep.FollowersCount,
		ElectionsCreatedCount: rep.ElectionsCreatedCount,
		CastVotesCount:        rep.CastVotesCount,
		ParticipationsCount:   rep.ParticipationsCount,
		CommunitiesCount:      rep.CommunitiesCount,
	}
	return &Reputation{
		ActivityCounts: activityPoints,
		ActivityPoints: ponderateActivityReputation(activityPoints),
		ActivityInfo:   ActivityPuntuationInfo,
		Boosters: &Boosters{
			HasVotecasterNFTPass:           rep.HasVotecasterNFTPass,
			HasVotecasterLaunchNFT:         rep.HasVotecasterLaunchNFT,
			IsVotecasterAlphafrensFollower: rep.IsVotecasterAlphafrensFollower,
			IsVotecasterFarcasterFollower:  rep.IsVotecasterFarcasterFollower,
			IsVocdoniFarcasterFollower:     rep.IsVocdoniFarcasterFollower,
			VotecasterAnnouncementRecasted: rep.VotecasterAnnouncementRecasted,
			HasKIWI:                        rep.HasKIWI,
			HasDegenDAONFT:                 rep.HasDegenDAONFT,
			HasHaberdasheryNFT:             rep.HasHaberdasheryNFT,
			Has10kDegenAtLeast:             rep.Has10kDegenAtLeast,
			HasTokyoDAONFT:                 rep.HasTokyoDAONFT,
			Has5ProxyAtLeast:               rep.Has5ProxyAtLeast,
			HasProxyStudioNFT:              rep.HasProxyStudioNFT,
			HasNameDegen:                   rep.HasNameDegen,
			HasFarcasterOGNFT:              rep.HasFarcasterOGNFT,
		},
		BoostersInfo:    BoostersPuntuationInfo,
		TotalReputation: rep.TotalReputation,
		TotalPoints:     rep.TotalPoints,
	}
}

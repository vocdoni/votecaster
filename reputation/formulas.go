package reputation

import (
	"math"

	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/log"
)

// totalReputation calculates the reputation of a user based on their activity
// and boosters and returns the mean value of both.
func totalReputation(ar *ActivityReputationCounts, b *Boosters) uint64 {
	return (activityReputation(ar) + boostersReputation(b)) / 2
}

// activityReputation calculates the reputation of a user based on their
// ponderated activity reputation values. If the reputation exceeds 100, it is
// capped at 100.
func activityReputation(rep *ActivityReputationCounts) uint64 {
	ponderated := ponderateActivityReputation(rep)
	reputation := ponderated.FollowersPoints +
		ponderated.ElectionsCreatedPoints +
		ponderated.CastVotesPoints +
		ponderated.ParticipationsPoints +
		ponderated.CommunitiesPoints
	if reputation > maxReputation {
		reputation = maxReputation
	}
	return reputation
}

// ponderateActivityReputation calculates the ponderated reputation values of a
// user activity. The reputation is calculated based on the following
// activities:
//   - FollowersCount: number of followers/2000 points (up to
//     'maxFollowersReputation' points)
//   - ElectionsCreated: number of elections/10 points (up to
//     'maxElectionsReputation' points)
//   - CastedVotes: number of casted votes/4 points (up to
//     'maxVotesReputation' points)
//   - VotesCastedOnCreatedElections: number of votes casted on created
//     elections/20 points (up to 'maxCastedReputation' points)
//   - CommunitiesCount: number of communities*2 points (up to
//     'maxCommunityReputation' points)
func ponderateActivityReputation(ar *ActivityReputationCounts) *ActivityReputationPoints {
	p := &ActivityReputationPoints{}
	if p.FollowersPoints = ar.FollowersCount / followersDividerPonderation; p.FollowersPoints > maxFollowersReputation {
		p.FollowersPoints = maxFollowersReputation
	}
	if p.ElectionsCreatedPoints = ar.ElectionsCreatedCount / electionsDividerPonderation; p.ElectionsCreatedPoints > maxElectionsReputation {
		p.ElectionsCreatedPoints = maxElectionsReputation
	}
	if p.CastVotesPoints = ar.CastVotesCount / votesDividerPonderation; p.CastVotesPoints > maxVotesReputation {
		p.CastVotesPoints = maxVotesReputation
	}
	if p.ParticipationsPoints = ar.ParticipationsCount / castedDividerPonderation; p.ParticipationsPoints > maxCastedReputation {
		p.ParticipationsPoints = maxCastedReputation
	}
	if p.CommunitiesPoints = ar.CommunitiesCount * communitiesMultiplierPonderation; p.CommunitiesPoints > maxCommunityReputation {
		p.CommunitiesPoints = maxCommunityReputation
	}
	return p
}

// boostersReputation calculates the reputation of a user based on their boosters,
// it returns a value between 0 and 100. The reputation is calculated based on
// the following boosters:
//   - VotecasterNFTPass: 'votecasterNFTPassPuntuaction' points
//   - VotecasterLaunchNFT: 'votecasterLaunchNFTPuntuaction' points
//   - VotecasterAlphafrensFollower: 'votecasterAlphafrensFollowerPuntuaction' points
//   - VotecasterFarcasterFollower: 'votecasterFarcasterFollowerPuntuaction' points
//   - VocdoniFarcasterFollower: 'vocdoniFarcasterFollowerPuntuaction' points
//   - VotecasterAnnouncementRecasted: 'votecasterAnnouncementRecastedPuntuaction' points
//   - KIWI: 'kiwiPuntuaction' points
//   - DegenDAO NFT: 'degenDAONFTPuntuaction' points
//   - Haberdashery NFT: 'haberdasheryFTPuntuaction' points
//   - >=10k Degen: 'degenAtLeast10kPuntuaction' points
//   - TokyoDAO NFT: 'tokyoDAONFTPuntuaction' points
//   - ProxyStudio NFT: 'proxyStudioNFTPuntuaction' points
//   - >=5 Proxy: 'proxyAtLeast5Puntuaction' points
//   - NameDegen: 'nameDegenPuntuaction' points
//   - FarcasterOG NFT: 'farcasterOGNFTPuntuaction' points
//
// If the reputation exceeds 100, it is capped at 100.
func boostersReputation(rep *Boosters) uint64 {
	reputation := uint64(0)
	// add votecaster nft pass puntuaction if user has it
	if rep.HasVotecasterNFTPass {
		reputation += votecasterNFTPassPuntuaction
	}
	// add votecaster launch nft puntuaction if user has it
	if rep.HasVotecasterLaunchNFT {
		reputation += votecasterLaunchNFTPuntuaction
	}
	// add votecaster alphafrens follower puntuaction if user is subscribed
	// to votecaster alphafrens
	if rep.IsVotecasterAlphafrensFollower {
		reputation += votecasterAlphafrensFollowerPuntuaction
	}
	// add votecaster farcaster follower puntuaction if user follows votecaster
	// farcaster profile
	if rep.IsVotecasterFarcasterFollower {
		reputation += votecasterFarcasterFollowerPuntuaction
	}
	// add vocdoni farcaster follower puntuaction if user follows vocdoni
	// farcaster profile
	if rep.IsVocdoniFarcasterFollower {
		reputation += vocdoniFarcasterFollowerPuntuaction
	}
	// add votecaster announcement recasted puntuaction if user recasted the
	// votecaster launch announcement cast
	if rep.VotecasterAnnouncementRecasted {
		reputation += votecasterAnnouncementRecastedPuntuaction
	}
	// add kiwi puntuaction if user has kiwi
	// (oeth:0x66747bdC903d17C586fA09eE5D6b54CC85bBEA45)
	if rep.HasKIWI {
		reputation += kiwiPuntuaction
	}
	// add degen dao nft puntuaction if user has degen dao nft
	// (base:0x980Fbdd1cF05080781Dca0AEf7026B0406743389)
	if rep.HasDegenDAONFT {
		reputation += degenDAONFTPuntuaction
	}
	// add haberdashery nft puntuaction if user has haberdashery nft
	// (base:0x85E7DF5708902bE39891d59aBEf8E21EDE91E8BF)
	if rep.HasHaberdasheryNFT {
		reputation += haberdasheryNFTPuntuaction
	}
	// add degen at least 10k puntuaction if user has at least 10k degen
	// (base:0x4ed4E862860beD51a9570b96d89aF5E1B0Efefed)
	if rep.Has10kDegenAtLeast {
		reputation += degenAtLeast10kPuntuaction
	}
	// add tokyo dao nft puntuaction if user has tokyo dao nft
	// (base:0x432073397Aead241cf2411e21D8fA949183E7151)
	if rep.HasTokyoDAONFT {
		reputation += tokyoDAONFTPuntuaction
	}
	// add 5 proxy at least puntuaction if user has at least 5 proxies
	// (degen:0xA051A2Cb19C00eCDffaE94D0Ff98c17758041D16)
	if rep.Has5ProxyAtLeast {
		reputation += proxyAtLeast5Puntuaction
	}
	// add name degen puntuaction if user has name degen
	// (degen:0x4087fb91A1fBdef05761C02714335D232a2Bf3a1)
	if rep.HasNameDegen {
		reputation += nameDegenPuntuaction
	}
	// add farcaster og nft puntuaction if user has farcaster og nft
	// (base:0xe03ef4b9db1a47464de84fb476f9baf493b3e886)
	if rep.HasFarcasterOGNFT {
		reputation += farcasterOGNFTPuntuaction
	}
	// add moxie pass puntuaction if user has moxie pass
	// (base:0x235CAD50d8a510Bc9081279996f01877827142D8)
	if rep.HasMoxiePass {
		reputation += moxiePassPuntuaction
	}
	// ensure the reputation does not exceed 100
	if reputation > maxReputation {
		reputation = maxReputation
	}
	return reputation
}

// Y = (A * participationRate + B * log(censusSize)) * ownerRep
// if DAO => Y = Y * daoMultiplier
// if Channel => Y = Y * channelMultiplier
func communityYieldRate(p, cs, r float64, dao, channel bool) float64 {
	if cs <= 0 {
		return 0
	}
	y := (yieldParamA*p + yieldParamB*math.Log(cs)) * r
	if dao {
		y *= daoMultiplier
	} else if channel {
		y *= channelMultiplier
	}
	return y
}

// communityTotalPoints calculates the total points of a community based on the
// census type, the participation rate, the census size and the reputation of
// the owner. The total points are calculated as the yield rate multiplied by
// the participation rate and the census size.
func communityTotalPoints(censusType string, m, p float64, cs, r uint64) uint64 {
	var y float64
	switch censusType {
	case mongo.TypeCommunityCensusERC20, mongo.TypeCommunityCensusNFT:
		y = communityYieldRate(p, float64(cs), float64(r), true, false)
	case mongo.TypeCommunityCensusChannel:
		y = communityYieldRate(p, float64(cs), float64(r), false, true)
	case mongo.TypeCommunityCensusFollowers:
		y = communityYieldRate(p, float64(cs), float64(r), false, false)
	default:
		return 0
	}
	log.Debugw("community total points", "y", y, "m", m)
	return uint64(y * m)
}

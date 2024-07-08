package reputation

import (
	"math"

	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/log"
)

// totalReputation calculates the reputation of a user based on their activity
// and boosters and returns the mean value of both.
func totalReputation(ar *ActivityReputation, b *Boosters) uint32 {
	log.Infow("user reputation", "activity", activityReputation(ar), "boosters", boostersReputation(b))
	return (activityReputation(ar) + boostersReputation(b)) / 2
}

// activityReputation calculates the reputation of a user based on their
// ponderated activity reputation values. If the reputation exceeds 100, it is
// capped at 100.
func activityReputation(rep *ActivityReputation) uint32 {
	ponderated := ponderateActivityReputation(rep)
	reputation := uint32(ponderated.FollowersCount +
		ponderated.ElectionsCreated +
		ponderated.CastedVotes +
		ponderated.VotesCastedOnCreatedElections +
		ponderated.CommunitiesCount)
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
func ponderateActivityReputation(ar *ActivityReputation) *ActivityReputation {
	p := &ActivityReputation{}
	if p.FollowersCount = ar.FollowersCount / followersDividerPonderation; p.FollowersCount > maxFollowersReputation {
		p.FollowersCount = maxFollowersReputation
	}
	if p.ElectionsCreated = ar.ElectionsCreated / electionsDividerPonderation; p.ElectionsCreated > maxElectionsReputation {
		p.ElectionsCreated = maxElectionsReputation
	}
	if p.CastedVotes = ar.CastedVotes / votesDividerPonderation; p.CastedVotes > maxVotesReputation {
		p.CastedVotes = maxVotesReputation
	}
	if p.VotesCastedOnCreatedElections = ar.VotesCastedOnCreatedElections / castedDividerPonderation; p.VotesCastedOnCreatedElections > maxCastedReputation {
		p.VotesCastedOnCreatedElections = maxCastedReputation
	}
	if p.CommunitiesCount = ar.CommunitiesCount * communitiesMultiplierPonderation; p.CommunitiesCount > maxCommunityReputation {
		p.CommunitiesCount = maxCommunityReputation
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
func boostersReputation(rep *Boosters) uint32 {
	reputation := uint32(0)
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
	// ensure the reputation does not exceed 100
	if reputation > maxReputation {
		reputation = maxReputation
	}
	return reputation
}

// Y = (A * participationRate + B * log(censusSize)) * ownerRep
// if DAO => Y = Y * daoMultiplier
// if Channel => Y = Y * channelMultiplier
func communityYieldRate(participationRate, censusSize, ownerRep float64, dao, channel bool) float64 {
	y := (yieldParamA*participationRate + yieldParamB*math.Log(censusSize)) * ownerRep
	if dao {
		y *= daoMultiplier
	} else if channel {
		y *= channelMultiplier
	}
	return y
}

// sumOfYields calculates the sum of the yields of the communities and the
// reputation and multiplier provided. It returns the sum of the yields for
// the reputation multiplied by the multiplier. It calculates the yield rate
// based on the participation rate, census size, reputation, and the type of
// census (DAO (nft or erc20), Channel, or other (followers)).
func sumOfYields(communities []mongo.Community, reputation uint32, multiplier float64) uint64 {
	var points uint64
	for _, community := range communities {
		var yieldRate float64
		switch community.Census.Type {
		case mongo.TypeCommunityCensusERC20, mongo.TypeCommunityCensusNFT:
			yieldRate = communityYieldRate(community.Participation, float64(community.CensusSize), float64(reputation), true, false)
		case mongo.TypeCommunityCensusChannel:
			yieldRate = communityYieldRate(community.Participation, float64(community.CensusSize), float64(reputation), false, true)
		default:
			yieldRate = communityYieldRate(community.Participation, float64(community.CensusSize), float64(reputation), false, false)
		}
		points += uint64(yieldRate * multiplier)
	}
	return points
}

package reputation

import "go.vocdoni.io/dvote/log"

const (
	maxFollowersReputation = 10
	maxElectionsReputation = 10
	maxVotesReputation     = 25
	maxCastedReputation    = 45
	maxCommunityReputation = 10
	maxReputation          = 100
)

const (
	votecasterNFTPassPuntuaction              = 25
	votecasterLaunchNFTPuntuaction            = 6
	votecasterAlphafrensFollowerPuntuaction   = 20
	votecasterFarcasterFollowerPuntuaction    = 3
	vocdoniFarcasterFollowerPuntuaction       = 3
	votecasterAnnouncementRecastedPuntuaction = 5
	kiwiPuntuaction                           = 3
	degenDAONFTPuntuaction                    = 7
	haberdasheryFTPuntuaction                 = 8
	degenAtLeast10kPuntuaction                = 5
	tokyoDAONFTPuntuaction                    = 5
	proxyPuntuaction                          = 5
	proxyStudioNFTPuntuaction                 = 5
	proxyAtLeast5Puntuaction                  = 5
	nameDegenPuntuaction                      = 5
)

// totalReputation calculates the reputation of a user based on their activity
// and boosters and returns the mean value of both.
func totalReputation(ar *ActivityReputation, b *Boosters) uint32 {
	log.Infow("user reputation", "activity", activityReputation(ar), "boosters", boostersReputation(b))
	return (activityReputation(ar) + boostersReputation(b)) / 2
}

// activityReputation calculates the reputation of a user based on their
// activity, it returns a value between 0 and 100. The reputation is calculated
// based on the following activities:
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
//
// If the reputation exceeds 100, it is capped at 100.
func activityReputation(rep *ActivityReputation) uint32 {
	reputation := 0.0
	// Calculate FollowersCount score (up to 10 points, max 20000 followers)
	if followersRep := float64(rep.FollowersCount) / 2000; followersRep <= maxFollowersReputation {
		reputation += followersRep
	} else {
		reputation += maxFollowersReputation
	}
	// Calculate ElectionsCreated score (up to 10 points, max 100 elections)
	if electionsRep := float64(rep.ElectionsCreated) / 10; electionsRep <= maxElectionsReputation {
		reputation += electionsRep
	} else {
		reputation += maxElectionsReputation
	}
	// Calculate CastedVotes score (up to 30 points, max 120 votes)
	if votesRep := float64(rep.CastedVotes) / 4; votesRep <= maxVotesReputation {
		reputation += votesRep
	} else {
		reputation += maxVotesReputation
	}
	// Calculate VotesCastedOnCreatedElections score (up to 50 points, max 1000 votes)
	if castedRep := float64(rep.VotesCastedOnCreatedElections) / 20; castedRep <= maxCastedReputation {
		reputation += castedRep
	} else {
		reputation += maxCastedReputation
	}
	// Calculate CommunitiesCount score (up to 10 points, max 5 communities)
	if comRep := float64(rep.CommunitiesCount) * 2; comRep <= maxCommunityReputation {
		reputation += comRep
	} else {
		reputation += maxCommunityReputation
	}
	// Ensure the reputation does not exceed 100
	if reputation > maxReputation {
		reputation = maxReputation
	}
	return uint32(reputation)
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
//   - Proxy: 'proxyPuntuaction' points
//   - ProxyStudio NFT: 'proxyStudioNFTPuntuaction' points
//   - >=5 Proxy: 'proxyAtLeast5Puntuaction' points
//   - NameDegen: 'nameDegenPuntuaction' points
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
	// add proxy puntuaction if user has a proxy
	// (degen:0xA051A2Cb19C00eCDffaE94D0Ff98c17758041D16)
	if rep.HasProxy {
		reputation += proxyPuntuaction
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
	// ensure the reputation does not exceed 100
	if reputation > maxReputation {
		reputation = maxReputation
	}
	return reputation
}

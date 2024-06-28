package reputation

import "github.com/ethereum/go-ethereum/common"

const (
	// User activity max reputation values
	maxFollowersReputation = 10
	maxElectionsReputation = 10
	maxVotesReputation     = 25
	maxCastedReputation    = 45
	maxCommunityReputation = 10
	maxReputation          = 100
	// Boosters puntuaction values
	votecasterNFTPassPuntuaction              = 25
	votecasterLaunchNFTPuntuaction            = 6
	votecasterAlphafrensFollowerPuntuaction   = 15
	votecasterFarcasterFollowerPuntuaction    = 3
	vocdoniFarcasterFollowerPuntuaction       = 3
	votecasterAnnouncementRecastedPuntuaction = 5
	kiwiPuntuaction                           = 3
	degenDAONFTPuntuaction                    = 7
	haberdasheryFTPuntuaction                 = 8
	degenAtLeast10kPuntuaction                = 5
	tokyoDAONFTPuntuaction                    = 5
	proxyStudioNFTPuntuaction                 = 5
	proxyAtLeast5Puntuaction                  = 5
	nameDegenPuntuaction                      = 5
)

// Boosters contract addresses
var (
	// Votecaster NFT Pass contract address
	// TODO: update
	VotecasterNFTPassAddress = common.HexToAddress("0x225D58E18218E8d87f365301aB6eEe4CbfAF820b")
	// Votecaster Launch NFT contract address
	// TODO: update
	VotecasterLaunchNFTAddress = common.HexToAddress("0x32B6BB4d1f7298d4a80c2Ece237e4474C0880B69")
	// Votecaster Alphafrens Channel address
	VotecasterAlphafrensChannelAddress = common.HexToAddress("0xa630fcc62165a3587c6857d73b556c8a61c8edd3")
	// $KIWI token contract address
	KIWIAddress = common.HexToAddress("0x66747bdC903d17C586fA09eE5D6b54CC85bBEA45")
	// DegenDAO NFT contract address
	DegenDAONFTAddress = common.HexToAddress("0x980Fbdd1cF05080781Dca0AEf7026B0406743389")
	// Haberdashery NFT contract address
	HaberdasheryNFTAddress = common.HexToAddress("0x85E7DF5708902bE39891d59aBEf8E21EDE91E8BF")
	// Degen token contract address
	DegenAddress = common.HexToAddress("0x4ed4E862860beD51a9570b96d89aF5E1B0Efefed")
	// TokyoDAO NFT contract address
	TokyoDAONFTAddress = common.HexToAddress("0x432073397Aead241cf2411e21D8fA949183E7151")
	// $PROXY token contract address
	ProxyAddress = common.HexToAddress("0xA051A2Cb19C00eCDffaE94D0Ff98c17758041D16")
	// ProxyStudio NFT contract address
	ProxyStudioNFTAddress = common.HexToAddress("0x7888b1f446c912ddec9bf582629e9ae8845fd8c6")
	// NameDegen NFT contract address
	NameDegenAddress = common.HexToAddress("0x4087fb91A1fBdef05761C02714335D232a2Bf3a1")
)

// Boosters costants (ids, hashesh and network information)
const (
	// Votecaster NFT Pass network short name
	VotecasterNFTPassChainShortName = "base"
	// Votecaster Launch NFT network short name
	VotecasterLaunchNFTChainShortName = "base"
	// Votecaster Farcaster ID
	// TODO: update
	VotecasterFarcasterFID uint64 = 521116
	// Vocdoni Farcaster ID
	VocdoniFarcasterFID uint64 = 7548
	// Votecaster Announcement Farcaster Cast Hash
	VotecasterAnnouncementCastHash = "0xe4528c4931127eb32e4c7c473622d4e3a1c6b0a3"
	// $KIWI token network ID
	KIWIChainID uint64 = 10
	// DegenDAO NFT network short name
	DegenDAONFTChainShortName = "base"
	// Haberdashery NFT network short name
	HaberdasheryNFTChainShortName = "base"
	// Degen token network short name
	DegenChainShortName = "base"
	// TokyoDAO NFT network short name
	TokyoDAONFTChainShortName = "base"
	// $PROXY token network short name
	ProxyChainShortName = "degen"
	// ProxyStudio NFT network short name
	ProxyStudioNFTShortName = "base"
	// NameDegen NFT network short name
	NameDegenChainShortName = "degen"
)

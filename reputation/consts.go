package reputation

import "github.com/ethereum/go-ethereum/common"

var (
	VotecasterNFTPassAddress           = common.HexToAddress("0x0")                                        // VotecasterNFTPass
	VotecasterLaunchNFTAddress         = common.HexToAddress("0x0")                                        // VotecasterLaunchNFT
	VotecasterAlphafrensChannelAddress = common.HexToAddress("0x0")                                        // VotecasterAlphafrens
	KIWIAddress                        = common.HexToAddress("0x66747bdC903d17C586fA09eE5D6b54CC85bBEA45") // KIWI
	DegenDAONFTAddress                 = common.HexToAddress("0x980Fbdd1cF05080781Dca0AEf7026B0406743389") // DegenDAO
	DegenAddress                       = common.HexToAddress("0x4ed4E862860beD51a9570b96d89aF5E1B0Efefed") // Degen
	TokyoDAONFTAddress                 = common.HexToAddress("0x432073397Aead241cf2411e21D8fA949183E7151") // TokyoDAO
	ProxyAddress                       = common.HexToAddress("0xA051A2Cb19C00eCDffaE94D0Ff98c17758041D16") // Proxy
	NameDegenAddress                   = common.HexToAddress("0x4087fb91A1fBdef05761C02714335D232a2Bf3a1") // NameDegen
)

const (
	VotecasterNFTPassChainShortName                  = "eth"   // VotecasterNFTPass
	VotecasterLaunchNFTChainShortName                = "eth"   // VotecasterLaunchNFT
	VotecasterAlphafrensChannelChainShortName        = "eth"   // VotecasterAlphafrens
	VotecasterFarcasterFID                    uint64 = 1       // VotecasterFarcaster
	VocdoniFarcasterFID                       uint64 = 2       // VocdoniFarcaster
	VotecasterAnnouncementCastHash                   = "0x0"   // VotecasterAnnouncementCast
	KIWIChainID                               uint64 = 10      // KIWI
	DegenDAONFTChainShortName                        = "base"  // DegenDAO
	DegenChainShortName                              = "degen" // Degen
	TokyoDAONFTChainShortName                        = "base"  // TokyoDAO
	ProxyChainShortName                              = "degen" // Proxy
	NameDegenChainShortName                          = "degen" // NameDegen
)

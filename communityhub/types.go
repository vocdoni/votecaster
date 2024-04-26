package communityhub

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// CensusType represents the type of census that a community is using to create
// polls
type CensusType string

const (
	// CensusTypeChannel represents the census that includes all the members of
	// a warpcast channel
	CensusTypeChannel CensusType = "channel"
	// CensusTypeERC20 represents the census that includes all the holders of
	// an ERC20 token
	CensusTypeERC20 CensusType = "erc20"
	// CensusTypeNFT represents the census that includes all the holders of an
	// NFT
	CensusTypeNFT CensusType = "nft"
)

const (
	// CONTRACT_CENSUS_TYPE_FC represents the census type for all farcaster
	// users in the CommunityHub contract
	CONTRACT_CENSUS_TYPE_FC = iota
	// CONTRACT_CENSUS_TYPE_CHANNEL represents the census type for all members
	// of a warpcast channel in the CommunityHub contract
	CONTRACT_CENSUS_TYPE_CHANNEL
	// CONTRACT_CENSUS_TYPE_FOLLOWERS represents the census type for all
	// followers of an user in the CommunityHub contract
	CONTRACT_CENSUS_TYPE_FOLLOWERS
	// CONTRACT_CENSUS_TYPE_CSV represents the census type for all addresses
	// in a CSV file (that are also farcaster users) in the CommunityHub
	// contract
	CONTRACT_CENSUS_TYPE_CSV
	// CONTRACT_CENSUS_TYPE_ERC20 represents the census type for all holders
	// of an ERC20 token (that are also farcaster users) in the CommunityHub
	// contract
	CONTRACT_CENSUS_TYPE_ERC20
	// CONTRACT_CENSUS_TYPE_NFT represents the census type for all holders of
	// an NFT (that are also farcaster users) in the CommunityHub contract
	// contract
	CONTRACT_CENSUS_TYPE_NFT
)

var internalCensusTypes = map[uint8]CensusType{
	CONTRACT_CENSUS_TYPE_CHANNEL: CensusTypeChannel,
	CONTRACT_CENSUS_TYPE_ERC20:   CensusTypeERC20,
	CONTRACT_CENSUS_TYPE_NFT:     CensusTypeNFT,
}

// ContractAddress represents the address of a contract in a certain blockchain,
// which is included in this struct
type ContractAddress struct {
	Blockchain string
	Address    common.Address
}

// HubCommunity represents a community in the CommunityHub package
type HubCommunity struct {
	ID             uint64
	Name           string
	ImageURL       string
	GroupChatURL   string
	CensusType     CensusType
	CensusName     string
	CensusAddesses []*ContractAddress
	CensusChannel  string   // warpcast channels id
	Channels       []string // warpcast channels ids
	Admins         []uint64 // farcaster users fids
	Notifications  bool
	Disabled       bool
	// internal
	createElectionPermission uint8
	funds                    *big.Int
}

// HubResult represents the result of a poll in the CommunityHub
type HubResults struct {
	Question         string
	Options          []string
	Date             string
	Tally            [][]*big.Int
	Turnout          *big.Int
	TotalVotingPower *big.Int
	Participants     []*big.Int
	CensusRoot       []byte
	CensusURI        string
	Disabled         bool
	VoteCount        *big.Int
}

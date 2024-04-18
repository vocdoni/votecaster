package communities

import "github.com/ethereum/go-ethereum/common"

type CensusType string

const (
	censusTypeChannel CensusType = "channel"
	censusTypeERC20   CensusType = "erc20"
	censusTypeNFT     CensusType = "nft"

	lastSyncedBlockKey = "communities_hub_last_scanned_block"
)

type Address struct {
	Blockchain string
	Address    common.Address
}

type HubCommunity struct {
	ID             uint64
	Name           string
	ImageUrl       string
	CensusType     CensusType
	CensusName     string
	CensusAddesses []*Address
	CensusChannel  string
	Channels       []string
	Admins         []uint64 // farcaster users fids
	Notifications  bool
}

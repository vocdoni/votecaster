package communities

type HubCommunity struct {
	ID             uint64
	Name           string
	ImageUrl       string
	CensusType     string // "channel", "erc20", "nft"
	CensusName     string
	CensusAddesses []string // forma network:address
	CensusChannel  string
	Channels       []string
	Admins         []uint64 // farcaster users fids
	Notifications  bool
}

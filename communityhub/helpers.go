package communityhub

import (
	comhub "github.com/vocdoni/vote-frame/communityhub/contracts/communityhubtoken"
)

// contractToHub converts a contract community struct (ICommunityHubCommunity)
// to a internal community struct (HubCommunity)
func contractToHub(id uint64, cc comhub.ICommunityHubCommunity) (*HubCommunity, error) {
	// decode admins
	admins := []uint64{}
	for _, bAdmin := range cc.Guardians {
		admins = append(admins, uint64(bAdmin.Int64()))
	}
	// initialize the resulting community struct
	community := &HubCommunity{
		ID:            id,
		Name:          cc.Metadata.Name,
		ImageURL:      cc.Metadata.ImageURI,
		GroupChatURL:  cc.Metadata.GroupChatURL,
		Channels:      cc.Metadata.Channels,
		Admins:        admins,
		Notifications: cc.Metadata.Notifications,
		Disabled:      cc.Disabled,
		// interanl
		createElectionPermission: cc.CreateElectionPermission,
		funds:                    cc.Funds,
	}
	community.CensusType = internalCensusTypes[cc.Census.CensusType]
	// decode census data according to the census type
	switch community.CensusType {
	case CensusTypeChannel:
		// if the census type is a channel, set the channel
		community.CensusChannel = cc.Census.Channel
	case CensusTypeERC20, CensusTypeNFT:
		// if the census type is an erc20 or nft, decode every census network
		// address to get the contract address and blockchain
		community.CensusAddesses = []*ContractAddress{}
		for _, addr := range cc.Census.Tokens {
			community.CensusAddesses = append(community.CensusAddesses, &ContractAddress{
				Blockchain: addr.Blockchain,
				Address:    addr.ContractAddress,
			})
		}
	default:
		return nil, ErrUnknownCensusType
	}
	return community, nil
}

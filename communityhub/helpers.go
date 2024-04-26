package communityhub

import (
	"math/big"

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
		Notifications: &cc.Metadata.Notifications,
		Disabled:      &cc.Disabled,
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

// hubToContract converts a internal community struct (HubCommunity) to a
// contract community struct (ICommunityHubCommunity)
func hubToContract(hcommunity *HubCommunity) (comhub.ICommunityHubCommunity, error) {
	// check the census type
	switch hcommunity.CensusType {
	case CensusTypeChannel:
		if hcommunity.CensusChannel == "" {
			return comhub.ICommunityHubCommunity{}, ErrNoChannelProvided
		}
	case CensusTypeERC20, CensusTypeNFT:
		if len(hcommunity.CensusAddesses) == 0 {
			return comhub.ICommunityHubCommunity{}, ErrBadCensusAddressees
		}
	default:
		return comhub.ICommunityHubCommunity{}, ErrUnknownCensusType
	}
	// convert the census addresses to a []*comhub.ICommunityHubTokenCensusToken
	censusTokens := []comhub.ICommunityHubToken{}
	for _, addr := range hcommunity.CensusAddesses {
		censusTokens = append(censusTokens, comhub.ICommunityHubToken{
			ContractAddress: addr.Address,
			Blockchain:      addr.Blockchain,
		})
	}
	// convert the admins to a []*big.Int
	guardians := []*big.Int{}
	for _, admin := range hcommunity.Admins {
		guardians = append(guardians, new(big.Int).SetUint64(admin))
	}
	// create the contract community
	ccomunity := comhub.ICommunityHubCommunity{
		Metadata: comhub.ICommunityHubCommunityMetadata{
			Name:         hcommunity.Name,
			ImageURI:     hcommunity.ImageURL,
			GroupChatURL: hcommunity.GroupChatURL,
			Channels:     hcommunity.Channels,
		},
		Guardians: guardians,
		Census: comhub.ICommunityHubCensus{
			CensusType: contractCensusTypes[hcommunity.CensusType],
			Channel:    hcommunity.CensusChannel,
			Tokens:     censusTokens,
		},
		CreateElectionPermission: hcommunity.createElectionPermission,
		Funds:                    hcommunity.funds,
	}
	if hcommunity.Notifications != nil {
		ccomunity.Metadata.Notifications = *hcommunity.Notifications
	}
	if hcommunity.Disabled != nil {
		ccomunity.Disabled = *hcommunity.Disabled
	}
	return ccomunity, nil
}

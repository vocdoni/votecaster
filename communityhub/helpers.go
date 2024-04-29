package communityhub

import (
	"fmt"
	"math/big"

	comhub "github.com/vocdoni/vote-frame/communityhub/contracts/communityhubtoken"
	dbmongo "github.com/vocdoni/vote-frame/mongo"
)

// ContractToHub converts a contract community struct (ICommunityHubCommunity)
// to a internal community struct (HubCommunity)
func ContractToHub(id uint64, cc comhub.ICommunityHubCommunity) (*HubCommunity, error) {
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

// HubToContract converts a internal community struct (HubCommunity) to a
// contract community struct (ICommunityHubCommunity)
func HubToContract(hcommunity *HubCommunity) (comhub.ICommunityHubCommunity, error) {
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

// HubToDB converts a internal community struct (HubCommunity) to a db community
// struct (*dbmongo.Community) to be stored or updated in the database. It
// creates the db census according to the community census type, and if the
// census type is a channel, it sets the channel. If the census type is an erc20
// or nft, it decodes every census network address to get the contract address
// and blockchain. It returns an error if the census type is unknown, if no
// channel is provided when the census type is a channel, or if no valid
// addresses were found when the census type is an erc20 or nft.
func HubToDB(hcommunity *HubCommunity) (*dbmongo.Community, error) {
	// create the db census according to the community census type
	dbCensus := dbmongo.CommunityCensus{
		Type: string(hcommunity.CensusType),
	}
	// if the census type is a channel, set the channel
	switch hcommunity.CensusType {
	case CensusTypeChannel:
		if hcommunity.CensusChannel == "" {
			return nil, fmt.Errorf("%w: %s", ErrNoChannelProvided, hcommunity.Name)
		}
		dbCensus.Channel = hcommunity.CensusChannel
	case CensusTypeERC20, CensusTypeNFT:
		// if the census type is an erc20 or nft, decode every census
		// network address to get the contract address and blockchain
		dbCensus.Addresses = []dbmongo.CommunityCensusAddresses{}
		for _, addr := range hcommunity.CensusAddesses {
			dbCensus.Addresses = append(dbCensus.Addresses,
				dbmongo.CommunityCensusAddresses{
					Address:    addr.Address.String(),
					Blockchain: addr.Blockchain,
				})
		}
		// if no valid addresses were found, skip the community and log
		// an error
		if len(dbCensus.Addresses) == 0 {
			return nil, fmt.Errorf("%w: %s", ErrBadCensusAddressees, hcommunity.Name)
		}
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownCensusType, hcommunity.CensusType)
	}
	// create the db community
	dbCommunity := &dbmongo.Community{
		ID:           hcommunity.ID,
		Name:         hcommunity.Name,
		ImageURL:     hcommunity.ImageURL,
		GroupChatURL: hcommunity.GroupChatURL,
		Census:       dbCensus,
		Channels:     hcommunity.Channels,
		Admins:       hcommunity.Admins,
	}
	// set the notifications and disabled fields if they are not nil
	if hcommunity.Notifications != nil {
		dbCommunity.Notifications = *hcommunity.Notifications
	}
	if hcommunity.Disabled != nil {
		dbCommunity.Disabled = *hcommunity.Disabled
	}
	return dbCommunity, nil
}

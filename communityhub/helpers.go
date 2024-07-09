package communityhub

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	comhub "github.com/vocdoni/vote-frame/communityhub/contracts/communityhubtoken"
	dbmongo "github.com/vocdoni/vote-frame/mongo"
)

const (
	// farcasterUserRefPrefix is the prefix used to encode a user reference from
	// farcaster to a user FID.
	farcasterUserRefPrefix = "fid:"
	// chainPrefixFormat is the format used to encode a chain prefixed content.
	chainPrefixFormat = "%s:%s"
	// chainPrefixSeparator is the separator used to encode a chain prefixed
	// content.
	chainPrefixSeparator = ":"
)

// UserRefToFID converts a user reference to a farcaster user to a user FID. It
// is used to encode the user reference from farcaster in CensusChannel contract
// field to support followers censuses from farcaster.
func UserRefToFID(userRef string) (uint64, error) {
	if strings.HasPrefix(userRef, farcasterUserRefPrefix) {
		return strconv.ParseUint(userRef[len(farcasterUserRefPrefix):], 10, 64)
	}
	return 0, fmt.Errorf("invalid user reference: %s", userRef)
}

// ContractToHub converts a contract community struct (ICommunityHubCommunity)
// to a internal community struct (HubCommunity)
func ContractToHub(contractID, chainID uint64, communityID string, cc comhub.ICommunityHubCommunity) (*HubCommunity, error) {
	// decode admins
	admins := []uint64{}
	for _, bAdmin := range cc.Guardians {
		admins = append(admins, uint64(bAdmin.Int64()))
	}
	// initialize the resulting community struct
	community := &HubCommunity{
		CommunityID:   communityID,
		ContractID:    contractID,
		ChainID:       chainID,
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
	case CensusTypeChannel, CensusTypeFollowers:
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
	case CensusTypeFollowers:
		if hcommunity.CensusChannel == "" {
			return comhub.ICommunityHubCommunity{}, ErrNoUserRefProvided
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
	case CensusTypeChannel, CensusTypeFollowers:
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
		ID:           hcommunity.CommunityID,
		Name:         hcommunity.Name,
		ImageURL:     hcommunity.ImageURL,
		GroupChatURL: hcommunity.GroupChatURL,
		Census:       dbCensus,
		Channels:     hcommunity.Channels,
		Creator:      hcommunity.Admins[0],
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

// DBToHub converts a db community struct (*dbmongo.Community) to a internal
// community struct (HubCommunity) to be used in the CommunityHub package. It
// decodes the census addresses according to the census type, and if the census
// type is a channel, it sets the channel. If the census type is an erc20 or nft,
// it decodes every census network address to get the contract address and
// blockchain. It returns an error if the census type is unknown.
func DBToHub(dbCommunity *dbmongo.Community, contractID, chainID uint64) (*HubCommunity, error) {
	censusAddresses := []*ContractAddress{}
	if ct := CensusType(dbCommunity.Census.Type); ct == CensusTypeERC20 || ct == CensusTypeNFT {
		for _, addr := range dbCommunity.Census.Addresses {
			censusAddresses = append(censusAddresses, &ContractAddress{
				Blockchain: addr.Blockchain,
				Address:    common.HexToAddress(addr.Address),
			})
		}
	}
	admins := []uint64{dbCommunity.Creator}
	for _, admin := range dbCommunity.Admins {
		if admin == dbCommunity.Creator {
			continue
		}
		admins = append(admins, admin)
	}
	return &HubCommunity{
		CommunityID:    dbCommunity.ID,
		ContractID:     contractID,
		ChainID:        chainID,
		Name:           dbCommunity.Name,
		ImageURL:       dbCommunity.ImageURL,
		GroupChatURL:   dbCommunity.GroupChatURL,
		CensusType:     CensusType(dbCommunity.Census.Type),
		CensusAddesses: censusAddresses,
		CensusChannel:  dbCommunity.Census.Channel,
		Channels:       dbCommunity.Channels,
		Admins:         admins,
		Notifications:  &dbCommunity.Notifications,
		Disabled:       &dbCommunity.Disabled,
	}, nil
}

// EncodePrefix encodes a content with a prefix following the format
// "prefix:content".
func EncodePrefix(prefix, content string) string {
	return fmt.Sprintf(chainPrefixFormat, prefix, content)
}

// DecodePrefix decodes a prefixed content following the format "prefix:content"
// and returns the prefix and the content separately.
func DecodePrefix(prefixed string) (string, string, bool) {
	if prefixed == "" {
		return "", "", false
	}
	// split the community ID into the chain short name and the ID
	parts := strings.Split(prefixed, chainPrefixSeparator)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

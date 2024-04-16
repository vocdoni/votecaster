package communities

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/log"
)

const zeroAddress = "0x0000000000000000000000000000000000000000"

type CommunitiesHubListenerConfig struct {
	ContractAddress string
	ChainID         uint64
	DB              *mongo.MongoStorage
}

type CommunitiesHubListener struct {
	db               *mongo.MongoStorage
	ctx              context.Context
	waiter           sync.WaitGroup
	cancel           context.CancelFunc
	lastScannedBlock atomic.Uint64

	Address common.Address
	ChainID uint64
}

func NewCommunitiesHubListener(
	goblalCtx context.Context,
	conf *CommunitiesHubListenerConfig,
) (*CommunitiesHubListener, error) {
	if conf.DB == nil {
		return nil, ErrMissingDB
	}
	// TODO: init the contract and request a RPC endpoint to dial
	// the contract
	ctx, cancel := context.WithCancel(goblalCtx)
	return &CommunitiesHubListener{
		db:               conf.DB,
		ctx:              ctx,
		cancel:           cancel,
		waiter:           sync.WaitGroup{},
		lastScannedBlock: atomic.Uint64{},
		Address:          common.HexToAddress(conf.ContractAddress),
		ChainID:          conf.ChainID,
	}, nil
}

func (l *CommunitiesHubListener) Start() {
	log.Infow("starting communities hub listener",
		"contract", l.Address.String(),
		"chainID", l.ChainID)
	// scan for new logs in background
	communitiesCh := make(chan HubCommunity)
	l.waiter.Add(1)
	go func() {
		defer l.waiter.Done()

	}()
	// handle new logs in background and create new communities in the database
	l.waiter.Add(1)
	go func() {
		defer l.waiter.Done()
		for c := range communitiesCh {
			// create the db census according to the community census type
			dbCensus := mongo.CommunityCensus{
				Type: c.CensusType,
				Name: c.CensusName,
			}
			// if the census type is a channel, set the channel
			if c.CensusType == mongo.TypeCommunityCensusChannel {
				dbCensus.Channel = c.CensusChannel
			} else {
				// if the census type is an erc20 or nft, decode every census
				// network address to get the contract address and blockchain
				dbCensus.Addresses = []mongo.CommunityAddresses{}
				for _, addr := range c.CensusAddesses {
					address, blockchain, err := decodeNetworkAddress(addr)
					if err != nil {
						log.Errorf("failed to decode network address: %s", err)
						continue
					}
					dbCensus.Addresses = append(dbCensus.Addresses, mongo.CommunityAddresses{
						Address:    address.String(),
						Blockchain: blockchain,
					})
				}
				// if no valid addresses were found, skip the community and log
				// an error
				if len(dbCensus.Addresses) == 0 {
					log.Errorf("no valid addresses found for community: %s", c.Name)
					continue
				}
			}
			// create community in the database
			if err := l.db.AddCommunity(c.ID, c.Name, c.ImageUrl, dbCensus,
				c.Channels, c.Admins, c.Notifications,
			); err != nil {
				log.Errorf("failed to add community: %s", err)
				continue
			}
		}
	}()
}

func (l *CommunitiesHubListener) Stop() {
	log.Info("stopping communities hub listener")
	l.cancel()
	l.waiter.Wait()
}

// decodeNetworkAddress decodes a network address string into an ethereum address
// and a blockchain name. The network address string must be in the format
// <network>:<contractAddress>, for example:
//
//	base:0x225D58E18218E8d87f365301aB6eEe4CbfAF820b
func decodeNetworkAddress(networkAddress string) (common.Address, string, error) {
	parts := strings.Split(networkAddress, ":")
	if len(parts) != 2 {
		return common.Address{}, "", fmt.Errorf("invalid network address: %s", networkAddress)
	}
	strAddr, blockchain := parts[1], parts[0]
	if strAddr == "" || blockchain == "" {
		return common.Address{}, "", fmt.Errorf("invalid address or network: %s:%s", strAddr, blockchain)
	}
	addr := common.HexToAddress(strAddr)
	if addr.String() == zeroAddress {
		return addr, blockchain, fmt.Errorf("invalid address: %s -> %s", strAddr, zeroAddress)
	}
	return addr, blockchain, nil
}

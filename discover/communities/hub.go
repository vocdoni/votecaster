package communities

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/log"
)

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
	community := &CommunitiesHubListener{
		db:               conf.DB,
		ctx:              ctx,
		cancel:           cancel,
		waiter:           sync.WaitGroup{},
		lastScannedBlock: atomic.Uint64{},
		Address:          common.HexToAddress(conf.ContractAddress),
		ChainID:          conf.ChainID,
	}
	// get the last scanned block from the database
	dbLastScannedBlock, err := conf.DB.Metadata(lastSyncedBlockKey)
	if err != nil && !strings.Contains(err.Error(), "no documents in result") {
		log.Errorf("failed to get last scanned block: %s", err)
	}
	if iLastSyncedBlock, ok := dbLastScannedBlock.(uint64); ok {
		community.lastScannedBlock.Store(iLastSyncedBlock)
	}
	return community, nil
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
		// 1. SCAN BLOCKS
		// scan from the last scanned block to the end of the iteration
		// the iteration ends in the last scanned block + 1000000 or the last
		// block in the chain, whichever is smaller
		// scan from the last scanned block to the end in batches of 2000 blocks
		// (or less if the end of the iteration is reached)
		// At the end of every batch, update the last scanned block in the
		// database

		// 2. FILTER LOGS
		// filter the logs to only get the ones that are related to the
		// communities creation: 'CommunityCreated'
		//  - using FilterCommunityCreated
		//  - and ParseCommunityCreated
		// which returns the id of the community and the address of the creator

		// 3. GET COMMUNITY INFO
		// get the community info using the community id calling to the contract
		// to get the community data using:
		//  - GetCommunity(id)
		// which returns the ICommunityHubCommunity and must be parsed to a
		// HubCommunity struct:
		//  - ID (previous id) -> ID
		//  - Metadata.Name -> Name
		//  - Metadata.ImageUrl -> ImageUrl
		//  - Metadata.Channels -> Channels
		//  - Metadata.Notifications -> Notifications
		//  - Guardians -> Admins
		//  - Census.CensusType -> CensusType
		//  - Census.Name -> CensusName
		//  - Census.Channel -> CensusChannel
		//  - Census.Addresses -> CensusAddesses
	}()
	// handle new logs in background and create new communities in the database
	l.waiter.Add(1)
	go func() {
		defer l.waiter.Done()
		for c := range communitiesCh {
			// create the db census according to the community census type
			dbCensus := mongo.CommunityCensus{
				Type: string(c.CensusType),
				Name: c.CensusName,
			}
			// if the census type is a channel, set the channel
			if c.CensusType == censusTypeChannel {
				dbCensus.Channel = c.CensusChannel
			} else {
				// if the census type is an erc20 or nft, decode every census
				// network address to get the contract address and blockchain
				dbCensus.Addresses = []mongo.CommunityAddresses{}
				for _, addr := range c.CensusAddesses {
					dbCensus.Addresses = append(dbCensus.Addresses, mongo.CommunityAddresses{
						Address:    addr.Address.String(),
						Blockchain: addr.Blockchain,
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

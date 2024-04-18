package communityhub

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	c3web3 "github.com/vocdoni/census3/helpers/web3"
	comhub "github.com/vocdoni/vote-frame/communityhub/contracts/communityhubtoken"
	dbmongo "github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/log"
)

const (
	// DefaultScannerCooldown is the default time that the scanner sleeps
	// between scan iterations
	DefaultScannerCooldown = time.Second * 10

	maxBlocksPerIteration = 1000000
	maxBlocksPerBatch     = 2000
)

// CommunityHubConfig struct defines the configuration for the CommunityHub.
// It includes the contract address, the start block, the chain ID where the
// contract is deployed, a database instance, and the scanner cooldown (by
// default 10s (DefaultScannerCooldown)).
type CommunityHubConfig struct {
	ContractAddress common.Address
	StartBlock      uint64
	ChainID         uint64
	DB              *dbmongo.MongoStorage
	ScannerCooldown time.Duration
}

// CommunityHub struct defines the CommunityHub wrapper. It includes the
// the functions to scan for new communities in the contract and create them
// in the database in background. It also includes some functions to get
// communities or set and get results using the contract.
type CommunityHub struct {
	db               *dbmongo.MongoStorage
	ctx              context.Context
	waiter           sync.WaitGroup
	cancel           context.CancelFunc
	lastScannedBlock atomic.Uint64
	w3cli            *c3web3.Client
	contract         *comhub.CommunityHubToken
	scannerCooldown  time.Duration

	Address common.Address
	ChainID uint64
}

// NewCommunityHub function initializes a new CommunityHub instance. It returns
// an error if the database is not defined in the configuration or if the web3
// client cannot be initialized. It initializes the contract with the web3 client
// and the contract address, and sets the last scanned block from the database.
// It also sets the scanner cooldown from the configuration if it is defined, or
// uses the default one. It receives the global context, the web3 pool, and the
// configuration of the CommunityHub.
func NewCommunityHub(
	goblalCtx context.Context,
	w3p *c3web3.Web3Pool,
	conf *CommunityHubConfig,
) (*CommunityHub, error) {
	// check that the database is defined in the provided configuration
	if conf.DB == nil {
		return nil, ErrMissingDB
	}
	// initialize the web3 client for the chain
	w3cli, err := w3p.Client(conf.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get web3 client: %w", err)
	}
	// initialize the contract with the web3 client and the contract address
	contract, err := comhub.NewCommunityHubToken(conf.ContractAddress, w3cli)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}
	// initialize the context and the listener
	ctx, cancel := context.WithCancel(goblalCtx)
	community := &CommunityHub{
		db:               conf.DB,
		ctx:              ctx,
		cancel:           cancel,
		waiter:           sync.WaitGroup{},
		lastScannedBlock: atomic.Uint64{},
		w3cli:            w3cli,
		contract:         contract,
		Address:          conf.ContractAddress,
		ChainID:          conf.ChainID,
	}
	// get the last scanned block from the database, by default the token
	// creation block
	lastBlock := conf.StartBlock
	// set the last scanned block in the listener and return it
	community.lastScannedBlock.Store(lastBlock)
	// set the scanner cooldown from the configuration if it is defined, or use
	// the default one
	if community.scannerCooldown = DefaultScannerCooldown; conf.ScannerCooldown > 0 {
		community.scannerCooldown = conf.ScannerCooldown
	}
	return community, nil
}

// ScanNewCommunities method starts the listener to scan for new communities
// in the contract and create them in the database in background. It starts two
// goroutines, one to scan for new logs in the contract and submit them to a
// channel, and another one to handle the new logs and create the communities
// in the database. It calculates the bounds of the iteration and the batch,
// and iterates between them to get the logs. It updates the last scanned block
// in the database after each iteration.
func (l *CommunityHub) ScanNewCommunities() {
	log.Infow("starting communities hub scanner",
		"contract", l.Address.String(),
		"chainID", l.ChainID)
	// scan for new logs in background
	communitiesCh := make(chan *HubCommunity)
	l.waiter.Add(1)
	go func() {
		defer l.waiter.Done()
		iteration := 0
		for {
			select {
			case <-l.ctx.Done():
				close(communitiesCh)
				return
			default:
				iteration++
				// calculate the bounds of the iteration
				currentBlock := l.lastScannedBlock.Load()
				startBlock, endBlock, err := l.iterationBounds(currentBlock)
				if err != nil {
					log.Warnw("error getting iteration bounds",
						"error", err,
						"lastBlock", currentBlock)
					continue
				}
				log.Infow("scanning for new communities",
					"iteration", iteration,
					"fromBlock", startBlock,
					"toBlock", endBlock)
				// iterate between bounds in batches
				for startBlock < endBlock {
					// calculate the end of the batch
					endBatchBlock := l.batchEndBlock(startBlock, endBlock)
					// filter the community creation logs
					logs, err := l.contract.FilterCommunityCreated(&bind.FilterOpts{
						Start: startBlock,
						End:   &endBatchBlock,
					})
					if err != nil {
						log.Warnw("error getting logs",
							"error", err,
							"fromBlock", startBlock,
							"toBlock", endBatchBlock)
						continue
					}
					// submit the community creation logs to the channel
					if err := l.submitCommunityCreation(logs, communitiesCh); err != nil {
						log.Warnw("error submitting community creation",
							"error", err,
							"fromBlock", startBlock,
							"toBlock", endBatchBlock)
						return
					}
					log.Debugw("communities logs batch processed",
						"fromBlock", startBlock,
						"toBlock", endBatchBlock)
					// update the start block for the next batch
					startBlock += maxBlocksPerBatch
					l.lastScannedBlock.Store(endBatchBlock)
				}
				// update the last scanned block
				l.lastScannedBlock.Store(endBlock)
				log.Infow("iteration finished",
					"iteration", iteration,
					"lastBlock", startBlock)
				time.Sleep(l.scannerCooldown)
			}
		}
	}()
	// handle new logs in background and create new communities in the database
	l.waiter.Add(1)
	go func() {
		defer l.waiter.Done()
		for c := range communitiesCh {
			l.storeCommunity(c)
			log.Infow("community created",
				"communityID", c.ID,
				"name", c.Name,
				"censusType", c.CensusType,
				"censusChannel", c.CensusChannel,
				"censusAddresses", c.CensusAddesses)
		}
	}()
}

// Stop method stops the listener and waits for the goroutines to finish.
func (l *CommunityHub) Stop() {
	log.Info("stopping communities hub scanner")
	l.cancel()
	l.waiter.Wait()
}

// CommunityFromContract  method gets the community data from the contract and
// returns it as a HubCommunity struct. It decodes the admins and census
// addresses from the contract data. If something goes wrong getting the
// community data, it returns an error.
func (l *CommunityHub) CommunityFromContract(communityID uint64) (*HubCommunity, error) {
	// get the community data from the contract
	cc, err := l.contract.GetCommunity(nil, new(big.Int).SetUint64(communityID))
	if err != nil {
		return nil, errors.Join(ErrGettingCommunity, err)
	}
	// decode admins
	admins := []uint64{}
	for _, bAdmin := range cc.Guardians {
		admins = append(admins, bAdmin.Uint64())
	}
	// initialize the resulting community struct
	community := &HubCommunity{
		ID:            communityID,
		Name:          cc.Metadata.Name,
		ImageURL:      cc.Metadata.ImageURI,
		GroupChatURL:  cc.Metadata.GroupChatURL,
		Channels:      cc.Metadata.Channels,
		Admins:        admins,
		Notifications: cc.Metadata.Notifications,
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
		return nil, errors.Join(ErrDecodingCreationLog, fmt.Errorf("unknown census type: %d", cc.Census.CensusType))
	}
	// return the community data
	return community, nil
}

// storeCommunity helper method creates a new community in the database. It
// creates the db census according to the community census type, and if the
// census type is a channel, it sets the channel. If the census type is an erc20
// or nft, it decodes every census network address to get the contract address
// and blockchain. If no valid addresses were found, it skips the community and
// logs an error. Then, it creates the community in the database. If something
// goes wrong creating the community, it logs an error.
func (l *CommunityHub) storeCommunity(c *HubCommunity) {
	// create the db census according to the community census type
	dbCensus := dbmongo.CommunityCensus{
		Type: string(c.CensusType),
		Name: c.CensusName,
	}
	// if the census type is a channel, set the channel
	switch c.CensusType {
	case CensusTypeChannel:
		dbCensus.Channel = c.CensusChannel
	case CensusTypeERC20, CensusTypeNFT:
		// if the census type is an erc20 or nft, decode every census
		// network address to get the contract address and blockchain
		dbCensus.Addresses = []dbmongo.CommunityCensusAddresses{}
		for _, addr := range c.CensusAddesses {
			dbCensus.Addresses = append(dbCensus.Addresses,
				dbmongo.CommunityCensusAddresses{
					Address:    addr.Address.String(),
					Blockchain: addr.Blockchain,
				})
		}
		// if no valid addresses were found, skip the community and log
		// an error
		if len(dbCensus.Addresses) == 0 {
			log.Errorf("no valid addresses found for community: %s", c.Name)
			return
		}
	default:
		log.Errorf("unknown census type: %s", c.CensusType)
		return
	}
	// create community in the database
	if err := l.db.AddCommunity(c.ID, c.Name, c.ImageURL, c.GroupChatURL,
		dbCensus, c.Channels, c.Admins, c.Notifications,
	); err != nil {
		log.Errorf("failed to add community: %s", err)
	}
}

// iterationEndBlock helper method calculates the end block of the iteration
// based on the start block provided. It gets the last block in the chain and
// returns the last block if it is less than the max last block, otherwise it
// returns the max last block.
func (l *CommunityHub) iterationBounds(startBlock uint64) (uint64, uint64, error) {
	// get the last block in the chain, if success and the is less than the max
	// last block, return the last block, otherwise return the max last block
	ctx, cancel := context.WithTimeout(l.ctx, time.Second*5)
	defer cancel()
	lastBlock, err := l.w3cli.BlockNumber(ctx)
	if err != nil {
		return 0, 0, err
	}
	// if the start block is greater than the last block, return the last block
	if startBlock > lastBlock {
		return lastBlock, lastBlock, nil
	}
	maxLastBlock := startBlock + maxBlocksPerIteration - 1
	if maxLastBlock < lastBlock {
		return startBlock, maxLastBlock, nil
	}
	return startBlock, lastBlock, nil
}

// batchEndBlock helper method calculates the end block of the batch based on
// the start block and the end block of the iteration. It returns the end block
// of the batch if it is less than the end block of the iteration, otherwise it
// returns the end block of the iteration.
func (l *CommunityHub) batchEndBlock(startBatchBlock, endIterBlock uint64) uint64 {
	endBatchBlock := startBatchBlock + maxBlocksPerBatch - 1
	if endBatchBlock > endIterBlock {
		return endIterBlock
	}
	return endBatchBlock
}

// submitCommunityCreation helper method submits the community creation logs to
// the channel. It iterates over the logs and decodes the community data from
// the contract. If something goes wrong getting the community data, it returns
// an error.
func (l *CommunityHub) submitCommunityCreation(
	iter *comhub.CommunityHubTokenCommunityCreatedIterator, comCh chan *HubCommunity,
) error {
	for iter.Next() {
		if iter.Event == nil || iter.Event.CommunityId == nil {
			continue
		}
		communityID := iter.Event.CommunityId.Uint64()
		community, err := l.CommunityFromContract(communityID)
		if err != nil {
			return err
		}
		comCh <- community
	}
	return nil
}

package communityhub

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	c3web3 "github.com/vocdoni/census3/helpers/web3"
	comhub "github.com/vocdoni/vote-frame/communityhub/contracts/communityhubtoken"
	dbmongo "github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/log"
)

// DefaultScannerCooldown is the default time that the scanner sleeps between
// scan iterations
const DefaultScannerCooldown = time.Second * 5

// zeroHexAddr is the zero address in hex format used to find new communities,
// discarding empty responses from the contract
const zeroHexAddr = "0x0000000000000000000000000000000000000000"

// CommunityHubConfig struct defines the configuration for the CommunityHub.
// It includes the contract address, the chain ID where the
// contract is deployed, a database instance, and the scanner cooldown (by
// default 10s (DefaultScannerCooldown)).
type CommunityHubConfig struct {
	ContractAddress common.Address
	ChainID         uint64
	DB              *dbmongo.MongoStorage
	ScannerCooldown time.Duration
}

// CommunityHub struct defines the CommunityHub wrapper. It includes the
// the functions to scan for new communities in the contract and create them
// in the database in background. It also includes some functions to get
// communities or set and get results using the contract.
type CommunityHub struct {
	db              *dbmongo.MongoStorage
	ctx             context.Context
	waiter          sync.WaitGroup
	cancel          context.CancelFunc
	nextCommunityID atomic.Uint64
	w3cli           *c3web3.Client
	contract        *comhub.CommunityHubToken
	scannerCooldown time.Duration

	Address common.Address
	ChainID uint64
}

// NewCommunityHub function initializes a new CommunityHub instance. It returns
// an error if the database is not defined in the configuration or if the web3
// client cannot be initialized. It initializes the contract with the web3
// client and the contract address, and sets the next community candidate to be
// created from the database to start to scan for new communities. It also sets
// the scanner cooldown from the configuration if it is defined, or uses the
// default one. It receives the global context, the web3 pool, and the
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
		db:              conf.DB,
		ctx:             ctx,
		cancel:          cancel,
		waiter:          sync.WaitGroup{},
		nextCommunityID: atomic.Uint64{},
		w3cli:           w3cli,
		contract:        contract,
		Address:         conf.ContractAddress,
		ChainID:         conf.ChainID,
	}
	// check the last community ID in the database and set it if it is defined
	if nextCommunityID, err := conf.DB.NextCommunityID(); err == nil {
		community.nextCommunityID.Store(nextCommunityID)
	}
	// set the scanner cooldown from the configuration if it is defined, or use
	// the default one
	if community.scannerCooldown = DefaultScannerCooldown; conf.ScannerCooldown > 0 {
		community.scannerCooldown = conf.ScannerCooldown
	}
	return community, nil
}

// ScanNewCommunities method starts the listener to scan for new communities
// in the contract and create them in the database in background. It gets
// the next community from the contract and, if it exists, it is created it
// in the database. If something goes wrong getting the community data or
// creating the community in the database, it logs an error and retries in
// the next iteration. It sleeps if no communities are found in the contract.
func (l *CommunityHub) ScanNewCommunities() {
	log.Infow("starting communities hub scanner",
		"contract", l.Address.String(),
		"chainID", l.ChainID)
	// scan for new logs in background
	l.waiter.Add(1)
	go func() {
		defer l.waiter.Done()
		for {
			select {
			case <-l.ctx.Done():
				return
			default:
				time.Sleep(l.scannerCooldown)
				// get the next community from the contract
				currentID := l.nextCommunityID.Load()
				newCommunity, err := l.CommunityFromContract(currentID)
				if err != nil {
					if !errors.Is(err, ErrCommunityNotFound) {
						log.Warnw("failed to get community from contract",
							"communityID", currentID,
							"error", err)
					}
					continue
				}
				log.Infow("new community found",
					"communityID", currentID,
					"name", newCommunity.Name,
					"censusType", newCommunity.CensusType,
					"censusChannel", newCommunity.CensusChannel,
					"censusAddresses", newCommunity.CensusAddesses)
				if err := l.storeCommunity(newCommunity); err != nil {
					// TODO: Change the admins of wrong communities
					if strings.Contains(err.Error(), "overflows int64") {
						log.Warnw("admins fids overflowed, skipping community",
							"communityID", newCommunity.ID,
							"admins", newCommunity.Admins)
						l.nextCommunityID.Add(1)
						continue
					}
					log.Warnw("failed to store community in database",
						"communityID", currentID,
						"error", err)
					continue
				}
				l.nextCommunityID.Add(1)
				log.Infow("community created",
					"communityID", newCommunity.ID)
			}
		}
	}()
}

// Stop method stops the listener and waits for the goroutines to finish.
func (l *CommunityHub) Stop() {
	log.Info("stopping communities hub scanner")
	l.cancel()
	l.waiter.Wait()
}

// CommunityFromContract method gets the community data from the contract and
// returns it as a HubCommunity struct. It decodes the admins and census
// addresses from the contract data. It checks if the community is was found
// in the contract by comparing the election results contract address with the
// zero address. It returns the community data, and if something goes wrong,
// it returns an error.
func (l *CommunityHub) CommunityFromContract(communityID uint64) (*HubCommunity, error) {
	// get the community data from the contract
	cc, err := l.contract.GetCommunity(nil, new(big.Int).SetUint64(communityID))
	if err != nil {
		return nil, errors.Join(ErrGettingCommunity, err)
	}
	// check if the community is not found in the contract, to do that, compare
	// the election results contract address with the zero address
	if cc.ElectionResultsContract.String() == zeroHexAddr {
		return nil, ErrCommunityNotFound
	}
	// decode admins
	admins := []uint64{}
	for _, bAdmin := range cc.Guardians {
		admins = append(admins, uint64(bAdmin.Int64()))
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
		return nil, errors.Join(ErrDecodingCommunity, fmt.Errorf("unknown census type: %d", cc.Census.CensusType))
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
// goes wrong creating the community, it returns an error.
func (l *CommunityHub) storeCommunity(c *HubCommunity) error {
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
			return fmt.Errorf("no valid addresses found for community: %s", c.Name)
		}
	default:
		return fmt.Errorf("unknown census type: %s", c.CensusType)
	}
	// create community in the database
	if err := l.db.AddCommunity(c.ID, c.Name, c.ImageURL, c.GroupChatURL,
		dbCensus, c.Channels, c.Admins, c.Notifications,
	); err != nil {
		return fmt.Errorf("failed to add community: %w", err)
	}
	return nil
}

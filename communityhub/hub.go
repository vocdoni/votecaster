package communityhub

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
// It includes the contract address, the chain ID where the contract is
// deployed, a database instance, and the scanner cooldown (by default 10s
// (DefaultScannerCooldown)).
type CommunityHubConfig struct {
	ContractAddress common.Address
	ChainID         uint64
	DB              *dbmongo.MongoStorage
	PrivKey         string
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
	privKey         *ecdsa.PrivateKey
	privAddress     common.Address

	ContractAddress common.Address
	ChainID         uint64
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
		return nil, errors.Join(ErrWeb3Client, err)
	}
	// initialize the contract with the web3 client and the contract address
	contract, err := comhub.NewCommunityHubToken(conf.ContractAddress, w3cli)
	if err != nil {
		return nil, errors.Join(ErrInitContract, err)
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
		ContractAddress: conf.ContractAddress,
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
	// parse the private key if it is defined
	if conf.PrivKey != "" {
		community.privKey, err = crypto.HexToECDSA(conf.PrivKey)
		if err != nil {
			log.Warnw("failed to parse CommunityHub private key", "error", err)
			return community, nil
		}
		community.privAddress = crypto.PubkeyToAddress(community.privKey.PublicKey)
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
		"contract", l.ContractAddress.String(),
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
				newCommunity, err := l.Community(currentID)
				if err != nil {
					// if community is disabled, skip it incrementing the
					// next community ID, otherwise log the error unless the
					// community is not found
					switch err {
					case ErrDisabledCommunity:
						l.nextCommunityID.Add(1)
					case ErrCommunityNotFound:
						break
					default:
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

// Community method gets the community data from the contract and returns it
// as a HubCommunity struct. It decodes the admins and census addresses from
// the contract data. It checks if the community is was found in the contract
// by comparing the election results contract address with the zero address.
// It returns the community data, and if something goes wrong, it returns an
// error.
func (l *CommunityHub) Community(communityID uint64) (*HubCommunity, error) {
	// get the community data from the contract
	cc, err := l.contract.GetCommunity(nil, new(big.Int).SetUint64(communityID))
	if err != nil {
		return nil, errors.Join(ErrGettingCommunity, err)
	}
	// return the community data
	return contractToHub(communityID, cc)
}

// Results method gets the election results using the community and elections
// IDs from the contract and returns them as a HubResults struct. If something
// goes wrong getting the results from the contract, it returns an error.
func (l *CommunityHub) Results(communityID uint64, electionID []byte) (*HubResults, error) {
	// convert the community ID to a *big.Int
	bCommunityID := new(big.Int).SetUint64(communityID)
	// convert the election ID to a [32]byte
	bElectionID := [32]byte{}
	copy(bElectionID[:], electionID)
	// get the election results from the contract
	contractResults, err := l.contract.GetResult(nil, bCommunityID, bElectionID)
	if err != nil {
		return nil, errors.Join(ErrGettingResults, err)
	}
	// return the results struct
	return &HubResults{
		Question:         contractResults.Question,
		Options:          contractResults.Options,
		Date:             contractResults.Date,
		Turnout:          contractResults.Turnout,
		TotalVotingPower: contractResults.TotalVotingPower,
		Participants:     contractResults.Participants,
		CensusRoot:       contractResults.CensusRoot[:],
		CensusURI:        contractResults.CensusURI,
	}, nil
}

// SetResults method sets the election results provided to the community and
// election IDs provided. If something goes wrong setting the results in the
// contract, it returns an error.
func (l *CommunityHub) SetResults(communityID uint64, electionID []byte, results *HubResults) error {
	transactOpts, err := l.authTransactOpts()
	if err != nil {
		return err
	}
	// convert the community ID to a *big.Int
	bCommunityID := new(big.Int).SetUint64(communityID)
	// convert the election ID to a [32]byte
	bElectionID := [32]byte{}
	copy(bElectionID[:], electionID)
	// convert census root to a [32]byte
	bCensusRoot := [32]byte{}
	copy(bCensusRoot[:], results.CensusRoot)
	// set the election results in the contract
	if _, err := l.contract.SetResult(transactOpts, bCommunityID, bElectionID,
		comhub.IElectionResultsResult{
			Question:         results.Question,
			Options:          results.Options,
			Date:             results.Date,
			Turnout:          results.Turnout,
			TotalVotingPower: results.TotalVotingPower,
			Participants:     results.Participants,
			CensusRoot:       bCensusRoot,
			CensusURI:        results.CensusURI,
		}); err != nil {
		return errors.Join(ErrSettingResults, err)
	}
	return nil
}

// DisableCommunity method disables the community in the contract by the
// community ID provided. If something goes wrong disabling the community in
// the contract, it returns an error.
func (l *CommunityHub) DisableCommunity(communityID uint64) error {
	transactOpts, err := l.authTransactOpts()
	if err != nil {
		return err
	}
	// convert the community ID to a *big.Int
	bCommunityID := new(big.Int).SetUint64(communityID)
	// get the community data from the contract
	community, err := l.contract.GetCommunity(nil, bCommunityID)
	if err != nil {
		return err
	}
	// disable the community in the contract
	_, err = l.contract.AdminManageCommunity(transactOpts,
		bCommunityID,
		community.Metadata,
		community.Census,
		community.Guardians,
		community.ElectionResultsContract,
		community.CreateElectionPermission,
		true)
	if err != nil {
		return err
	}
	return nil
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
			return fmt.Errorf("%w: %s", ErrBadCensusAddressees, c.Name)
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnknownCensusType, c.CensusType)
	}
	// create community in the database
	if err := l.db.AddCommunity(c.ID, c.Name, c.ImageURL, c.GroupChatURL,
		dbCensus, c.Channels, c.Admins, c.Notifications,
	); err != nil {
		return errors.Join(ErrAddCommunity, err)
	}
	return nil
}

// authTransactOpts helper method creates the transact options with the private
// key configured in the CommunityHub. It sets the nonce, gas price, and gas
// limit. If something goes wrong creating the signer, getting the nonce, or
// getting the gas price, it returns an error.
func (l *CommunityHub) authTransactOpts() (*bind.TransactOpts, error) {
	if l.privKey == nil {
		return nil, ErrNoPrivKeyConfigured
	}
	bChainID := new(big.Int).SetUint64(l.ChainID)
	auth, err := bind.NewKeyedTransactorWithChainID(l.privKey, bChainID)
	if err != nil {
		return nil, errors.Join(ErrCreatingSigner, err)
	}
	// create the context with a timeout
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel()
	// set the nonce
	nonce, err := l.w3cli.PendingNonceAt(ctx, l.privAddress)
	if err != nil {
		return nil, errors.Join(ErrSendingTx, err)
	}
	auth.Nonce = new(big.Int).SetUint64(nonce)
	// set the gas price
	if auth.GasPrice, err = l.w3cli.SuggestGasPrice(ctx); err != nil {
		return nil, errors.Join(ErrSendingTx, err)
	}
	// set the gas limit
	auth.GasLimit = 0
	return auth, nil
}

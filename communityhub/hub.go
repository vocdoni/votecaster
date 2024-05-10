package communityhub

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	c3web3 "github.com/vocdoni/census3/helpers/web3"
	comhub "github.com/vocdoni/vote-frame/communityhub/contracts/communityhubtoken"
	dbmongo "github.com/vocdoni/vote-frame/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.vocdoni.io/dvote/log"
)

// DefaultScannerCooldown is the default time that the scanner sleeps between
// scan iterations
const DefaultScannerCooldown = time.Second * 5

// CommunityHubConfig struct defines the configuration for the CommunityHub.
// It includes the contract address, the chain ID where the contract is
// deployed, a database instance, and the scanner cooldown (by default 10s
// (DefaultScannerCooldown)).
type CommunityHubConfig struct {
	ContractAddress  common.Address
	ChainID          uint64
	DB               *dbmongo.MongoStorage
	PrivKey          string
	DiscoverCooldown time.Duration
	SyncCooldown     time.Duration
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
	nextCommunityID  atomic.Uint64
	w3cli            *c3web3.Client
	contract         *comhub.CommunityHubToken
	discoverCooldown time.Duration
	syncCooldown     time.Duration
	privKey          *ecdsa.PrivateKey
	privAddress      common.Address

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
	// set the scanner cooldowns from the configuration if they are defined, or
	// use the default one
	if community.discoverCooldown = DefaultScannerCooldown; conf.DiscoverCooldown > 0 {
		community.discoverCooldown = conf.DiscoverCooldown
	}
	if community.syncCooldown = DefaultScannerCooldown * 2; conf.SyncCooldown > 0 {
		community.syncCooldown = conf.SyncCooldown
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
				time.Sleep(l.discoverCooldown)
				// get the next community from the contract
				currentID := l.nextCommunityID.Load()
				newCommunity, err := l.Community(currentID)
				if err != nil {
					// if community is disabled, skip it incrementing the
					// next community ID, otherwise log the error unless the
					// community is not found
					switch err {
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
					"censusAddresses", newCommunity.CensusAddesses,
					"disabled", newCommunity.Disabled)
				if err := l.addCommunity(newCommunity); err != nil {
					// return if database is closed
					if err == ErrClosedDB {
						return
					}
					log.Warnw("failed to add community in database",
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

// SyncCommunities method starts the listener to sync the communities in the
// database with the contract. It gets the community data from the contract
// and updates it in the database. It iterates from the first community (id: 1)
// to the last one (next - 1) in the contract updating the community data in
// the database. If something goes wrong it logs an error and continues with
// the next iteration. It sleeps between iterations the sync cooldown time.
func (l *CommunityHub) SyncCommunities() {
	l.waiter.Add(1)
	go func() {
		defer l.waiter.Done()
		for {
			select {
			case <-l.ctx.Done():
				return
			default:
				time.Sleep(l.syncCooldown)
				// iterate from the first community (id: 1) to the last one
				// (next - 1) in  the contract updating the community data in
				// the database
				lastID := l.nextCommunityID.Load() - 1
				log.Infow("syncing communities", "communities", lastID)
				for id := uint64(1); id <= lastID; id++ {
					// get the community data from the contract
					newCommunity, err := l.Community(id)
					if err != nil {
						if err != ErrCommunityNotFound {
							log.Warnw("failed to get community from contract",
								"communityID", id,
								"error", err)
						}
						continue
					}
					// update the community data in the database
					if err := l.updateCommunity(newCommunity); err != nil {
						// return if database is closed
						if err == ErrClosedDB {
							return
						}
						log.Warnw("failed to update community in database",
							"communityID", id,
							"error", err)
						continue
					}
				}
				log.Info("communities synced")
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
	// check if the community ID provided exists in the contract
	bNextID, err := l.contract.GetNextCommunityId(nil)
	if err != nil {
		return nil, errors.Join(ErrGettingCommunity, err)
	}
	if communityID >= bNextID.Uint64() {
		return nil, ErrCommunityNotFound
	}
	// get the community data from the contract
	cc, err := l.contract.GetCommunity(nil, new(big.Int).SetUint64(communityID))
	if err != nil {
		return nil, errors.Join(ErrGettingCommunity, err)
	}
	// return the community data
	return ContractToHub(communityID, cc)
}

// SetCommunity method sets the community data provided in the contract. It
// gets the current community data from the contract and updates it with the
// new data checking every field and updating it if it is different from the
// current one. It ensures that the creator of the community (the first admin)
// remains as an admin after the update. If something goes wrong, it returns an
// error.
func (l *CommunityHub) SetCommunity(communityID uint64, newData *HubCommunity) (*HubCommunity, error) {
	if l.privKey == nil {
		return nil, ErrNoPrivKeyConfigured
	}
	// get the current community data from the contract
	community, err := l.Community(communityID)
	if err != nil {
		return nil, err
	}
	// update the community data with the new data checking every field and
	// updating it if it is different from the current one

	// if a new name is provided and it is different from the current one, set
	// the new name
	if newData.Name != "" && newData.Name != community.Name {
		community.Name = newData.Name
	}
	// if a new image URL is provided and it is different from the current one,
	// set the new image URL
	if newData.ImageURL != "" && newData.ImageURL != community.ImageURL {
		community.ImageURL = newData.ImageURL
	}
	// if a new census type is provided and it is different from the current one,
	// set the new census type
	if newData.CensusType != "" && newData.CensusType != community.CensusType {
		community.CensusType = newData.CensusType
	}
	// if a new census channel is provided and it is different from the current
	// one, set the new census channel
	if newData.CensusChannel != "" && newData.CensusChannel != community.CensusChannel {
		community.CensusChannel = newData.CensusChannel
	}
	// if a new census addresses list is provided with at least one address update
	// the census addresses list
	if newData.CensusAddesses != nil && len(newData.CensusAddesses) > 0 {
		community.CensusAddesses = newData.CensusAddesses
	}
	// if a new admins list is provided with at least one admin update the
	// admins list only if the first admin continues being the creator
	if len(newData.Admins) > 0 {
		if newData.Admins[0] != community.Admins[0] {
			return nil, ErrNoAdminCreator
		}
		community.Admins = newData.Admins
	}
	// overwrite the group chat URL with the new group chat URL
	community.GroupChatURL = newData.GroupChatURL
	// overwrite the channels list with the new channels list
	community.Channels = newData.Channels
	// if a new notifications value is provided set the new notifications value
	if newData.Notifications != nil {
		community.Notifications = newData.Notifications
	}
	// if a new disabled value is provided set the new disabled value
	if newData.Disabled != nil {
		community.Disabled = newData.Disabled
	}
	// set the community data in the contract
	cc, err := HubToContract(community)
	if err != nil {
		return nil, err
	}
	// convert the community ID to a *big.Int
	bCommunityID := new(big.Int).SetUint64(communityID)
	// get auth opts and set the community data in the contract
	transactOpts, err := l.authTransactOpts()
	if err != nil {
		return nil, err
	}
	if _, err := l.contract.AdminManageCommunity(transactOpts, bCommunityID, cc.Metadata,
		cc.Census, cc.Guardians, cc.CreateElectionPermission, cc.Disabled); err != nil {
		return nil, errors.Join(ErrSettingCommunity, err)
	}
	return community, nil
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
	if l.privKey == nil {
		return ErrNoPrivKeyConfigured
	}
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
		comhub.IResultResult{
			Question:         results.Question,
			Options:          results.Options,
			Date:             results.Date,
			Tally:            results.Tally,
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

// addCommunity helper method creates a new community in the database. It uses
// the HubToDB helper method to convert the HubCommunity struct to a dbmongo
// Community struct. If something goes wrong creating the community, it returns
// an error.
func (l *CommunityHub) addCommunity(hcommunity *HubCommunity) error {
	dbCommunity, err := HubToDB(hcommunity)
	if err != nil {
		return err
	}
	// create community in the database including the first admin as the creator
	if err := l.db.AddCommunity(dbCommunity.ID, dbCommunity.Name, dbCommunity.ImageURL,
		dbCommunity.GroupChatURL, dbCommunity.Census, dbCommunity.Channels,
		dbCommunity.Admins[0], dbCommunity.Admins, dbCommunity.Notifications,
		dbCommunity.Disabled,
	); err != nil {
		if err == mongo.ErrClientDisconnected {
			return ErrClosedDB
		}
		return errors.Join(ErrAddCommunity, err)
	}
	return nil
}

// updateCommunity helper method updates a community in the database. It uses
// the HubToDB helper method to convert the HubCommunity struct to a dbmongo
// Community struct. If something goes wrong creating the community, it returns
// an error.
func (l *CommunityHub) updateCommunity(hcommunity *HubCommunity) error {
	dbCommunity, err := HubToDB(hcommunity)
	if err != nil {
		return err
	}
	// create community in the database including the first admin as the creator
	if err := l.db.UpdateCommunity(dbCommunity); err != nil {
		if err == mongo.ErrClientDisconnected {
			return ErrClosedDB
		}
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
	// set the gas tip cap
	if auth.GasTipCap, err = l.w3cli.SuggestGasTipCap(ctx); err != nil {
		return nil, errors.Join(ErrSendingTx, err)
	}
	// set the gas limit
	auth.GasLimit = 10000000
	return auth, nil
}

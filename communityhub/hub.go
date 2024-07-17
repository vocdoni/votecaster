package communityhub

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	c3web3 "github.com/vocdoni/census3/helpers/web3"
	dbmongo "github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/log"
)

// DefaultScannerCooldown is the default time that the scanner sleeps between
// scan iterations
const DefaultScannerCooldown = time.Second * 20

// CommunityHubConfig struct defines the configuration for the CommunityHub.
// It includes the contract address, the chain ID where the contract is
// deployed, a database instance, and the scanner cooldown (by default 10s
// (DefaultScannerCooldown)).
type CommunityHubConfig struct {
	ChainAliases      map[string]uint64
	ContractAddresses map[string]common.Address
	DB                *dbmongo.MongoStorage
	PrivKey           string
	DiscoverCooldown  time.Duration
	SyncCooldown      time.Duration
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
	w3pool           *c3web3.Web3Pool
	discoverCooldown time.Duration
	syncCooldown     time.Duration

	ChainAliases map[string]uint64
	contracts    map[string]*HubContract
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
	// check that the contracts addresses and chain ids are defined in the
	// provided configuration
	if len(conf.ContractAddresses) == 0 {
		return nil, ErrMissingContracts
	}
	// check if the chain aliases are defined in the provided configuration
	if len(conf.ChainAliases) == 0 {
		return nil, ErrMissingContracts
	}
	// initialize the context and the listener
	ctx, cancel := context.WithCancel(goblalCtx)
	communityHub := &CommunityHub{
		db:           conf.DB,
		ctx:          ctx,
		cancel:       cancel,
		waiter:       sync.WaitGroup{},
		w3pool:       w3p,
		ChainAliases: conf.ChainAliases,
		contracts:    map[string]*HubContract{},
	}
	// set the scanner cooldowns from the configuration if they are defined, or
	// use the default one
	if communityHub.discoverCooldown = DefaultScannerCooldown; conf.DiscoverCooldown > 0 {
		communityHub.discoverCooldown = conf.DiscoverCooldown
	}
	if communityHub.syncCooldown = DefaultScannerCooldown * 2; conf.SyncCooldown > 0 {
		communityHub.syncCooldown = conf.SyncCooldown
	}
	// initialize contracts
	failed := 0
	for chainAlias, addr := range conf.ContractAddresses {
		chainID, ok := communityHub.ChainIDFromAlias(chainAlias)
		if !ok {
			log.Warnw("failed to get chain ID from alias", "alias", chainAlias)
			failed++
			continue
		}
		// load contract and add it to the contracts map
		contract, err := LoadContract(chainID, chainAlias, addr, communityHub.w3pool, conf.PrivKey)
		if err != nil {
			log.Warnw("failed to load contract", "error", err)
			failed++
			continue
		}
		communityHub.contracts[chainAlias] = contract
	}
	if failed == len(conf.ContractAddresses) {
		return nil, ErrNoValidContracts
	}
	return communityHub, nil
}

// ScanNewCommunities method starts the listener to scan for new communities
// in the contract and create them in the database in background. It gets
// the next community from the contract and, if it exists, it is created it
// in the database. If something goes wrong getting the community data or
// creating the community in the database, it logs an error and retries in
// the next iteration. It sleeps if no communities are found in the contract.
func (l *CommunityHub) ScanNewCommunities() {
	// scan for new logs in background
	l.waiter.Add(1)
	go func() {
		defer l.waiter.Done()
		for {
			for _, contract := range l.contracts {
				select {
				case <-l.ctx.Done():
					return
				default:
					nextID, err := contract.NextContractID()
					if err != nil {
						log.Warnw("failed to get next community ID", "error", err)
						continue
					}
					// get the community from the contract
					communityID, ok := l.CommunityIDByChainAlias(nextID, contract.ChainAlias)
					if !ok {
						log.Warnw("failed to get community ID by chain alias", "chainAlias", contract.ChainAlias, "ID", nextID)
						continue
					}
					onchainCommunity, err := contract.Community(communityID)
					if err != nil {
						if err != ErrCommunityNotFound {
							log.Warnw("failed to get community data", "error", err)
						}
						continue
					}
					if err := l.validateData(onchainCommunity); err != nil {
						log.Warnw("failed to validate community data", "error", err)
						continue
					}
					// store the community in the database
					if err := l.addCommunityToDB(onchainCommunity); err != nil {
						log.Warnw("failed to add community to database", "error", err)
						continue
					}
				}
			}
			time.Sleep(l.discoverCooldown)
		}
	}()
}

// SyncCommunities method starts the listener to sync the communities in the
// database with the contract. It gets the community data from the contract
// and updates it in the database. It iterates from the first community (id: 1)
// to the last one (next - 1) in the contract updating the community data in
// the database. If something goes wrong it logs an error and continues with
// the next iteration. It sleeps between iterations the sync cooldown time.
func (ch *CommunityHub) SyncCommunities() {
	ch.waiter.Add(1)
	go func() {
		defer ch.waiter.Done()
		for {
			// iterate over all the community contracts and sync them,
			// getting the info of the communities stored in the database
			// from the contract and updating them in the database
			for _, contract := range ch.contracts {
				log.Infow("syncing communities", "chainAlias", contract.ChainAlias, "contract", contract.Address.String())
				nextID, err := contract.NextContractID()
				if err != nil {
					log.Warnw("failed to get next community ID", "error", err)
					continue
				}
				// iterate from 1 to the last inserted ID in the database
				// getting community data from the contract and updating it
				// in the database
				for id := uint64(1); id < nextID; id++ {
					select {
					case <-ch.ctx.Done():
						return
					default:
						// get the community from the contract
						communityID, ok := ch.CommunityIDByChainAlias(id, contract.ChainAlias)
						if !ok {
							log.Warnw("failed to get community ID by chain alias", "chainAlias", contract.ChainAlias)
							continue
						}
						onchainCommunity, err := contract.Community(communityID)
						if err != nil {
							log.Warnw("failed to get community data", "error", err, "communityID", communityID)
							continue
						}
						// get the community from the database
						dbCommunity, err := ch.communityFromDB(communityID)
						if err != nil {
							if errors.Is(err, ErrClosedDB) {
								return
							}
							if !errors.Is(err, ErrCommunityNotFound) {
								log.Warnw("failed to get community from database", "error", err)
							}
							if err := ch.addCommunityToDB(onchainCommunity); err != nil {
								log.Warnw("failed to add community to database", "error", err)
							}
							continue
						}
						// join the community data from the contract with the
						// community data from the database
						community, err := ch.joinCommunityData(dbCommunity, onchainCommunity)
						if err != nil {
							log.Warnw("failed to join community data", "error", err)
							continue
						}
						// update the community in the database
						if err := ch.updateCommunityToDB(community); err != nil {
							log.Warnw("failed to update community in database", "error", err)
							continue
						}
					}
				}
			}
			time.Sleep(ch.syncCooldown)
		}
	}()
}

// Stop method stops the listener and waits for the goroutines to finish.
func (l *CommunityHub) Stop() {
	log.Info("stopping communities hub scanner")
	l.cancel()
	l.waiter.Wait()
}

// CommunityContract method gets the contract of a community by the community ID.
// It decodes the chain ID from the community ID and gets the contract from the
// contracts map. If the contract is not found, it returns an error.
func (ch *CommunityHub) CommunityContract(communityID string) (*HubContract, error) {
	chainAlias, _, ok := DecodePrefix(communityID)
	if !ok {
		return nil, ErrDecodeCommunityID
	}
	contract, ok := ch.contracts[chainAlias]
	if !ok {
		return nil, ErrContractNotFound
	}
	return contract, nil
}

// UpdateCommunity method updates a community in the contract and the database.
// It merges the new data with the current data of the community and updates it
// in the contract and the database. If something goes wrong updating the
// community in the contract or the database, it returns an error.
func (ch *CommunityHub) UpdateCommunity(newData *HubCommunity) error {
	chainAlias, _, ok := ch.ChainAliasAndContractIDFromCommunityID(newData.CommunityID)
	if !ok {
		return ErrDecodeCommunityID
	}
	contract, ok := ch.contracts[chainAlias]
	if !ok {
		return ErrContractNotFound
	}
	currentData, err := ch.communityFromDB(newData.CommunityID)
	if err != nil {
		return err
	}
	resultData, err := ch.joinCommunityData(currentData, newData)
	if err != nil {
		return err
	}
	if err := contract.SetCommunity(resultData); err != nil {
		return errors.Join(ErrSettingCommunity, err)
	}
	return ch.updateCommunityToDB(resultData)
}

// CommunityIDByChainID method gets the community ID by the chain ID and the
// ID of the community. It gets the chain alias from the chain ID and creates
// the community ID using the chain alias and the ID. If the chain alias is not
// found, it returns an empty string and false.
func (ch *CommunityHub) CommunityIDByChainID(id, chainID uint64) (string, bool) {
	chainAlias, ok := ch.ChainAliasFromID(chainID)
	if !ok {
		return "", false
	}
	return fmt.Sprintf(chainPrefixFormat, chainAlias, fmt.Sprint(id)), true
}

// CommunityIDByChainAlias method gets the community ID by the chain ID and the
// ID of the community. It gets the chain alias from the chain ID and creates
// the community ID using the chain alias and the ID. If the chain alias is not
// found, it returns an empty string and false.
func (h *CommunityHub) CommunityIDByChainAlias(id uint64, chainAlias string) (string, bool) {
	if _, ok := h.ChainIDFromAlias(chainAlias); !ok {
		return "", false
	}
	return fmt.Sprintf(chainPrefixFormat, chainAlias, fmt.Sprint(id)), true
}

// ChainAliasAndContractIDFromCommunityID method gets the chain alias and the
// ID of the community by the community ID. It decodes the chain alias and the
// ID from the community ID. If the community ID is not valid, it returns an
// empty string, 0, and false.
func (hc *CommunityHub) ChainAliasAndContractIDFromCommunityID(communityID string) (string, uint64, bool) {
	chainAlias, strID, ok := DecodePrefix(communityID)
	if !ok {
		return "", 0, false
	}
	if _, ok := hc.ChainAliases[chainAlias]; !ok {
		return "", 0, false
	}
	id, err := strconv.ParseUint(strID, 10, 64)
	if err != nil {
		return "", 0, false
	}
	return chainAlias, id, true
}

// ChainIDAndIDFromCommunityID method gets the chain ID and the ID of the
// community by the community ID. It decodes the chain alias and the ID from
// the community ID and gets the chain ID from the chain alias. If the chain
// alias is not found, it returns 0, 0, and false.
func (h *CommunityHub) ChainIDAndIDFromCommunityID(communityID string) (uint64, uint64, bool) {
	chainAlias, id, ok := h.ChainAliasAndContractIDFromCommunityID(communityID)
	if !ok {
		return 0, 0, false
	}
	chainID, ok := h.ChainIDFromAlias(chainAlias)
	if !ok {
		return 0, 0, false
	}
	return chainID, id, true
}

// ChainIDFromAlias method gets the chain ID by the chain alias. It iterates
// over the chain aliases map and returns the chain ID if the chain alias is
// found. If the chain alias is not found, it returns 0 and false.
func (ch *CommunityHub) ChainIDFromAlias(alias string) (uint64, bool) {
	chainID, ok := ch.ChainAliases[alias]
	return chainID, ok
}

// ChainAliasFromID method returns the chain alias by the chain ID. It iterates
// over the chain aliases map and returns the chain alias if the chain ID is
// found.
func (ch *CommunityHub) ChainAliasFromID(chainID uint64) (string, bool) {
	for alias, id := range ch.ChainAliases {
		if id == chainID {
			return alias, true
		}
	}
	return "", false
}

// validateData method validates the data of a community. It checks that the
// chain ID, the ID, the community ID, the name, the census type, the channel,
// the addresses, the admins, the notifications, and the disabled fields are
// valid. If something is wrong, it returns an error.
func (ch *CommunityHub) validateData(data *HubCommunity) error {
	if data.ChainID == 0 {
		return fmt.Errorf("%w: no chain id", ErrInvalidCommunityData)
	}
	if data.Name == "" {
		return fmt.Errorf("%w: invalid community name", ErrInvalidCommunityData)
	}
	switch data.CensusType {
	case CensusTypeChannel, CensusTypeFollowers:
		if data.CensusChannel == "" {
			return fmt.Errorf("%w: invalid channel", ErrInvalidCommunityData)
		}
	case CensusTypeERC20, CensusTypeNFT:
		if len(data.CensusAddesses) == 0 {
			return fmt.Errorf("%w: invalid addresses", ErrInvalidCommunityData)
		}
	default:
		return fmt.Errorf("%w: unknown census type", ErrInvalidCommunityData)
	}
	if len(data.Admins) == 0 {
		return fmt.Errorf("%w: no admins", ErrInvalidCommunityData)
	}
	if data.Notifications == nil {
		data.Notifications = new(bool)
	}
	if data.Disabled == nil {
		enabled := true
		data.Disabled = &enabled
	}
	return nil
}

// joinCommunityData method joins the data of a community. If no old data is
// provided, it returns the new data if it is valid. If old data is provided,
// it validates it and updates the fields that are different and valid in the
// new data. It returns the updated data or an error if the creator is not an
// admin.
func (ch *CommunityHub) joinCommunityData(data, newData *HubCommunity) (*HubCommunity, error) {
	// if no old data is provided, return the new data if it is valid
	if data == nil {
		if err := ch.validateData(newData); err != nil {
			return nil, err
		}
		return newData, nil
	}
	// if old data is provided, validate it
	if err := ch.validateData(data); err != nil {
		return nil, err
	}
	if data.CommunityID != newData.CommunityID {
		return nil, ErrCommunityIDMismatch
	}
	// if old data is provided and valid, update the fields that are different
	// and valid if the new data
	if newData.Name != "" && data.Name != newData.Name {
		data.Name = newData.Name
	}
	if newData.ImageURL != "" && data.ImageURL != newData.ImageURL {
		data.ImageURL = newData.ImageURL
	}
	if newData.CensusType != "" && data.CensusType != newData.CensusType {
		data.CensusType = newData.CensusType
	}
	switch data.CensusType {
	case CensusTypeChannel, CensusTypeFollowers:
		if newData.CensusChannel != "" && data.CensusChannel != newData.CensusChannel {
			data.CensusChannel = newData.CensusChannel
		}
	case CensusTypeERC20, CensusTypeNFT:
		if len(newData.CensusAddesses) > 0 {
			data.CensusAddesses = newData.CensusAddesses
		}
	}
	if len(newData.Admins) > 0 {
		// check if the creator is still an admin
		if newData.Admins[0] != data.Admins[0] {
			log.Warnw("creator is not an admin", "admins", data.Admins, "newAdmins", newData.Admins)
			return nil, ErrNoAdminCreator
		}
		data.Admins = newData.Admins
	}
	if newData.Notifications != nil {
		data.Notifications = newData.Notifications
	}
	if newData.Disabled != nil {
		data.Disabled = newData.Disabled
	}
	data.GroupChatURL = newData.GroupChatURL
	data.Channels = newData.Channels
	return data, nil
}

// communityFromDB helper method gets a community from the database by the
// community ID. It decodes the chain ID and the ID from the community ID and
// gets the community from the database. If the community is not found, it
// returns an error.
func (ch *CommunityHub) communityFromDB(communityID string) (*HubCommunity, error) {
	community, err := ch.db.Community(communityID)
	if err != nil {
		if dbmongo.IsDBClosed(err) {
			return nil, ErrClosedDB
		}
		return nil, errors.Join(ErrGettingCommunity, err)
	}
	if community == nil {
		return nil, ErrCommunityNotFound
	}
	chainID, id, ok := ch.ChainIDAndIDFromCommunityID(communityID)
	if !ok {
		return nil, ErrDecodeCommunityID
	}
	return DBToHub(community, id, chainID)
}

// addCommunity helper method creates a new community in the database. It uses
// the HubToDB helper method to convert the HubCommunity struct to a dbmongo
// Community struct. If something goes wrong creating the community, it returns
// an error.
func (l *CommunityHub) addCommunityToDB(hcommunity *HubCommunity) error {
	// if community already exists in the database, update it
	current, err := l.db.Community(hcommunity.CommunityID)
	if err != nil {
		if dbmongo.IsDBClosed(err) {
			return ErrClosedDB
		}
		return errors.Join(ErrAddCommunity, err)
	}
	if current != nil {
		return l.updateCommunityToDB(hcommunity)
	}
	// if community does not exist in the database, create it in the database
	dbc, err := HubToDB(hcommunity)
	if err != nil {
		return err
	}
	// create community in the database including the first admin as the creator
	if err := l.db.AddCommunity(dbc); err != nil {
		if dbmongo.IsDBClosed(err) {
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
func (l *CommunityHub) updateCommunityToDB(hcommunity *HubCommunity) error {
	dbCommunity, err := HubToDB(hcommunity)
	if err != nil {
		return err
	}
	// create community in the database including the first admin as the creator
	if err := l.db.UpdateCommunity(dbCommunity); err != nil {
		if dbmongo.IsDBClosed(err) {
			return ErrClosedDB
		}
		return errors.Join(ErrAddCommunity, err)
	}
	return nil
}

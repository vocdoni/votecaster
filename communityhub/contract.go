package communityhub

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	c3web3 "github.com/vocdoni/census3/helpers/web3"
	comhub "github.com/vocdoni/vote-frame/communityhub/contracts/communityhubtoken"
	"go.vocdoni.io/dvote/log"
)

// HubContract struct represents the CommunityHub contract with in a specific
// chain. It contains the chain ID, the contract address, the web3 client, the
// contract, the private key, and the private address. It provides a set of
// methods to interact with the contract, such as getting the next community ID,
// getting and setting the community data, getting and setting the election
// results of a community.
type HubContract struct {
	ChainID     uint64
	ChainAlias  string
	Address     common.Address
	privKey     *ecdsa.PrivateKey
	privAddress common.Address

	w3cli    *c3web3.Client
	contract *comhub.CommunityHubToken
}

// LoadContract method initializes the CommunityHub struct with the chain ID,
// contract address, web3 pool, and private key provided. If something goes
// wrong initializing the web3 client, the contract, or the private key, it
// returns an error.
func LoadContract(chainID uint64, chainAlias string, addr common.Address, w3p *c3web3.Web3Pool, pk string) (*HubContract, error) {
	hc := &HubContract{
		ChainID:    chainID,
		Address:    addr,
		ChainAlias: chainAlias,
	}
	// initialize the web3 client for the chain
	var err error
	hc.w3cli, err = w3p.Client(chainID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWeb3Client, err)
	}
	// initialize the contract with the web3 client and the contract addr
	hc.contract, err = comhub.NewCommunityHubToken(addr, hc.w3cli)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInitContract, err)
	}
	// loading the private key
	if err := hc.initiPrivateKey(pk); err != nil {
		log.Warn(err)
	}
	// get the next community ID
	return hc, nil
}

// NextID method gets the next community ID from the contract and returns it as
// a uint64. If something goes wrong getting the next community ID from the
// contract, it returns an error.
func (hc *HubContract) NextContractID() (uint64, error) {
	nextID, err := hc.contract.GetNextCommunityId(nil)
	if err != nil {
		return 0, err
	}
	iNextID := nextID.Uint64()
	if iNextID == 0 {
		return 1, nil
	}
	return iNextID, nil
}

// Community method gets the community data using the community ID from the
// contract and returns it as a HubCommunity struct. If something goes wrong
// getting the community data from the contract, it returns an error.
func (hc *HubContract) Community(communityID string) (*HubCommunity, error) {
	if hc.contract == nil {
		return nil, ErrInitContract
	}
	chainAlias, strID, ok := DecodePrefix(communityID)
	if !ok || chainAlias != hc.ChainAlias {
		return nil, ErrDecodeCommunityID
	}
	id, err := strconv.ParseUint(strID, 10, 64)
	if err != nil {
		return nil, err
	}
	nextID, err := hc.NextContractID()
	if err != nil {
		return nil, err
	}
	if id > nextID {
		return nil, ErrCommunityNotFound
	}
	// convert the community ID to a *big.Int
	bCommunityID := new(big.Int).SetUint64(id)
	// get the community data from the contract
	cc, err := hc.contract.GetCommunity(nil, bCommunityID)
	if err != nil {
		return nil, errors.Join(ErrGettingCommunity, err)
	}
	if cc.Metadata.Name == "" && big.NewInt(0).Cmp(cc.Funds) == 0 {
		return nil, ErrCommunityNotFound
	}
	// convert the contract community to a HubCommunity
	community, err := ContractToHub(id, hc.ChainID, communityID, cc)
	if err != nil {
		return nil, err
	}
	return community, nil
}

// SetCommunity method sets the community data provided in the contract. If
// something goes wrong setting the community data in the contract, it returns
// an error.
func (hc *HubContract) SetCommunity(community *HubCommunity) error {
	if hc.contract == nil {
		return ErrInitContract
	}
	if hc.privKey == nil {
		return ErrNoPrivKeyConfigured
	}
	// set the community data in the contract
	cc, err := HubToContract(community)
	if err != nil {
		return err
	}
	// convert the community ID to a *big.Int
	bCommunityID := new(big.Int).SetUint64(community.ContractID)
	// get auth opts and set the community data in the contract
	transactOpts, err := hc.authTransactOpts()
	if err != nil {
		return err
	}
	if _, err := hc.contract.AdminManageCommunity(transactOpts, bCommunityID, cc.Metadata,
		cc.Census, cc.Guardians, cc.CreateElectionPermission, cc.Disabled); err != nil {
		return errors.Join(ErrSettingCommunity, err)
	}
	return nil
}

// Results method gets the election results using the community and elections
// IDs from the contract and returns them as a HubResults struct. If something
// goes wrong getting the results from the contract, it returns an error.
func (hc *HubContract) Results(communityID uint64, electionID []byte) (*HubResults, error) {
	if hc.contract == nil {
		return nil, ErrInitContract
	}
	// convert the community ID to a *big.Int
	bCommunityID := new(big.Int).SetUint64(communityID)
	// convert the election ID to a [32]byte
	bElectionID := [32]byte{}
	copy(bElectionID[:], electionID)
	// get the election results from the contract
	contractResults, err := hc.contract.GetResult(nil, bCommunityID, bElectionID)
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
func (hc *HubContract) SetResults(community *HubCommunity, results *HubResults) error {
	if hc.contract == nil {
		return ErrInitContract
	}
	if hc.privKey == nil {
		return ErrNoPrivKeyConfigured
	}
	transactOpts, err := hc.authTransactOpts()
	if err != nil {
		return err
	}
	// convert the community ID to a *big.Int
	bCommunityID := new(big.Int).SetUint64(community.ContractID)
	// convert the election ID to a [32]byte
	bElectionID := [32]byte{}
	copy(bElectionID[:], results.ElectionID)
	// convert census root to a [32]byte
	bCensusRoot := [32]byte{}
	copy(bCensusRoot[:], results.CensusRoot)
	// set the election results in the contract
	if _, err := hc.contract.SetResult(transactOpts, bCommunityID, bElectionID,
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

// initiPrivateKey helper method initializes the private key in the CommunityHub
// struct. If the private key is not defined, it returns nil. If the private key
// is defined, it parses it and sets the private key and the private address in
// the struct. If something goes wrong parsing the private key, it returns an
// error.
func (hc *HubContract) initiPrivateKey(privKey string) error {
	// parse the private key if it is defined
	if privKey == "" {
		return nil
	}
	var err error
	hc.privKey, err = crypto.HexToECDSA(privKey)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInitializingPrivateKey, err)
	}
	hc.privAddress = crypto.PubkeyToAddress(hc.privKey.PublicKey)
	return nil
}

// authTransactOpts helper method creates the transact options with the private
// key configured in the CommunityHub. It sets the nonce, gas price, and gas
// limit. If something goes wrong creating the signer, getting the nonce, or
// getting the gas price, it returns an error.
func (hc *HubContract) authTransactOpts() (*bind.TransactOpts, error) {
	if hc.privKey == nil {
		return nil, ErrNoPrivKeyConfigured
	}
	if hc.contract == nil {
		return nil, ErrInitContract
	}
	bChainID := new(big.Int).SetUint64(hc.ChainID)
	auth, err := bind.NewKeyedTransactorWithChainID(hc.privKey, bChainID)
	if err != nil {
		return nil, errors.Join(ErrCreatingSigner, err)
	}
	// create the context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// set the nonce
	nonce, err := hc.w3cli.PendingNonceAt(ctx, hc.privAddress)
	if err != nil {
		return nil, errors.Join(ErrSendingTx, err)
	}
	auth.Nonce = new(big.Int).SetUint64(nonce)
	// set the gas tip cap
	if auth.GasTipCap, err = hc.w3cli.SuggestGasTipCap(ctx); err != nil {
		return nil, errors.Join(ErrSendingTx, err)
	}
	// set the gas limit
	auth.GasLimit = 10000000
	return auth, nil
}

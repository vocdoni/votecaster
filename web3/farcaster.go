package web3

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	fckr "github.com/vocdoni/vote-frame/web3/contracts"
)

const (
	KeyRegistryAddress = "0x00000000Fc1237824fb747aBDE0FF18990E59b7e"
)

type FarcasterProviderConf struct {
	Web3Endpoint string
	DB           *DB
}

type FarcasterProvider struct {
	// web3
	endpoint string
	client   *ethclient.Client
	contract fckr.FarcasterKeyRegistry
	// db
	db *DB
}

func (p *FarcasterProvider) Init(conf FarcasterProviderConf) error {
	if p.endpoint == "" {
		return errors.New("endpoints not defined")
	}
	p.endpoint = conf.Web3Endpoint
	p.db = conf.DB

	// connect to the endpoint
	client, err := ethclient.Dial(p.endpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to the Ethereum client: %w", err)
	}
	// set the client, parse the addresses and initialize the contracts
	p.client = client
	keyRegistryAddress := common.HexToAddress(KeyRegistryAddress)
	if p.contract, err = fckr.NewFarcasterKeyRegistry(keyRegistryAddress, client); err != nil {
		return fmt.Errorf("failed to instantiate Farcaster KeyRegistry contract: %w", err)
	}
	return nil
}

func (p *FarcasterProvider) GetAppKeysByFid(fid *big.Int) ([][]byte, error) {
	return p.contract.FarcasterKeyRegistryCaller.KeysOf(nil, fid, 1)
}

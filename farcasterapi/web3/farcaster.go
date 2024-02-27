package web3

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	fckr "github.com/vocdoni/vote-frame/farcasterapi/web3/contracts"
)

const (
	KeyRegistryAddress = "0x00000000Fc1237824fb747aBDE0FF18990E59b7e"
)

type FarcasterProvider struct {
	endpoint string
	client   *ethclient.Client
	contract *fckr.FarcasterKeyRegistry
}

func NewFarcasterProvider(web3Endpoint string) (*FarcasterProvider, error) {
	// connect to the endpoint
	client, err := ethclient.Dial(web3Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %w", err)
	}
	// set the client, parse the addresses and initialize the contracts
	keyRegistryAddress := common.HexToAddress(KeyRegistryAddress)
	contract, err := fckr.NewFarcasterKeyRegistry(keyRegistryAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate Farcaster KeyRegistry contract: %w", err)
	}
	return &FarcasterProvider{
		endpoint: web3Endpoint,
		client:   client,
		contract: contract,
	}, nil

}

func (p *FarcasterProvider) GetAppKeysByFid(fid *big.Int) ([][]byte, error) {
	return p.contract.FarcasterKeyRegistryCaller.KeysOf(nil, fid, 1)
}

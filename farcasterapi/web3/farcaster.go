package web3

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vocdoni/census3/helpers/web3"
	fckr "github.com/vocdoni/vote-frame/farcasterapi/web3/contracts"
)

const (
	KeyRegistryAddress        = "0x00000000Fc1237824fb747aBDE0FF18990E59b7e"
	KeyRegistryChainID uint64 = 10
	maxRetries                = 5
)

type ProviderInfo struct {
	endpoint  string
	available bool
}

// FarcasterProvider is a Web3 provider that connects to multiple Ethereum clients.
type FarcasterProvider struct {
	Address common.Address
	ChainID uint64

	contract *fckr.FarcasterKeyRegistry
	w3p      *web3.Web3Pool
}

func NewFarcasterProvider(w3p *web3.Web3Pool) (*FarcasterProvider, error) {
	fp := &FarcasterProvider{
		Address: common.HexToAddress(KeyRegistryAddress),
		ChainID: KeyRegistryChainID,
		w3p:     w3p,
	}

	cli, err := fp.w3p.Client(fp.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get web3 client: %w", err)
	}
	fp.contract, err = fckr.NewFarcasterKeyRegistry(fp.Address, cli)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate Farcaster KeyRegistry contract: %w", err)
	}
	return fp, nil
}

func (p *FarcasterProvider) GetAppKeysByFid(fid *big.Int) ([][]byte, error) {
	keys, err := p.contract.FarcasterKeyRegistryCaller.KeysOf(nil, fid, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}
	return keys, nil
}

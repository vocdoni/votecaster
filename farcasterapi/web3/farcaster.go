package web3

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	fckr "github.com/vocdoni/vote-frame/farcasterapi/web3/contracts"
	"go.vocdoni.io/dvote/log"
)

const (
	KeyRegistryAddress = "0x00000000Fc1237824fb747aBDE0FF18990E59b7e"
	maxRetries         = 5
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
	var keys [][]byte
	var err error
	retryDelay := time.Second // Starting delay, increments with each retry

	for attempt := 0; attempt < maxRetries; attempt++ {
		keys, err = p.contract.FarcasterKeyRegistryCaller.KeysOf(nil, fid, 1)
		if err == nil {
			return keys, nil // Success, return the result
		}

		if strings.Contains(err.Error(), "429 Too Many Requests") {
			log.Warnw("encountered 429 on web3 call, retrying", "attempt", attempt+1, "retryDelay", retryDelay.Seconds())
			time.Sleep(retryDelay)
			retryDelay++ // Increase delay for next retry
			continue
		}

		log.Errorw(err, "failed to get keys from Farcaster KeyRegistry")
		time.Sleep(time.Second)
	}

	return nil, fmt.Errorf("reached maximum retry attempts")
}

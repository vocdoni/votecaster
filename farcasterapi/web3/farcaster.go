package web3

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
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

type clientInfo struct {
	client    *ethclient.Client
	contract  *fckr.FarcasterKeyRegistry
	endpoint  string
	available bool
}

// FarcasterProvider is a Web3 provider that connects to multiple Ethereum clients.
type FarcasterProvider struct {
	clients []*clientInfo
	mu      sync.Mutex // Protects clients slice
	current int
}

func NewFarcasterProvider() *FarcasterProvider {
	return &FarcasterProvider{}
}

func (p *FarcasterProvider) AddEndpoint(web3Endpoint string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	client, err := ethclient.Dial(web3Endpoint)
	if err != nil {
		log.Warnw("web3 endpoint is not available, skip", "endpoint", web3Endpoint, "error", err)
		return nil
	}

	keyRegistryAddress := common.HexToAddress(KeyRegistryAddress)
	contract, err := fckr.NewFarcasterKeyRegistry(keyRegistryAddress, client)
	if err != nil {
		return fmt.Errorf("failed to instantiate Farcaster KeyRegistry contract: %w", err)
	}
	log.Debugw("added web3 endpoint", "endpoint", web3Endpoint, "count", len(p.clients)+1)
	p.clients = append(p.clients, &clientInfo{
		client:    client,
		contract:  contract,
		endpoint:  web3Endpoint,
		available: true,
	})

	return nil
}

func (p *FarcasterProvider) DelEndpoint(web3Endpoint string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, c := range p.clients {
		if c.endpoint == web3Endpoint {
			c.client.Close()
			p.clients = append(p.clients[:i], p.clients[i+1:]...)
			break
		}
	}
}

func (p *FarcasterProvider) getNextAvailableClient() *clientInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < len(p.clients); i++ {
		idx := (p.current + i + 1) % len(p.clients)
		if p.clients[idx].available {
			p.current = idx
			return p.clients[idx]
		}
	}

	// If no clients are available, start over and mark all as available
	for _, c := range p.clients {
		c.available = true
	}
	p.current = 0
	return p.clients[p.current]
}

func (p *FarcasterProvider) markClientAsNotWorking(endpoint string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, c := range p.clients {
		if c.endpoint == endpoint {
			c.available = false
			break
		}
	}
}

func (p *FarcasterProvider) GetAppKeysByFid(fid *big.Int) ([][]byte, error) {
	var keys [][]byte
	var err error
	retryDelay := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		clientInfo := p.getNextAvailableClient()
		if clientInfo == nil {
			return nil, fmt.Errorf("no available Ethereum clients")
		}
		keys, err = clientInfo.contract.FarcasterKeyRegistryCaller.KeysOf(nil, fid, 1)
		if err == nil {
			return keys, nil
		}

		if strings.Contains(err.Error(), "429 Too Many Requests") {
			log.Warnw("encountered 429 on web3 call, retrying", "attempt", attempt+1, "retryDelay", retryDelay.Seconds())
			time.Sleep(retryDelay)
			retryDelay++
			continue
		}

		log.Errorw(fmt.Errorf("failed to get keys from Farcaster KeyRegistry"), fmt.Sprintf("endpoint: %s", clientInfo.endpoint))
		p.markClientAsNotWorking(clientInfo.endpoint)
	}

	return nil, fmt.Errorf("reached maximum retry attempts")
}

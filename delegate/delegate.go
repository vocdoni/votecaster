package web3

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	delegate "github.com/vocdoni/vote-frame/delegate/web3"
	"go.vocdoni.io/dvote/log"
)

const (
	DelegateRegistryAddress = "0x00000000000000447e69651d841bD8D104Bed493"
	maxRetries              = 3
)

type clientInfo struct {
	client    *ethclient.Client
	contract  *delegate.DelegateRegistry
	endpoint  string
	available bool
}

// DelegateRegistryProvider is a Web3 provider that connects to multiple Ethereum clients.
type DelegateRegistryProvider struct {
	clients []*clientInfo
	mu      sync.Mutex // Protects clients slice
	current int
}

func NewDelegateRegistryProvider() *DelegateRegistryProvider {
	return &DelegateRegistryProvider{}
}

func (p *DelegateRegistryProvider) AddEndpoint(web3Endpoint string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	client, err := ethclient.Dial(web3Endpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}

	delegateRegistryAddress := common.HexToAddress(DelegateRegistryAddress)
	contract, err := delegate.NewDelegateRegistry(delegateRegistryAddress, client)
	if err != nil {
		return fmt.Errorf("failed to instantiate delegate registry contract: %w", err)
	}

	p.clients = append(p.clients, &clientInfo{
		client:    client,
		contract:  contract,
		endpoint:  web3Endpoint,
		available: true,
	})

	return nil
}

func (p *DelegateRegistryProvider) DelEndpoint(web3Endpoint string) {
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

func (p *DelegateRegistryProvider) getNextAvailableClient() *clientInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < len(p.clients); i++ {
		idx := (p.current + i) % len(p.clients)
		if p.clients[idx].available {
			p.current = (idx + 1) % len(p.clients) // Prepare next index
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

func (p *DelegateRegistryProvider) markClientAsNotWorking(endpoint string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, c := range p.clients {
		if c.endpoint == endpoint {
			c.available = false
			break
		}
	}
}

// ContractType represents the supported contract types
// ERC20 -> 0, ERC721 -> 1, ERC1155 -> 2
type ContractType int

const (
	ERC20 ContractType = iota
	ERC721
	ERC1155
)

// CheckDelegate returns the delegation amount from one address to another
func (p *DelegateRegistryProvider) CheckDelegate(
	contractAddress, fromAddress, toAddress common.Address, tokenId *big.Int, contractType ContractType,
) (*big.Int, error) {
	retryDelay := time.Second
	dummyRights := [32]byte{}

	for attempt := 0; attempt < maxRetries; attempt++ {
		clientInfo := p.getNextAvailableClient()
		if clientInfo == nil {
			return nil, fmt.Errorf("no available Ethereum clients")
		}

		var err error
		var amount *big.Int
		switch contractType {
		case ERC20:
			// rights: Specific rights to check for, pass the zero value to ignore subdelegations and check full delegations only
			amount, err = clientInfo.contract.CheckDelegateForERC20(nil, toAddress, fromAddress, contractAddress, dummyRights)
			if err == nil {
				return amount, nil
			}
		case ERC1155:
			// rights: Specific rights to check for, pass the zero value to ignore subdelegations and check full delegations only
			amount, err = clientInfo.contract.CheckDelegateForERC1155(nil, toAddress, fromAddress, contractAddress, tokenId, dummyRights)
			if err == nil {
				return amount, nil
			}
		case ERC721:
			// rights: Specific rights to check for, pass the zero value to ignore subdelegations and check full delegations only
			var delegated bool
			delegated, err = clientInfo.contract.CheckDelegateForERC721(nil, toAddress, fromAddress, contractAddress, tokenId, dummyRights)
			if delegated {
				amount = big.NewInt(1)
			} else {
				amount = big.NewInt(0)
			}
			if err == nil {
				return amount, nil
			}
		default:
			return nil, fmt.Errorf("contract type not supported")
		}

		if strings.Contains(err.Error(), "429 Too Many Requests") {
			log.Warnw("encountered 429 on web3 call, retrying", "attempt", attempt+1, "retryDelay", retryDelay.Seconds())
			time.Sleep(retryDelay)
			retryDelay++
			continue
		}

		log.Errorw(err, "failed to get delegations from DelegateRegistry")
		p.markClientAsNotWorking(clientInfo.endpoint)
	}

	return nil, fmt.Errorf("reached maximum retry attempts")
}

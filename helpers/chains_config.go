package helpers

import (
	"encoding/json"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

// ChainConfig represents the configuration of a chain, including the chain ID,
// the chain alias, the chain name, the endpoints, and the address of the
// CommunityHub contract.
type ChainConfig struct {
	ChainID             uint64
	ChainAlias          string
	Name                string
	Endpoints           []string
	CommunityHubAddress string
}

// ChainsConfig is a slice of ChainConfig.
type ChainsConfig []*ChainConfig

// chainsConfigFile is a representation of the chains_file.json file format to
// load the chains configuration.
type chainsConfigFile map[string]struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	RPCs map[string]struct {
		HTTP      []string `json:"http"`
		WebSocket []string `json:"ws"`
	} `json:"rpcUrls"`
	Contracts struct {
		CommunityHub struct {
			Address string `json:"address"`
		} `json:"communityHub"`
	} `json:"contracts"`
}

// LoadChainsConfig loads the chains configuration from the file at the given
// path. It returns a ChainsConfig struct with the configuration loaded from the
// file. If something goes wrong loading the file or parsing the JSON, it returns
// an error.
func LoadChainsConfig(path string) (ChainsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := chainsConfigFile{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	configs := ChainsConfig{}
	for alias, info := range cfg {
		endpoints := []string{}
		for _, rpcs := range info.RPCs {
			endpoints = append(endpoints, rpcs.HTTP...)
		}
		config := &ChainConfig{
			ChainID:    info.ID,
			ChainAlias: alias,
			Name:       info.Name,
			Endpoints:  endpoints,
		}
		if info.Contracts.CommunityHub.Address != "" {
			config.CommunityHubAddress = info.Contracts.CommunityHub.Address
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// ContractsAddressesByChainID returns a map with the chain ID as the key and
// the CommunityHub contract address as the value.
func (c ChainsConfig) ContractsAddressesByChainAlias() map[string]common.Address {
	contracts := map[string]common.Address{}
	for _, config := range c {
		contracts[config.ChainAlias] = common.HexToAddress(config.CommunityHubAddress)
	}
	return contracts
}

// ChainAliasesByChainID returns a map with the chain ID as the key and the
// chain alias as the value.
func (c ChainsConfig) ChainChainIDByAlias() map[string]uint64 {
	aliases := map[string]uint64{}
	for _, config := range c {
		aliases[config.ChainAlias] = config.ChainID
	}
	return aliases
}

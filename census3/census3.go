package census3

import (
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	c3types "github.com/vocdoni/census3/api"
	c3client "github.com/vocdoni/census3/apiclient"
)

const (
	// maxRetries is the maximum number of retries for a request.
	maxRetries = 3
	// cooldown is the time to wait between retries.
	cooldown = time.Second * 2
)

// Client wraps a client for the Census3 API and other revelant data.
type Client struct {
	c3client.HTTPclient
}

// NewClient creates a new client for the Census3 API.
func NewClient(endpoint string, bearerToken string) (*Client, error) {
	bt, err := uuid.Parse(bearerToken)
	if err != nil {
		return nil, fmt.Errorf("invalid bearer token: %v", err)
	}
	addr, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %v", err)
	}
	httpClient, err := c3client.NewHTTPclient(addr, &bt)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP client: %v", err)
	}
	c3 := &Client{
		HTTPclient: *httpClient,
	}
	return c3, nil
}

// SupportedChains returns the information of the Census3 endpoint supported chains.
func (c3 *Client) SupportedChains() ([]c3types.SupportedChain, error) {
	for i := 0; i < maxRetries; i++ {
		info, err := c3.Info()
		if err != nil {
			time.Sleep(cooldown)
			continue
		}
		return info.SupportedChains, nil
	}
	return nil, fmt.Errorf("failed to get info from census3 endpoint")
}

// Tokens returns the list of tokens registered in the Census3 endpoint.
func (c3 *Client) Tokens() ([]*c3types.TokenListItem, error) {
	for i := 0; i < maxRetries; i++ {
		tokens, err := c3.GetTokens(-1, "", "")
		if err != nil {
			time.Sleep(cooldown)
			continue
		}
		return tokens, nil
	}
	return nil, fmt.Errorf("failed to get tokens from census3 endpoint")
}

// Token returns the token with the given ID.
func (c3 *Client) Token(tokenID, externalID string, chainID uint64) (*c3types.Token, error) {
	for i := 0; i < maxRetries; i++ {
		token, err := c3.GetToken(tokenID, chainID, externalID)
		if err != nil {
			time.Sleep(cooldown)
			continue
		}
		return token, nil
	}
	return nil, fmt.Errorf("failed to get token from census3 endpoint")
}

// SupportedTokens returns the list of tokens supported by the Census3 endpoint.
func (c3 *Client) SupportedTokens() ([]string, error) {
	for i := 0; i < maxRetries; i++ {
		tokens, err := c3.GetTokenTypes()
		if err != nil {
			time.Sleep(cooldown)
			continue
		}
		return tokens, nil
	}
	return nil, fmt.Errorf("failed to get supported tokens from census3 endpoint")
}

// Strategies returns the list of strategies registered in the Census3 endpoint.
func (c3 *Client) Strategies() ([]*c3types.Strategy, error) {
	for i := 0; i < maxRetries; i++ {
		strategies, err := c3.GetStrategies(-1, "", "")
		if err != nil {
			time.Sleep(cooldown)
			continue
		}
		return strategies, nil
	}
	return nil, fmt.Errorf("failed to get strategies from census3 endpoint")
}

// Strategy returns the strategy with the given ID.
func (c3 *Client) Strategy(strategyID uint64) (*c3types.Strategy, error) {
	for i := 0; i < maxRetries; i++ {
		strategy, err := c3.GetStrategy(strategyID)
		if err != nil {
			time.Sleep(cooldown)
			continue
		}
		return strategy, nil
	}
	return nil, fmt.Errorf("failed to get strategy from census3 endpoint")
}

// GetStrategyHolders returns the list of holders for the given strategy.
// The returned map has the holder's address as key and the holder's balance as a string encoded big.Int
func (c3 *Client) GetStrategyHolders(strategyID uint64) (map[string]string, error) {
	for i := 0; i < maxRetries; i++ {
		holders, err := c3.GetHoldersByStrategy(strategyID)
		if err != nil {
			time.Sleep(cooldown)
			continue
		}
		return holders.Holders, nil
	}
	return nil, fmt.Errorf("failed to get strategy holders from census3 endpoint")
}

// Census returns the census with the given ID.
func (c3 *Client) Census(censusID uint64) (*c3types.Census, error) {
	for i := 0; i < maxRetries; i++ {
		census, err := c3.GetCensus(censusID)
		if err != nil {
			time.Sleep(cooldown)
			continue
		}
		return census, nil
	}
	return nil, fmt.Errorf("failed to get census from census3 endpoint")
}

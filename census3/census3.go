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

// reqFunc is a function type that encapsulates an API call
type reqFunc[T any] func() (T, error)

// requestWithRetry handles the retry logic for a request
func requestWithRetry[T any](fn reqFunc[T]) (T, error) {
	var result T
	for i := 0; i < maxRetries; i++ {
		result, err := fn()
		if err != nil {
			time.Sleep(cooldown)
			continue
		}
		return result, nil
	}
	return result, fmt.Errorf("failed after %d retries", maxRetries)
}

// SupportedChains returns the information of the Census3 endpoint supported chains.
func (c3 *Client) SupportedChains() ([]c3types.SupportedChain, error) {
	return requestWithRetry(func() ([]c3types.SupportedChain, error) {
		info, err := c3.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to get supported chains: %w", err)
		}
		return info.SupportedChains, nil
	})
}

// Tokens returns the list of tokens registered in the Census3 endpoint.
func (c3 *Client) Tokens() ([]*c3types.TokenListItem, error) {
	return requestWithRetry(func() ([]*c3types.TokenListItem, error) {
		tokens, err := c3.GetTokens(-1, "", "")
		if err != nil {
			return nil, fmt.Errorf("failed to get tokens: %w", err)
		}
		return tokens, nil

	})
}

// Token returns the token with the given ID.
func (c3 *Client) Token(tokenID, externalID string, chainID uint64) (*c3types.Token, error) {
	return requestWithRetry(func() (*c3types.Token, error) {
		token, err := c3.GetToken(tokenID, chainID, externalID)
		if err != nil {
			return nil, fmt.Errorf("failed to get token: %w", err)
		}
		return token, nil
	})
}

// SupportedTokens returns the list of tokens supported by the Census3 endpoint.
func (c3 *Client) SupportedTokens() ([]string, error) {
	return requestWithRetry(func() ([]string, error) {
		tokens, err := c3.GetTokenTypes()
		if err != nil {
			return nil, fmt.Errorf("failed to get supported tokens: %w", err)
		}
		return tokens, nil
	})
}

// Strategies returns the list of strategies registered in the Census3 endpoint.
func (c3 *Client) Strategies() ([]*c3types.Strategy, error) {
	return requestWithRetry(func() ([]*c3types.Strategy, error) {
		strategies, err := c3.GetStrategies(-1, "", "")
		if err != nil {
			return nil, fmt.Errorf("failed to get strategies: %w", err)
		}
		return strategies, nil
	})
}

// Strategy returns the strategy with the given ID.
func (c3 *Client) Strategy(strategyID uint64) (*c3types.Strategy, error) {
	return requestWithRetry(func() (*c3types.Strategy, error) {
		strategy, err := c3.GetStrategy(strategyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get strategy: %w", err)
		}
		return strategy, nil
	})
}

// GetStrategyHolders returns the list of holders for the given strategy.
// The returned map has the holder's address as key and the holder's balance as a string encoded big.Int
func (c3 *Client) GetStrategyHolders(strategyID uint64) (map[string]string, error) {
	return requestWithRetry(func() (map[string]string, error) {
		holders, err := c3.GetHoldersByStrategy(strategyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get strategy holders: %w", err)
		}
		return holders.Holders, nil
	})
}

// Census returns the census with the given ID.
func (c3 *Client) Census(censusID uint64) (*c3types.Census, error) {
	return requestWithRetry(func() (*c3types.Census, error) {
		census, err := c3.GetCensus(censusID)
		if err != nil {
			return nil, fmt.Errorf("failed to get census: %w", err)
		}
		return census, nil

	})
}

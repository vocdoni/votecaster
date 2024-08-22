package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gqlClient "github.com/Khan/genqlient/graphql"
	gql "github.com/vocdoni/vote-frame/airstack/graphql"
	"go.vocdoni.io/dvote/log"
)

// Constants
const (
	apiTimeout       = 10 * time.Second
	airstackAPIlimit = 200
	maxAPIRetries    = 3
)

// airstackSupportedBlockchains represent all supported airstack networks
// add new blockchains if airstack schema changed and bindings regenerated
var airstackSupportedBlockchains = map[string]gql.TokenBlockchain{
	"ethereum": gql.TokenBlockchainEthereum,
	"base":     gql.TokenBlockchainBase,
	"zora":     gql.TokenBlockchainZora,
	"degen":    gql.TokenBlockchainDegen,
}

type httpTransportWithAuth struct {
	key        string
	underlying http.RoundTripper
}

// RoundTrip executes a single HTTP transaction after adding the Authorization header.
func (t *httpTransportWithAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	// ensure the original request is not modified
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.key)
	return t.underlying.RoundTrip(clone)
}

// Client manages the API client for Airstack.
type Client struct {
	gqlClient.Client
	apiKey      string
	url         string
	blockchains []string
	ctx         context.Context
}

// NewClient initializes a new Airstack client.
func NewClient(ctx context.Context, endpoint, apiKey string, blockchains []string) (*Client, error) {
	if endpoint == "" || apiKey == "" {
		return nil, fmt.Errorf("endpoint and apiKey are required")
	}
	ac := &Client{
		apiKey: apiKey,
		url:    endpoint,
		ctx:    ctx,
	}
	// check if api requested blockchains are supported
	for _, chain := range blockchains {
		if _, ok := airstackSupportedBlockchains[chain]; !ok {
			log.Warnf("requested network %s not supported", chain)
			continue
		}
		ac.blockchains = append(ac.blockchains, chain)
	}
	ac.Client = gqlClient.NewClient(endpoint, &http.Client{
		Timeout: apiTimeout,
		Transport: &httpTransportWithAuth{
			key:        apiKey,
			underlying: http.DefaultTransport,
		},
	})

	return ac, nil
}

func (c *Client) ApiKey() string {
	return c.apiKey
}

func (c *Client) blockchainToTokenBlockchain(b string) (gql.TokenBlockchain, bool) {
	tokenBlockchain, ok := airstackSupportedBlockchains[b]
	if !ok {
		return "", false
	}
	return tokenBlockchain, true
}

// Blockchains return an array of available Airstack EVM networks
func (c *Client) Blockchains() []string {
	return c.blockchains
}

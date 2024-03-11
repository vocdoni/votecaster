package airstack

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gqlClient "github.com/Khan/genqlient/graphql"
	gql "github.com/vocdoni/vote-frame/airstack/graphql"
)

// Constants
const (
	apiTimeout       = 10 * time.Second
	airstackAPIlimit = 200
	maxAPIRetries    = 3
)

// keep up to date when regenerating bindings for Airstack schema
const (
	// BlockchainEthereum is the blockchain type used for querying Ethereum Mainnet
	BlockchainEthereum = "ethereum"
	// BlockchainPolygon is the blockchain type used for querying Polygon
	BlockchainPolygon = "polygon"
	// BlockchainBase is the blockchain type used for querying Base
	BlockchainBase = "base"
	// BlockchainZora is the blockchain type used for querying Zora
	BlockchainZora = "zora"
	// BlockchainALL is a special blockchain type required for some queries
	BlockchainALL = "ALL"
)

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
	apiKey string
	url    string
	ctx    context.Context
}

// NewClient initializes a new Airstack client.
func NewClient(ctx context.Context, endpoint, apiKey string) (*Client, error) {
	if endpoint == "" || apiKey == "" {
		return nil, fmt.Errorf("endpoint and apiKey are required")
	}

	ac := &Client{
		apiKey: apiKey,
		url:    endpoint,
		ctx:    ctx,
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

func BlockchainToTokenBlockchain(b string) (gql.TokenBlockchain, error) {
	switch b {
	case "ethereum":
		return gql.TokenBlockchainEthereum, nil
	case "polygon":
		return gql.TokenBlockchainPolygon, nil
	case "base":
		return gql.TokenBlockchainBase, nil
	case "zora":
		return gql.TokenBlockchainZora, nil
	}
	return "", fmt.Errorf("invalid blockchain")
}

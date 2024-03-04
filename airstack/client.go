package airstack

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gqlClient "github.com/Khan/genqlient/graphql"
)

// Constants
const (
	apiTimeout       = 60 * time.Second
	airstackAPIlimit = 200
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

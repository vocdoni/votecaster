package airstack

import (
	"context"
	"fmt"
	"strings"

	ac "github.com/vocdoni/vote-frame/airstack/client"
)

// Airstack wraps all the required artifacts for interacting with the Airstack API
type Airstack struct {
	*ac.Client
}

// NewAirstack creates a new Airstack artifact with a reference to a MongoDB and an Airstack client that
// enables to make predefined queries to the Airstack GraphQL API.
func NewAirstack(ctx context.Context, endpoint, apiKey, supportedBlockchains string) (*Airstack, error) {
	client, err := ac.NewClient(ctx, endpoint, apiKey, strings.Split(supportedBlockchains, ","))
	if err != nil {
		return nil, fmt.Errorf("error creating Airstack: %w", err)
	}
	return &Airstack{
		Client: client,
	}, nil
}

package main

import (
	"context"
	"fmt"

	ac "github.com/vocdoni/vote-frame/airstack"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
)

// Airstack wraps all the required artifacts for interacting with the Airstack API
type Airstack struct {
	client *ac.Client
}

// NewAirstack creates a new Airstack artifact with a reference to a MongoDB and an Airstack client that
// enables to make predefined queries to the Airstack GraphQL API.
func NewAirstack(ctx context.Context, endpoint, apiKey string) (*Airstack, error) {
	client, err := ac.NewClient(ctx, endpoint, apiKey)
	if err != nil {
		return nil, fmt.Errorf("error creating Airstack: %w", err)
	}
	return &Airstack{
		client: client,
	}, nil
}

func (v *vocdoniHandler) createCensusChannel(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	return nil
}

func (v *vocdoniHandler) createCensusTwoTokens(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	return nil
}

package airstack

import (
	"context"
	"fmt"

	airstack "github.com/vocdoni/vote-frame/airstack/client"
	"github.com/vocdoni/vote-frame/mongo"
)

// Airstack wraps all the required artifacts for interacting with the Airstack API
type Airstack struct {
	db     *mongo.MongoStorage
	client *airstack.Client
}

// NewAirstack creates a new Airstack artifact with a reference to a MongoDB and an Airstack client that
// enables to make predefined queries to the Airstack GraphQL API.
func NewAirstack(ctx context.Context, db *mongo.MongoStorage, clientConf *airstack.ClientConf) (*Airstack, error) {
	client, err := airstack.NewClient(ctx, clientConf)
	if err != nil {
		return nil, fmt.Errorf("error creating Airstack: %w", err)
	}
	return &Airstack{
		db:     db,
		client: client,
	}, nil
}

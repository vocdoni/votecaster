package airstack

import (
	"context"
	"fmt"

	airstack "github.com/vocdoni/vote-frame/airstack/client"
	"github.com/vocdoni/vote-frame/mongo"
)

type Airstack struct {
	db     *mongo.MongoStorage
	client *airstack.Client
}

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

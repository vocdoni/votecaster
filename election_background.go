package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

// createApiClientsForElectionRecovery creates a list of API clients that will be used to retrieve the election metadata.
func createApiClientsForElectionRecovery() []*apiclient.HTTPclient {
	// Define the API clients that will be used to retrieve the election metadata
	apiClientsHost := []string{"https://api.vocdoni.io/v2", "https://api-stg.vocdoni.net/v2", "https://api-dev.vocdoni.net/v2"}
	var apiClients []*apiclient.HTTPclient
	for _, apiClient := range apiClientsHost {
		cli, err := apiclient.New(apiClient)
		if err != nil {
			log.Errorw(err, "failed to create API client")
			continue
		}
		apiClients = append(apiClients, cli)
	}
	return apiClients
}

// recoverElectionFromMultipleEndpoints retrieves the election metadata from multiple API endpoints.
// If not found, it returns nil.
func recoverElectionFromMultipleEndpoints(electionID types.HexBytes, apiClients []*apiclient.HTTPclient) *api.Election {
	var apiElection *api.Election
	var err error
	for _, cli := range apiClients {
		apiElection, err = cli.Election(electionID)
		if err != nil || apiElection == nil {
			continue
		}
		break
	}
	return apiElection
}

// finalizeElectionsAtBackround checks for elections without results and finalizes them.
// Stores the final results as a static PNG image in the database. It must run in the background.
func finalizeElectionsAtBackround(ctx context.Context, v *vocdoniHandler) {
	apiClients := createApiClientsForElectionRecovery()
	if len(apiClients) == 0 {
		log.Error("failed to create any API client, aborting")
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(60 * time.Second):
			electionIDs, err := v.db.ElectionsWithoutResults()
			if err != nil {
				if mongo.IsDBClosed(err) {
					log.Warn("database client is disconnected")
					return
				}
				log.Errorw(err, "failed to get elections without results")
				continue
			}
			for _, electionID := range electionIDs {
				electionIDbytes, err := hex.DecodeString(electionID)
				if err != nil {
					log.Errorw(err, fmt.Sprintf("failed to decode electionID: %s", electionID))
					continue
				}
				election := recoverElectionFromMultipleEndpoints(electionIDbytes, apiClients)
				if election == nil {
					continue
				}
				if election.FinalResults {
					electiondb, err := v.db.Election(electionIDbytes)
					if err != nil {
						continue
					}
					if _, err = v.finalizeElectionResults(election, electiondb); err != nil {
						log.Errorw(err, fmt.Sprintf("failed to finalize election results: %x", electionIDbytes))
					}
				}
			}
		}
	}
}

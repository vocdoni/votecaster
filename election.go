package main

import (
	"context"
	"fmt"
	"time"

	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

// ElectionDescription defines the parameters for a new election.
type ElectionDescription struct {
	Question  string        `json:"question"`
	Options   []string      `json:"options"`
	Duration  time.Duration `json:"duration"`
	Overwrite bool          `json:"overwrite"`
}

func newElectionDescription(description *ElectionDescription, census *CensusInfo) *api.ElectionDescription {
	choices := []api.ChoiceMetadata{}

	for i, choice := range description.Options {
		choices = append(choices, api.ChoiceMetadata{
			Title: map[string]string{"default": choice},
			Value: uint32(i + 1),
		})
	}

	size := census.Size
	if size > maxElectionSize {
		size = maxElectionSize
	}

	return &api.ElectionDescription{
		Title:       map[string]string{"default": description.Question},
		Description: map[string]string{"default": ""},
		EndDate:     time.Now().Add(description.Duration),

		Questions: []api.Question{
			{
				Title:       map[string]string{"default": description.Question},
				Description: map[string]string{"default": ""},
				Choices:     choices,
			},
		},

		ElectionType: api.ElectionType{
			Autostart: true,
		},
		VoteType: api.VoteType{
			MaxVoteOverwrites: func() int {
				if description.Overwrite {
					return 1
				}
				return 0
			}(),
		},
		TempSIKs: false,
		Census: api.CensusTypeDescription{
			Type:     api.CensusTypeFarcaster,
			RootHash: census.Root,
			URL:      census.Url,
			Size:     size,
		},
	}
}

// createElection creates a new election with the given description and census. Waits until the election is created or returns an error.
func createElection(cli *apiclient.HTTPclient, description *ElectionDescription, census *CensusInfo) (types.HexBytes, error) {
	electionID, err := cli.NewElection(newElectionDescription(description, census))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	election, err := cli.WaitUntilElectionCreated(ctx, electionID)
	if err != nil {
		log.Fatal(err)
	}
	log.Infow("created new election", "id", election.ElectionID.String())
	return electionID, nil
}

// ensureAccountExist checks if the account exists and creates it if it doesn't.
func ensureAccountExist(cli *apiclient.HTTPclient) error {
	account, err := cli.Account("")
	if err == nil {
		log.Infow("account already exists", "address", account.Address)
		return nil
	}

	log.Infow("creating new account", "address", cli.MyAddress().Hex())
	faucetPkg, err := apiclient.GetFaucetPackageFromDefaultService(cli.MyAddress().Hex(), cli.ChainID())
	if err != nil {
		return fmt.Errorf("failed to get faucet package: %w", err)
	}
	accountMetadata := &api.AccountMetadata{
		Name:        map[string]string{"default": "Farcaster frame proxy " + cli.MyAddress().Hex()},
		Description: map[string]string{"default": "Farcaster frame proxy account"},
		Version:     "1.0",
	}
	hash, err := cli.AccountBootstrap(faucetPkg, accountMetadata, nil)
	if err != nil {
		return fmt.Errorf("failed to bootstrap account: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*40)
	defer cancel()
	if _, err := cli.WaitUntilTxIsMined(ctx, hash); err != nil {
		return fmt.Errorf("failed to wait for tx to be mined: %w", err)
	}
	return nil
}

package main

import (
	"encoding/hex"
	"encoding/json"
	"os"

	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
	"go.vocdoni.io/dvote/vochain/state"
)

var testPubKeys = []string{
	"ec327cd438995a59ce78fddd29631e9b2e41eafc3a6946dd26b4da749f47140d",
	"d843cf9636184c99fe6ed7db8b5934c2cd31dd19b63a2409b305898fa560459c",
	"f2ccdd260028c7b0c3e6425178534d7b33948cf6bd388fd9b77db93ffd2bd875",
	"d2cd0b44adc7355bd407a6c9216bfc053f2de7e26e48dd3c7e5a44f07af004e8",
	"b44b21f817a968423a3669c13865deafa7389ae0df89c6ad615cfc17b86118f8",
	"d6424e655287aa61df38205da19ddab23b0ff9683c6800e0dbc3e8b65d3eb2e3",
}

const (
	// maxElectionSize is the maximum number of participants in an election
	maxElectionSize = 15000
)

type CensusInfo struct {
	Root []byte `json:"root"`
	Url  string `json:"url"`
	Size uint64 `json:"size"`
}

func (c *CensusInfo) FromFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}

// createTestCensus creates a test census with the hardcoded public keys for testing purposes.
func createTestCensus(cli *apiclient.HTTPclient) (*CensusInfo, error) {
	censusID, err := cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return nil, err
	}
	censusList := api.CensusParticipants{}
	for _, pubkey := range testPubKeys {
		pubBytes, err := hex.DecodeString(pubkey)
		if err != nil {
			log.Error(err)
			continue
		}
		voterID := state.NewVoterID(state.VoterIDTypeEd25519, pubBytes)
		censusList.Participants = append(censusList.Participants, api.CensusParticipant{
			Key:    voterID.Address(),
			Weight: new(types.BigInt).SetUint64(1),
		})
	}

	if err := cli.CensusAddParticipants(censusID, &censusList); err != nil {
		return nil, err
	}

	root, url, err := cli.CensusPublish(censusID)
	if err != nil {
		return nil, err
	}
	size, err := cli.CensusSize(censusID)
	if err != nil {
		return nil, err
	}

	return &CensusInfo{
		Root: root,
		Url:  url,
		Size: size,
	}, nil
}

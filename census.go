package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
	"go.vocdoni.io/dvote/vochain/state"
)

const (
	devMaxElectionSize     = 5000
	stageMaxElectionSize   = 100000
	defaultMaxElectionSize = 200000
)

var (
	// maxElectionSize is the maximum number of participants in an election
	maxElectionSize = defaultMaxElectionSize

	// ErrNoValidParticipants is returned when no valid participants are found
	ErrNoValidParticipants = fmt.Errorf("no valid participants")
)

type CensusInfo struct {
	Root types.HexBytes `json:"root"`
	Url  string         `json:"uri"`
	Size uint64         `json:"size"`
}

func (c *CensusInfo) FromFile(file string) error {
	log.Debugw("loading census from file", "file", file)
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}

// FarcasterParticipant is a participant in the Farcaster network to be included in the census.
type FarcasterParticipant struct {
	PubKey []byte   `json:"pubkey"`
	Weight *big.Int `json:"weight"`
}

// createTestCensus creates a test census with the hardcoded public keys for testing purposes.
func CreateCensus(cli *apiclient.HTTPclient, participants []*FarcasterParticipant) (*CensusInfo, error) {
	censusList := api.CensusParticipants{}
	for _, p := range participants {
		voterID := state.NewVoterID(state.VoterIDTypeEd25519, p.PubKey)
		censusList.Participants = append(censusList.Participants, api.CensusParticipant{
			Key:    voterID.Address(),
			Weight: (*types.BigInt)(p.Weight),
		})
	}
	if len(censusList.Participants) == 0 {
		return nil, ErrNoValidParticipants
	}

	censusID, err := cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return nil, err
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

func (v *vocdoniHandler) farcasterCensusFromEthereumCSV(csv []byte) ([]*FarcasterParticipant, error) {
	records, err := ParseCSV(csv)
	if err != nil {
		return nil, err
	}
	participants := make([]*FarcasterParticipant, 0, len(records))
	for i, record := range records {
		if len(record) != 2 {
			return nil, fmt.Errorf("invalid record: %v", record)
		}
		var address common.Address
		var weight *big.Int
		var ok bool
		if common.IsHexAddress(record[0]) {
			address = common.HexToAddress(record[0])
			weight, ok = new(big.Int).SetString(record[1], 10)
			if !ok {
				return nil, fmt.Errorf("invalid weight on line %d: %v", i, record[1])
			}
		} else if common.IsHexAddress(record[1]) {
			address = common.HexToAddress(record[1])
			weight, ok = new(big.Int).SetString(record[0], 10)
			if !ok {
				return nil, fmt.Errorf("invalid weight on line %d: %v", i, record[0])
			}
		} else {
			log.Warnw("invalid record", "c1", record[0], "c2", record[1])
			continue
		}
		// Discover the farcaster pubkeys from address
		participants = append(participants, &FarcasterParticipant{
			// ...
			PubKey: address.Bytes(),
			Weight: weight,
		})
	}
	return participants, nil
}

func ParseCSV(csv []byte) ([][]string, error) {
	records := [][]string{}
	lines := bytes.Split(csv, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		record := strings.Split(string(line), ",")
		// remove spaces and tabs
		for i, field := range record {
			record[i] = strings.TrimSpace(field)
		}
		records = append(records, record)
	}
	return records, nil
}

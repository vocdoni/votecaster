package main

import (
	"context"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vocdoni/farcaster-poc/farcasterapi"
	"github.com/vocdoni/farcaster-poc/mongo"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
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
	Root      types.HexBytes `json:"root"`
	Url       string         `json:"uri"`
	Size      uint64         `json:"size"`
	Error     string         `json:"-"`
	Usernames []string       `json:"usernames,omitempty"`
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
	PubKey   []byte   `json:"pubkey"`
	Weight   *big.Int `json:"weight"`
	Username string   `json:"username"`
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

// censusCSV creates a new census from a CSV file containing Ethereum addresses and weights.
// It builds the census async and returns the census ID.
func (v *vocdoniHandler) censusCSV(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	censusID, err := v.cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return err
	}
	v.censusCreationMap.Store(censusID.String(), &CensusInfo{})

	go func() {
		fmt.Println(string(msg.Data))
		participants, err := v.farcasterCensusFromEthereumCSV(msg.Data)
		if err != nil {
			log.Warnw("failed to build census from ethereum csv", "err", err.Error())
			v.censusCreationMap.Store(censusID.String(), &CensusInfo{Error: err.Error()})
			return
		}
		ci, err := CreateCensus(v.cli, participants)
		if err != nil {
			log.Errorw(err, "failed to create census")
			v.censusCreationMap.Store(censusID.String(), &CensusInfo{Error: err.Error()})
			return
		}
		for _, p := range participants {
			ci.Usernames = append(ci.Usernames, p.Username)
		}
		v.censusCreationMap.Store(censusID.String(), ci)
	}()
	data, err := json.Marshal(map[string]string{"censusId": censusID.String()})
	if err != nil {
		return err
	}
	return ctx.Send(data, http.StatusOK)
}

// censusQueueInfo returns the status of the census creation process.
// Returns 204 if the census is not yet ready or not found.
func (v *vocdoniHandler) censusQueueInfo(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	censusID := ctx.URLParam("censusID")
	censusInfo, ok := v.censusCreationMap.Load(censusID)
	if !ok {
		return ctx.Send(nil, http.StatusNotFound)
	}
	if censusInfo.(*CensusInfo).Error != "" {
		return ctx.Send([]byte(censusInfo.(*CensusInfo).Error), http.StatusInternalServerError)
	}
	if censusInfo.(*CensusInfo).Root == nil {
		return ctx.Send(nil, http.StatusNoContent)
	}
	data, err := json.Marshal(censusInfo)
	if err != nil {
		return err
	}
	return ctx.Send(data, http.StatusOK)
}

func (v *vocdoniHandler) farcasterCensusFromEthereumCSV(csv []byte) ([]*FarcasterParticipant, error) {
	records, err := ParseCSV(csv)
	if err != nil {
		return nil, err
	}
	log.Debugw("parsed csv", "records", len(records))
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

		var signers []string
		var username string
		// Try to fetch the user from the database
		user, err := v.db.UserByAddress(address.String())
		if err != nil {
			if !errors.Is(err, mongo.ErrUserUnknown) {
				log.Errorw(err, "failed to get user from database")
			}
			// Fetch the user data from the farcaster API
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			userData, err := v.fcapi.UserDataByVerificationAddress(ctx, address.String())
			if err != nil {
				if !errors.Is(err, farcasterapi.ErrNoDataFound) {
					log.Warnw("error fetching user data", "address", address.String(), "err", err)
				}
				continue
			}
			// Add or update the user on the database
			dbUser, err := v.db.User(userData.FID)
			if err != nil {
				if !errors.Is(err, mongo.ErrUserUnknown) {
					log.Errorw(err, "failed to get user from database")
				}
				if err := v.db.AddUser(userData.FID, userData.Username, userData.VerificationsAddresses, userData.Signers, 0); err != nil {
					log.Errorw(err, "failed to add user to database")
				}
			} else {
				dbUser.Addresses = userData.VerificationsAddresses
				dbUser.Username = userData.Username
				dbUser.Signers = userData.Signers
				if err := v.db.UpdateUser(dbUser); err != nil {
					log.Errorw(err, "failed to update user in database")
				}
			}
			signers = userData.Signers
			username = userData.Username
		} else {
			// Use the user data from the database
			log.Debugw("user found in database", "user", user)
			signers = user.Signers
			username = user.Username
		}
		// Add all the signers to the participants list
		for _, signer := range signers {
			signerBytes, err := hex.DecodeString(strings.TrimPrefix(signer, "0x"))
			if err != nil {
				log.Warnw("error decoding signer", "signer", signer, "err", err)
				continue
			}
			participants = append(participants, &FarcasterParticipant{
				PubKey:   signerBytes,
				Weight:   weight,
				Username: username,
			})
		}
	}
	log.Infow("farcaster census from ethereum csv", "signers", len(participants))
	return participants, nil
}

func ParseCSV(csvData []byte) ([][]string, error) {
	// Convert the byte slice to a reader
	r := csv.NewReader(strings.NewReader(string(csvData)))
	r.Comment = '#'
	r.TrimLeadingSpace = true // trim leading space of each field
	r.FieldsPerRecord = 2     // expect 2 fields per record
	var records [][]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

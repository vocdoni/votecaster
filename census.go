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
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/farcasterapi/neynar"
	"github.com/vocdoni/vote-frame/mongo"
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
	maxNumOfCsvRecords     = 2000
)

var (
	// maxElectionSize is the maximum number of participants in an election
	maxElectionSize = defaultMaxElectionSize

	// ErrNoValidParticipants is returned when no valid participants are found
	ErrNoValidParticipants = fmt.Errorf("no valid participants")
	// ErrUserNotFoundInFarcaster is returned when a user is not found in the farcaster API
	ErrUserNotFoundInFarcaster = fmt.Errorf("user not found in farcaster")
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
	return processCensusRecords(records, v.db, v.fcapi)
}

// processRecord processes a single record of a plain-text census and returns the corresponding Farcaster participants.
// The record is expected to be a string containing the address and the weight.
func processCensusRecords(records [][]string, db *mongo.MongoStorage, fcapi farcasterapi.API) ([]*FarcasterParticipant, error) {
	// build a map for unique addresses and their weights
	addressMap := make(map[string]*big.Int)
	for _, record := range records {
		if len(record) != 2 {
			return nil, fmt.Errorf("invalid record: %v", record)
		}
		var weight *big.Int
		var ok bool
		address := ""
		weightRecord := ""
		if common.IsHexAddress(record[0]) {
			address = common.HexToAddress(record[0]).Hex()
			weightRecord = record[1]
		} else if common.IsHexAddress(record[1]) {
			address = common.HexToAddress(record[1]).Hex()
			weightRecord = record[0]
		} else {
			log.Warnf("invalid record: %v", record)
			continue
		}
		weight, ok = new(big.Int).SetString(weightRecord, 10)
		if !ok {
			log.Warnf("invalid weight for address %s: %s", address, weightRecord)
			continue
		}
		// Add the weight to the address if it already exists
		if _, ok := addressMap[address]; ok {
			addressMap[address].Add(addressMap[address], weight)
		} else {
			addressMap[address] = weight
		}
	}

	// Fetch the users from the database concurrently
	var wg sync.WaitGroup
	participantsCh := make(chan *FarcasterParticipant, len(addressMap)) // Channel to collect participants
	pendingAddressesCh := make(chan string, len(addressMap))            // Channel to collect pending addresses

	for address := range addressMap {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			user, err := db.UserByAddress(addr)
			if err != nil {
				if !errors.Is(err, mongo.ErrUserUnknown) {
					log.Warnw("error fetching user from database", "address", addr, "err", err)
					return
				}
				pendingAddressesCh <- addr
				return
			}
			if len(user.Signers) == 0 {
				pendingAddressesCh <- addr
				return
			}
			for _, signer := range user.Signers {
				signerBytes, err := hex.DecodeString(strings.TrimPrefix(signer, "0x"))
				if err != nil {
					log.Warnw("error decoding signer", "signer", signer, "err", err)
					continue
				}
				participantsCh <- &FarcasterParticipant{
					PubKey:   signerBytes,
					Weight:   addressMap[addr],
					Username: user.Username,
				}
			}
		}(address)
	}

	wg.Wait()
	close(participantsCh)
	close(pendingAddressesCh)

	// Collect data from channels
	participants := []*FarcasterParticipant{}
	pendingAddresses := []string{}
	for participant := range participantsCh {
		participants = append(participants, participant)
	}
	for addr := range pendingAddressesCh {
		pendingAddresses = append(pendingAddresses, addr)
	}

	if len(addressMap) > 0 {
		log.Infow("users found on database", "count", len(participants), "ratio", fmt.Sprintf("%.2f%%", 100*float64(len(participants))/float64(len(addressMap))))
	}

	// Fetch the remaining users from the farcaster API
	count := 0
	log.Debugw("fetching users from farcaster", "count", len(pendingAddresses))
	for i := 0; i < len(pendingAddresses); i += neynar.MaxAddressesPerRequest {
		// Fetch the user data from the farcaster API
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		to := i + neynar.MaxAddressesPerRequest
		if to > len(pendingAddresses) {
			to = len(pendingAddresses)
		}
		log.Debugw("fetching users from farcaster", "from", i, "to", to)
		usersData, err := fcapi.UserDataByVerificationAddress(ctx, pendingAddresses[i:to])
		if err != nil {
			if errors.Is(err, farcasterapi.ErrNoDataFound) {
				break
			}
			return nil, err
		}
		for _, userData := range usersData {
			// Add or update the user on the database
			dbUser, err := db.User(userData.FID)
			if err != nil {
				if !errors.Is(err, mongo.ErrUserUnknown) {
					return nil, err
				}
				if err := db.AddUser(
					userData.FID,
					userData.Username,
					userData.VerificationsAddresses,
					userData.Signers,
					userData.CustodyAddress,
					0,
				); err != nil {
					return nil, err
				}
			} else {
				dbUser.Addresses = userData.VerificationsAddresses
				dbUser.Username = userData.Username
				dbUser.Signers = userData.Signers
				dbUser.CustodyAddress = userData.CustodyAddress
				if err := db.UpdateUser(dbUser); err != nil {
					return nil, err
				}
			}
			// find the addres on the map to get the weight
			var weight *big.Int
			var ok bool
			for _, addr := range userData.VerificationsAddresses {
				weight, ok = addressMap[common.HexToAddress(addr).Hex()]
				if ok {
					break
				}
			}
			if weight == nil {
				log.Warnf("weight not found for user %s with address %v", userData.Username, userData.VerificationsAddresses)
				continue
			}
			// Add the user to the participants list (with all the signers)
			for _, signer := range userData.Signers {
				signerBytes, err := hex.DecodeString(strings.TrimPrefix(signer, "0x"))
				if err != nil {
					log.Warnw("error decoding signer", "signer", signer, "err", err)
					continue
				}
				participants = append(participants, &FarcasterParticipant{
					PubKey:   signerBytes,
					Weight:   weight,
					Username: userData.Username,
				})
			}
			count++
		}
	}
	if len(pendingAddresses) > 0 {
		log.Infow("users found on farcaster", "count", count, "ratio", fmt.Sprintf("%.2f%%", 100*float64(count)/float64(len(pendingAddresses))))
	}
	return participants, nil
}

func ParseCSV(csvData []byte) ([][]string, error) {
	// Convert the byte slice to a reader
	r := csv.NewReader(strings.NewReader(string(csvData)))
	r.Comment = '#'
	r.TrimLeadingSpace = true // trim leading space of each field
	r.FieldsPerRecord = 2     // expect 2 fields per record
	count := 0
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
		count++
		if count > maxNumOfCsvRecords {
			return nil, fmt.Errorf("max number of records exceeded")
		}
	}
	return records, nil
}

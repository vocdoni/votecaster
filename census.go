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
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	maxNumOfCsvRecords     = 10000

	POAP_CSV_HEADER = "ID,Collection,ENS,Minting Date,Tx Count,Power"
)

var (
	// maxElectionSize is the maximum number of participants in an election
	maxElectionSize = defaultMaxElectionSize

	// ErrNoValidParticipants is returned when no valid participants are found
	ErrNoValidParticipants = fmt.Errorf("no valid participants")
	// ErrUserNotFoundInFarcaster is returned when a user is not found in the farcaster API
	ErrUserNotFoundInFarcaster = fmt.Errorf("user not found in farcaster")
)

// FrameCensusType is a custom type to identify the different types of censuses.
type FrameCensusType int

const (
	// FrameCensusTypeAllFarcaster is the default census type and includes all
	// the users in the Farcaster network.
	FrameCensusTypeAllFarcaster FrameCensusType = iota
	// FrameCensusTypeCSV is a census created from a CSV file containing
	// Ethereum addresses and weights.
	FrameCensusTypeCSV
	// FrameCensusTypeChannelGated is a census created from the users who follow
	// a specific Warpcast Channel.
	FrameCensusTypeChannelGated
	// FrameCensusTypeFollowers is a census created from the users who follow a
	// specific user in the Farcaster network.
	FrameCensusTypeFollowers
	// FrameCensusTypeFile is a census created from a file.
	FrameCensusTypeFile
)

// CensusInfo contains the information of a census.
type CensusInfo struct {
	Root               types.HexBytes `json:"root"`
	Url                string         `json:"uri"`
	Size               uint64         `json:"size"`
	Usernames          []string       `json:"usernames,omitempty"`
	FromTotalAddresses uint32         `json:"fromTotalAddresses,omitempty"`

	Error    string          `json:"-"`
	Progress uint32          `json:"-"` // Progress of the census creation process (0-100)
	Type     FrameCensusType `json:"-"` // Type of the census
}

// FromFile loads the census information from a file.
func (c *CensusInfo) FromFile(file string) error {
	log.Debugw("loading census from file", "file", file)
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}
	// Set the type of the census
	c.Type = FrameCensusTypeFile
	return nil
}

// FarcasterParticipant is a participant in the Farcaster network to be included in the census.
type FarcasterParticipant struct {
	PubKey   []byte   `json:"pubkey"`
	Weight   *big.Int `json:"weight"`
	Username string   `json:"username"`
}

// CreateCensus creates a new census from a list of participants.
func CreateCensus(cli *apiclient.HTTPclient, participants []*FarcasterParticipant,
	censusType FrameCensusType,
) (*CensusInfo, error) {
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
		Type: censusType,
	}, nil
}

// censusCSV creates a new census from a CSV file containing Ethereum addresses and weights.
// It builds the census async and returns the census ID.
func (v *vocdoniHandler) censusCSV(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	censusID, err := v.cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return err
	}
	v.censusCreationMap.Store(censusID.String(), CensusInfo{})
	totalCSVaddresses := uint32(0)
	go func() {
		startTime := time.Now()
		log.Debugw("building census from csv", "censusID", censusID)
		var participants []*FarcasterParticipant
		participants, totalCSVaddresses, err = v.farcasterCensusFromEthereumCSV(msg.Data, censusID)
		if err != nil {
			log.Warnw("failed to build census from ethereum csv", "err", err.Error())
			v.censusCreationMap.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		ci, err := CreateCensus(v.cli, participants, FrameCensusTypeCSV)
		if err != nil {
			log.Errorw(err, "failed to create census")
			v.censusCreationMap.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}

		// since each participant can have multiple signers, we need to get the unique usernames
		uniqueParticipantsMap := make(map[string]struct{})
		for _, p := range participants {
			uniqueParticipantsMap[p.Username] = struct{}{}
		}
		uniqueParticipants := []string{}
		for k := range uniqueParticipantsMap {
			uniqueParticipants = append(uniqueParticipants, k)
		}
		ci.Usernames = uniqueParticipants
		ci.FromTotalAddresses = totalCSVaddresses
		log.Infow("census created from CSV",
			"censusID", censusID.String(),
			"size", len(ci.Usernames),
			"duration", time.Since(startTime),
			"fromTotalAddresses", totalCSVaddresses)
		v.censusCreationMap.Store(censusID.String(), *ci)
	}()
	data, err := json.Marshal(map[string]string{"censusId": censusID.String()})
	if err != nil {
		return err
	}
	return ctx.Send(data, http.StatusOK)
}

// censusChannelExists checks if a Warpcast Channel exists. It returns a NotFound
// error if the channel does not exist. If the channelID is not provided, it
// returns a BadRequest error. If the channel exists, it returns a 200 OK.
func (v *vocdoniHandler) censusChannelExists(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	channelID := ctx.URLParam("channelID")
	if channelID == "" {
		return ctx.Send([]byte("channelID is required"), http.StatusBadRequest)
	}
	exists, err := v.fcapi.ChannelExists(channelID)
	if err != nil {
		return err
	}
	if !exists {
		return ctx.Send(nil, http.StatusNotFound)
	}
	return ctx.Send(nil, http.StatusOK)
}

// censusChannel creates a new census that includes the users who follow a
// specific Warpcast Channel. It builds the census async and returns the census
// ID. If no channelID is provided, it returns a BadRequest error. If the
// channel does not exist, it returns a NotFound error. The process of creating
// the census includes fetching the users FIDs from the farcaster API and
// querying the database to get the users signer keys. The census is created
// from the participants and the progress is updated in the queue.
func (v *vocdoniHandler) censusChannel(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// check if channelID is provided, it is required so if it's not provided
	// return a BadRequest error
	channelID := ctx.URLParam("channelID")
	if channelID == "" {
		return ctx.Send([]byte("channelID is required"), http.StatusBadRequest)
	}
	// check if the channel exists, if not return a NotFound error. If something
	// fails when checking the channel existence, return the error.
	exists, err := v.fcapi.ChannelExists(channelID)
	if err != nil {
		return err
	}
	if !exists {
		return ctx.Send([]byte("channel not found"), http.StatusNotFound)
	}
	// create a censusID for the queue and store into it
	censusID, err := v.cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return err
	}
	v.censusCreationMap.Store(censusID.String(), CensusInfo{})
	// run a goroutine to create the census, update the queue with the progress,
	// and update the queue result when it's ready
	go func() {
		internalCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// get the fids of the users in the channel from neynar farcaster API, if
		// the channel does not exist, return a NotFound error
		users, err := v.fcapi.ChannelFIDs(internalCtx, channelID)
		if err != nil {
			v.censusCreationMap.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		participants, errFids := v.farcasterCensusFromFids(internalCtx, users, censusID)
		for fid, err := range errFids {
			if errors.Is(err, mongo.ErrUserUnknown) {
				log.Warnw("user not found in database", "fid", fid)
			} else {
				log.Warnw("error fetching user from database", "fid", fid, "error", err)
			}
		}
		if len(participants) == 0 {
			v.censusCreationMap.Store(censusID.String(), CensusInfo{Error: "no valid participants"})
			return
		}
		// create the census from the participants
		censusInfo, err := CreateCensus(v.cli, participants, FrameCensusTypeChannelGated)
		if err != nil {
			v.censusCreationMap.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		v.censusCreationMap.Store(censusID.String(), *censusInfo)
		log.Infow("census created from channel",
			"channelID", channelID,
			"errorFids", len(errFids),
			"successFids", len(users)-len(errFids),
			"totalFids", len(users),
			"participants", len(participants))
	}()
	// return the censusID to the client
	data, err := json.Marshal(map[string]string{"censusId": censusID.String()})
	if err != nil {
		return err
	}
	return ctx.Send(data, http.StatusOK)
}

func (v *vocdoniHandler) censusFollowers(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// check if userFid is provided, it is required so if it's not provided
	// return a BadRequest error
	strUserFid := ctx.URLParam("userFid")
	if strUserFid == "" {
		return ctx.Send([]byte("userFid is required"), http.StatusBadRequest)
	}
	userFid, err := strconv.ParseUint(strUserFid, 10, 64)
	if err != nil {
		return ctx.Send([]byte("invalid userFid"), http.StatusBadRequest)
	}
	// create a censusID for the queue and store into it
	censusID, err := v.cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return err
	}
	v.censusCreationMap.Store(censusID.String(), CensusInfo{})
	// run a goroutine to create the census, update the queue with the progress,
	// and update the queue result when it's ready
	go func() {
		internalCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		users, err := v.fcapi.UserFollowers(internalCtx, userFid)
		if err != nil {
			v.censusCreationMap.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		participants, errFids := v.farcasterCensusFromFids(internalCtx, users, censusID)
		for fid, err := range errFids {
			if errors.Is(err, mongo.ErrUserUnknown) {
				log.Warnw("user not found in database", "fid", fid)
			} else {
				log.Warnw("error fetching user from database", "fid", fid, "error", err)
			}
		}
		if len(participants) == 0 {
			v.censusCreationMap.Store(censusID.String(), CensusInfo{Error: "no valid participants"})
			return
		}
		// create the census from the participants
		censusInfo, err := CreateCensus(v.cli, participants, FrameCensusTypeFollowers)
		if err != nil {
			v.censusCreationMap.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		v.censusCreationMap.Store(censusID.String(), *censusInfo)
		log.Infow("census created from user followers",
			"userFid", userFid,
			"errorFids", len(errFids),
			"successFids", len(users)-len(errFids),
			"totalFids", len(users),
			"participants", len(participants))
	}()
	// return the censusID to the client
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
	iCensusInfo, ok := v.censusCreationMap.Load(censusID)
	if !ok {
		return ctx.Send(nil, http.StatusNotFound)
	}
	censusInfo, ok := iCensusInfo.(CensusInfo)
	if !ok {
		return ctx.Send(nil, http.StatusNotFound)
	}
	if censusInfo.Error != "" {
		return ctx.Send([]byte(censusInfo.Error), http.StatusInternalServerError)
	}
	if censusInfo.Root == nil {
		data, err := json.Marshal(map[string]uint32{
			"progress": censusInfo.Progress,
		})
		if err != nil {
			return err
		}
		return ctx.Send(data, http.StatusAccepted)
	}
	if censusInfo.Type != FrameCensusTypeCSV {
		censusInfo.Usernames = nil
	}
	data, err := json.Marshal(censusInfo)
	if err != nil {
		return err
	}
	return ctx.Send(data, http.StatusOK)
}

func (v *vocdoniHandler) farcasterCensusFromEthereumCSV(csv []byte, censusID types.HexBytes) ([]*FarcasterParticipant, uint32, error) {
	records, err := ParseCSV(csv)
	if err != nil {
		return nil, 0, err
	}
	return v.processCensusRecords(records, censusID)
}

// farcasterCensusFromFids creates a list of Farcaster participants from a list
// of FIDs. It queries the database to get the users signer keys and creates the
// participants from them. It returns the list of participants and a map of the
// FIDs that failed to get the users from the database or decoding the keys.
func (v *vocdoniHandler) farcasterCensusFromFids(ctx context.Context, fids []uint64,
	censusID types.HexBytes,
) ([]*FarcasterParticipant, map[uint64]error) {
	// get participants from the users fids, quering the database and to get the
	// users public keys
	participants := []*FarcasterParticipant{}
	totalFids := uint32(len(fids))
	errorFids := make(map[uint64]error)
	for i, fid := range fids {
		v.updateCensusProgress(ctx, censusID, 100*uint32(i)/totalFids)
		user, err := v.db.User(fid)
		if err != nil {
			errorFids[fid] = err
			continue
		}
		for _, signer := range user.Signers {
			signerBytes, err := hex.DecodeString(strings.TrimPrefix(signer, "0x"))
			if err != nil {
				errorFids[fid] = err
				continue
			}
			participants = append(participants, &FarcasterParticipant{
				PubKey:   signerBytes,
				Weight:   big.NewInt(1),
				Username: user.Username,
			})
		}
	}
	return participants, errorFids
}

func (v *vocdoniHandler) updateCensusProgress(ctx context.Context, censusID types.HexBytes, progress uint32) {
	select {
	case <-ctx.Done():
		return
	default:
		censusInfo, ok := v.censusCreationMap.Load(censusID.String())
		if !ok {
			return
		}
		ci := censusInfo.(CensusInfo)
		// no need to update the progress if the census is already ready
		if ci.Root != nil {
			return
		}
		ci.Progress = progress
		v.censusCreationMap.Store(censusID.String(), ci)
	}
}

// processRecord processes a single record of a plain-text census and returns the corresponding Farcaster participants.
// The record is expected to be a string containing the address and the weight.
// Returns the list of participants and the total number of unique addresses available in the records.
func (v *vocdoniHandler) processCensusRecords(records [][]string, censusID types.HexBytes) ([]*FarcasterParticipant, uint32, error) {
	// Create a context to cancel the goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// build a map for unique addresses and their weights
	addressMap := make(map[string]*big.Int)
	for _, record := range records {
		if len(record) != 2 {
			return nil, 0, fmt.Errorf("invalid record: %v", record)
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

	uniqueAddressesCount := uint32(len(addressMap))
	if uniqueAddressesCount == 0 {
		return nil, 0, ErrNoValidParticipants
	}

	// Fetch the users from the database concurrently
	var wg sync.WaitGroup
	participantsCh := make(chan *FarcasterParticipant) // Channel to collect participants
	pendingAddressesCh := make(chan string)            // Channel to collect pending addresses
	concurrencyLimit := make(chan struct{}, 10)        // Concurrency limiter, N is the max number of goroutines
	participants := []*FarcasterParticipant{}
	pendingAddresses := []string{}
	var processedAddresses atomic.Uint32

	// Start goroutines to consume data from channels
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case participant, ok := <-participantsCh:
				if !ok {
					log.Debugw("collected valid database participants", "count", len(participants))
					return
				}
				participants = append(participants, participant)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case addr, ok := <-pendingAddressesCh:
				if !ok {
					log.Debugw("collected pending participants", "count", len(pendingAddresses))
					return
				}
				pendingAddresses = append(pendingAddresses, addr)
			}
		}
	}()

	// Progress update goroutine
	go func() {
		logTicker := time.NewTicker(2 * time.Second)
		defer logTicker.Stop()
		for {
			select {
			case <-ctx.Done():
				v.updateCensusProgress(ctx, censusID, 100*processedAddresses.Load()/uniqueAddressesCount)
				return
			case <-logTicker.C:
				processed := processedAddresses.Load()
				v.updateCensusProgress(ctx, censusID, 100*processed/uniqueAddressesCount)
				log.Debugw("census creation",
					"processed", processed,
					"total", uniqueAddressesCount,
					"progress", 100*processed/uniqueAddressesCount,
				)
			}
		}
	}()

	// Processing addresses
	for address := range addressMap {
		concurrencyLimit <- struct{}{}
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			defer func() { <-concurrencyLimit }() // Release semaphore

			user, err := v.db.UserByAddress(addr)
			if err != nil {
				if !errors.Is(err, mongo.ErrUserUnknown) {
					log.Warnw("error fetching user from database", "address", addr, "error", err)
				} else {
					pendingAddressesCh <- addr
				}
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
			processedAddresses.Add(1)
		}(address)
	}

	wg.Wait()
	close(participantsCh)
	close(pendingAddressesCh)
	close(concurrencyLimit)

	// Fetch the remaining users from the farcaster API
	count := 0
	log.Debugw("fetching users from farcaster", "count", len(pendingAddresses))
	for i := 0; i < len(pendingAddresses); i += neynar.MaxAddressesPerRequest {
		// Fetch the user data from the farcaster API
		ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		to := i + neynar.MaxAddressesPerRequest
		if to > len(pendingAddresses) {
			to = len(pendingAddresses)
		}
		log.Debugw("fetching users from farcaster", "from", i, "to", to)
		usersData, err := v.fcapi.UserDataByVerificationAddress(ctx2, pendingAddresses[i:to])
		if err != nil {
			if errors.Is(err, farcasterapi.ErrNoDataFound) {
				break
			}
			return nil, 0, err
		}
		for _, userData := range usersData {
			// Add or update the user on the database
			dbUser, err := v.db.User(userData.FID)
			if err != nil {
				if !errors.Is(err, mongo.ErrUserUnknown) {
					return nil, 0, err
				}
				if err := v.db.AddUser(
					userData.FID,
					userData.Username,
					userData.VerificationsAddresses,
					userData.Signers,
					userData.CustodyAddress,
					0,
				); err != nil {
					return nil, 0, err
				}
			} else {
				dbUser.Addresses = userData.VerificationsAddresses
				dbUser.Username = userData.Username
				dbUser.Signers = userData.Signers
				dbUser.CustodyAddress = userData.CustodyAddress
				if err := v.db.UpdateUser(dbUser); err != nil {
					return nil, 0, err
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
		processedAddresses.Add(uint32(to - i))
	}
	if len(pendingAddresses) > 0 {
		log.Infow("users found on farcaster", "count", count, "ratio", fmt.Sprintf("%.2f%%", 100*float64(count)/float64(len(pendingAddresses))))
	}
	return participants, uint32(len(addressMap)), nil
}

func ParseCSV(csvData []byte) ([][]string, error) {
	if len(csvData) == 0 {
		return nil, fmt.Errorf("empty csv")
	}
	csvType := "Weight-Balance"

	// get the first line of the csv to check if it's a POAP csv
	firstLine := strings.Split(string(csvData), "\n")[0]
	if strings.Contains(firstLine, POAP_CSV_HEADER) {
		csvType = "POAP"
	}
	var records [][]string
	log.Infow("parsing csv", "type", csvType)
	switch csvType {
	case "Weight-Balance":
		// Convert the byte slice to a reader
		r := csv.NewReader(strings.NewReader(string(csvData)))
		r.Comment = '#'
		r.TrimLeadingSpace = true // trim leading space of each field
		r.FieldsPerRecord = 2     // expect 2 fields per record
		count := 0
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
	case "POAP":
		// Convert the byte slice to a reader
		r := csv.NewReader(strings.NewReader(string(csvData)))
		r.Comment = '#'
		r.TrimLeadingSpace = true // trim leading space of each field
		r.FieldsPerRecord = 6     // expect 6 fields per record
		count := 0
		firstLine := true
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			if firstLine {
				firstLine = false
				continue
			}
			records = append(records, []string{record[1], "1"})
			count++
			if count > maxNumOfCsvRecords {
				return nil, fmt.Errorf("max number of records exceeded")
			}
		}

	}
	return records, nil
}

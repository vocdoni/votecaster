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
	"github.com/vocdoni/vote-frame/alfafrens"
	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/farcasterapi/neynar"
	"github.com/vocdoni/vote-frame/helpers"
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
	maxBatchParticipants   = 8000
	maxUsersNamesToReturn  = 10000

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
	// FrameCensusTypeNFT is a census created from the token holders of an NFT
	FrameCensusTypeNFT
	// FrameCensusTypeERC20 is a census created from the token holders of an ERC20
	FrameCensusTypeERC20
	// FrameCensusTypeAlfaFrensChannel is a census created from the users who follow a specific AlfaFrens Channel
	FrameCensusTypeAlfaFrensChannel
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
	FID      uint64   `json:"fid"`
}

// CreateCensus creates a new census from a list of participants.
func CreateCensus(cli *apiclient.HTTPclient, participants []*FarcasterParticipant,
	censusType FrameCensusType, progress chan int,
) (*CensusInfo, error) {
	censusList := api.CensusParticipants{}
	for _, p := range participants {
		voterID := state.NewFarcasterVoterID(p.PubKey, p.FID)
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
	// Add the participants to the census, if the number of participants is less
	// than the maxBatchParticipants add them all at once, otherwise split them
	// into batches
	if len(censusList.Participants) < maxBatchParticipants {
		if err := cli.CensusAddParticipants(censusID, &censusList); err != nil {
			return nil, err
		}
	} else {
		log.Debugw("max batch participants exceeded", "participants", len(censusList.Participants))
		// Split the participants into batches
		idxBatch := 0
		for i := 0; i < len(censusList.Participants); i += maxBatchParticipants {
			to := i + maxBatchParticipants
			if to > len(censusList.Participants) {
				to = len(censusList.Participants)
			}
			batch := api.CensusParticipants{Participants: censusList.Participants[i:to]}
			if err := cli.CensusAddParticipants(censusID, &batch); err != nil {
				return nil, err
			}
			idxBatch++
			log.Debugw("census batch added, sleeping 100ms...", "index", idxBatch, "from", i, "to", to)
			time.Sleep(100 * time.Millisecond)
			if progress != nil {
				progress <- 100 * i / len(censusList.Participants)
			}
		}
	}
	// increase the http client timeout to 5 minutes to allow to publish large
	// censuses
	cli.SetTimeout(5 * time.Minute)
	root, url, err := cli.CensusPublish(censusID)
	if err != nil {
		log.Warnw("failed to publish census", "censusID", censusID, "error", err, "participants", len(censusList.Participants))
		return nil, err
	}
	cli.SetTimeout(apiclient.DefaultTimeout)

	size, err := cli.CensusSize(censusID)
	if err != nil {
		log.Warnw("failed to get census size", "censusID", censusID, "error", err, "participants", len(censusList.Participants))
		return nil, err
	}
	return &CensusInfo{
		Root: root,
		Url:  url,
		Size: size,
		Type: censusType,
	}, nil
}

// censusFromDatabaseByElectionID retrieves a census from the database by its election ID.
func (v *vocdoniHandler) censusFromDatabaseByElectionID(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	eID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return err
	}
	census, err := v.db.CensusFromElection(eID)
	if err != nil {
		return ctx.Send(nil, http.StatusNotFound)
	}
	data, err := json.Marshal(census)
	if err != nil {
		return err
	}
	return ctx.Send(data, http.StatusOK)
}

// censusFromDatabaseByRoot retrieves a census from the database by its root.
func (v *vocdoniHandler) censusFromDatabaseByRoot(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	root, err := hex.DecodeString(ctx.URLParam("root"))
	if err != nil {
		return err
	}
	census, err := v.db.CensusFromRoot(root)
	if err != nil {
		return ctx.Send(nil, http.StatusNotFound)
	}
	data, err := json.Marshal(census)
	if err != nil {
		return err
	}
	return ctx.Send(data, http.StatusOK)
}

// censusCSV creates a new census from a CSV file containing Ethereum addresses and weights.
// It builds the census async and returns the census ID.
func (v *vocdoniHandler) censusCSV(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	censusID, err := v.cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return err
	}
	v.backgroundQueue.Store(censusID.String(), CensusInfo{})
	userFID, err := v.db.UserFromAuthToken(msg.AuthToken)
	if err != nil {
		return fmt.Errorf("cannot get user from auth token: %w", err)
	}
	if err := v.db.AddCensus(censusID, userFID); err != nil {
		return fmt.Errorf("cannot add census to database: %w", err)
	}
	totalCSVaddresses := uint32(0)
	go func() {
		startTime := time.Now()
		log.Debugw("building census from csv", "censusID", censusID)
		var participants []*FarcasterParticipant
		var err error
		v.trackStepProgress(censusID, 1, 2, func(progress chan int) {
			participants, totalCSVaddresses, err = v.farcasterCensusFromEthereumCSV(msg.Data, progress)
		})
		if err != nil {
			log.Warnw("failed to build census from ethereum csv", "err", err.Error())
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		var ci *CensusInfo
		v.trackStepProgress(censusID, 2, 2, func(progress chan int) {
			ci, err = CreateCensus(v.cli, participants, FrameCensusTypeCSV, progress)
		})
		if err != nil {
			log.Errorw(err, "failed to create census")
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		// since each participant can have multiple signers, we need to get the unique usernames
		uniqueParticipantsMap := make(map[string]*big.Int)
		totalWeight := new(big.Int).SetUint64(0)
		for _, p := range participants {
			if _, ok := uniqueParticipantsMap[p.Username]; ok {
				// if the username is already in the map, continue
				continue
			}
			uniqueParticipantsMap[p.Username] = p.Weight
			totalWeight.Add(totalWeight, p.Weight)
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
			"totalWeight", totalWeight.String(),
			"duration", time.Since(startTime),
			"fromTotalAddresses", totalCSVaddresses)

		// store the census info in the map
		v.backgroundQueue.Store(censusID.String(), *ci)

		// add participants to the census in the database
		if err := v.db.AddParticipantsToCensus(censusID, uniqueParticipantsMap, ci.FromTotalAddresses, totalWeight, ci.Url); err != nil {
			log.Errorw(err, fmt.Sprintf("failed to add participants to census %s", censusID.String()))
		}
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
	exists, err := v.fcapi.ChannelExists(ctx.Request.Context(), channelID)
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
func (v *vocdoniHandler) censusChannel(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// extract userFID from auth token
	userFID, err := v.db.UserFromAuthToken(msg.AuthToken)
	if err != nil {
		return fmt.Errorf("cannot get user from auth token: %w", err)
	}
	// check if channelID is provided, it is required so if it's not provided
	// return a BadRequest error
	channelID := ctx.URLParam("channelID")
	if channelID == "" {
		return ctx.Send([]byte("channelID is required"), http.StatusBadRequest)
	}
	// check if the channel exists, if not return a NotFound error. If something
	// fails when checking the channel existence, return the error.
	exists, err := v.fcapi.ChannelExists(ctx.Request.Context(), channelID)
	if err != nil {
		return err
	}
	if !exists {
		return ctx.Send([]byte("channel not found"), http.StatusNotFound)
	}
	// create a censusID for the queue and store into it
	data, err := v.censusWarpcastChannel(channelID, userFID)
	if err != nil {
		log.Warnf("error creating census for the chanel: %s: %v", channelID, err)
		return ctx.Send([]byte("error creating channel census"), http.StatusInternalServerError)
	}
	return ctx.Send(data, http.StatusOK)
}

func (v *vocdoniHandler) censusFollowersHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	req := struct {
		Profile FarcasterProfile `json:"profile"`
	}{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return err
	}
	// check if userFid is provided, it is required so if it's not provided
	// return a BadRequest error
	strUserFid := ctx.URLParam("userFid")
	if strUserFid == "" {
		return ctx.Send([]byte("userFid is required"), http.StatusBadRequest)
	}
	userFID, err := strconv.ParseUint(strUserFid, 10, 64)
	if err != nil {
		return ctx.Send([]byte("invalid userFid"), http.StatusBadRequest)
	}
	// create the census from the followers of the user and return the data as
	// response
	data, err := v.censusFollowers(userFID)
	if err != nil {
		return err
	}
	return ctx.Send(data, http.StatusOK)
}

// censusAlfafrensChannelHandler creates a new census from the users who follow the AlfaFrens channel of the user
// making the request.
func (v *vocdoniHandler) censusAlfafrensChannelHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// extract userFID from auth token
	userFID, err := v.db.UserFromAuthToken(msg.AuthToken)
	if err != nil {
		return fmt.Errorf("cannot get user from auth token: %w", err)
	}

	// create a censusID for the queue and store into it
	censusID, err := v.cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return err
	}
	data, err := v.censusAlfafrensChannel(censusID, userFID)
	if err != nil {
		log.Warnf("error creating census for alfafrens channel of user: %d: %v", userFID, err)
		return ctx.Send([]byte("error creating channel census"), http.StatusInternalServerError)
	}
	return ctx.Send(data, http.StatusOK)
}

// censusCommunity creates a new census from a community. The census of the
// community can be of type channel, NFT, or ERC20. If the community is a
// channel, the census is created from the users who follow the channel, and
// the process is async. If the community is an NFT or ERC20, the census is
// created from the token holders of the token addresses in the community, using
// the AirStack API. The process is sync and the census is created in the same
// request. The census is created from the participants and the progress is
// updated in the queue.
func (v *vocdoniHandler) censusCommunity(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// extract userFID from auth token
	userFID, err := v.db.UserFromAuthToken(msg.AuthToken)
	if err != nil {
		return fmt.Errorf("cannot get user from auth token: %w", err)
	}
	req := struct {
		CommunityID uint64 `json:"communityID"`
	}{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// get the community from the database
	community, err := v.db.Community(req.CommunityID)
	if err != nil {
		return ctx.Send([]byte("error getting community"), http.StatusInternalServerError)
	}
	if community == nil {
		return ctx.Send([]byte("community not found"), http.StatusNotFound)
	}

	// check if the user is admin of the community
	if !v.db.IsCommunityAdmin(userFID, req.CommunityID) {
		return fmt.Errorf("user is not an admin of the community")
	}
	// check the type to create it from the correct source (channel, airstak
	// (nft/erc20) or user followers) and in the correct way (async or sync)
	switch community.Census.Type {
	case mongo.TypeCommunityCensusFollowers:
		// if the census type is followers, create the census from the users who
		// follow the user, the process is async so return add the censusID to the
		// queue and return it to the client
		data, err := v.censusFollowers(userFID)
		if err != nil {
			log.Warnf("error creating census for the user: %d: %v", userFID, err)
			return ctx.Send([]byte("error creating user followers census"), http.StatusInternalServerError)
		}
		return ctx.Send(data, http.StatusOK)
	case mongo.TypeCommunityCensusChannel:
		// if the census type is a channel, create the census from the users who
		// follow the channel, the process is async so return add the censusID
		// to the queue and return it to the client
		data, err := v.censusWarpcastChannel(community.Census.Channel, userFID)
		if err != nil {
			log.Warnf("error creating census for the chanel: %s: %v", community.Census.Channel, err)
			return ctx.Send([]byte("error creating channel census"), http.StatusInternalServerError)
		}
		return ctx.Send(data, http.StatusOK)
	case mongo.TypeCommunityCensusNFT, mongo.TypeCommunityCensusERC20:
		// if the census type is not a channel, the type is NFT or ERC20, so create
		// the census sync and from the token holders from airstack
		censusAddresses := []*CensusToken{}
		for _, addr := range community.Census.Addresses {
			censusAddresses = append(censusAddresses, &CensusToken{
				Address:    addr.Address,
				Blockchain: addr.Blockchain,
			})
		}
		// check valid token
		if err := v.checkTokens(censusAddresses); err != nil {
			return err
		}
		// convert the census type to the correct type for the CreateCensus function
		var censusType int
		switch community.Census.Type {
		case mongo.TypeCommunityCensusNFT:
			// set the census type to NFT
			censusType = NFTtype
		case mongo.TypeCommunityCensusERC20:
			// set the census type to ERC20 and check the number of tokens is 1
			censusType = ERC20type
			if len(censusAddresses) != 1 {
				return fmt.Errorf("erc20 census must have only one token address")
			}
		}
		// create the census from the token holders
		data, err := v.censusTokenAirstack(censusAddresses, censusType, userFID)
		if err != nil {
			return fmt.Errorf("cannot create erc20/nft based census: %w", err)
		}
		return ctx.Send(data, http.StatusOK)
	default:
		return ctx.Send([]byte("invalid census type"), http.StatusBadRequest)
	}
}

// censusQueueInfo returns the status of the census creation process.
// Returns 204 if the census is not yet ready or not found.
func (v *vocdoniHandler) censusQueueInfo(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	var censusID types.HexBytes
	var err error
	censusID, err = hex.DecodeString(ctx.URLParam("censusID"))
	if err != nil {
		return err
	}
	iCensusInfo, ok := v.backgroundQueue.Load(censusID.String())
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
	if len(censusInfo.Usernames) > maxUsersNamesToReturn {
		censusInfo.Usernames = nil
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
	if err = v.db.SetRootForCensus(censusID, censusInfo.Root); err != nil {
		return fmt.Errorf("cannot set root for census: %w", err)
	}
	data, err := json.Marshal(censusInfo)
	if err != nil {
		return err
	}
	return ctx.Send(data, http.StatusOK)
}

const (
	// NFTtype represents an NFT token type used in the census
	NFTtype = iota
	// ERC20type representing an ERC20 token type used in the census
	ERC20type

	MAXNFTTokens   = 3
	MAXERC20Tokens = 1
)

func (v *vocdoniHandler) censusBlockchainsAirstack(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	data, err := json.Marshal(map[string][]string{"blockchains": v.airstack.Blockchains()})
	if err != nil {
		return err
	}
	return ctx.Send(data, http.StatusOK)
}

func (v *vocdoniHandler) censusTokenNFTAirstack(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	req := &CensusTokensRequest{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return err
	}
	// check valid tokens length
	if len(req.Tokens) > MAXNFTTokens || len(req.Tokens) == 0 {
		return fmt.Errorf("invalid number of NFT tokens, bounds between 1 and 3")
	}

	// extract userFID from auth token
	userFID, err := v.db.UserFromAuthToken(msg.AuthToken)
	if err != nil {
		return fmt.Errorf("cannot get user from auth token: %w", err)
	}

	// check valid tokens
	if err := v.checkTokens(req.Tokens); err != nil {
		return err
	}

	data, err := v.censusTokenAirstack(req.Tokens, NFTtype, userFID)
	if err != nil {
		return fmt.Errorf("cannot create nft census: %w", err)
	}
	return ctx.Send(data, http.StatusOK)
}

func (v *vocdoniHandler) checkTokens(tokens []*CensusToken) error {
	for _, token := range tokens {
		if len(token.Address) == 0 {
			return fmt.Errorf("invalid token information: %v", token)
		}
		ok := false
		for _, bk := range v.airstack.Blockchains() {
			if bk == token.Blockchain {
				ok = true
			}
		}
		if !ok {
			return fmt.Errorf("invalid blockchain for token %s provided", token.Address)
		}
		// check max holders
		if holders, err := v.airstack.NumHoldersByTokenAnkrAPI(token.Address, token.Blockchain); err != nil {
			log.Warnf("cannot get holders for token %s: %s", token.Address, err)
		} else if holders > v.airstack.MaxHolders() {
			// check whitelist
			if _, ok := v.airstack.TokenWhitelist()[token.Address]; ok {
				continue
			}
			return fmt.Errorf("token %s has too many holders: %d, maximum allowed is %d", token.Address, holders, v.airstack.MaxHolders())
		}
	}
	return nil
}

func (v *vocdoniHandler) censusTokenERC20Airstack(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	req := &CensusTokensRequest{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return err
	}
	if len(req.Tokens) != MAXERC20Tokens {
		return fmt.Errorf("invalid number of ERC20 tokens, must be %d", MAXERC20Tokens)
	}

	// extract userFID from auth token
	userFID, err := v.db.UserFromAuthToken(msg.AuthToken)
	if err != nil {
		return fmt.Errorf("cannot get user from auth token: %w", err)
	}

	// check valid token
	if err := v.checkTokens(req.Tokens); err != nil {
		return err
	}

	data, err := v.censusTokenAirstack(req.Tokens, ERC20type, userFID)
	if err != nil {
		return fmt.Errorf("cannot create erc20 census: %w", err)
	}
	return ctx.Send(data, http.StatusOK)
}

func (v *vocdoniHandler) censusTokenAirstack(tokens []*CensusToken, tokenType int, createdByFID uint64) ([]byte, error) {
	if v.airstack == nil {
		return nil, fmt.Errorf("airstack service not available")
	}
	censusID, err := v.cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return nil, err
	}
	if err := v.db.AddCensus(censusID, createdByFID); err != nil {
		return nil, fmt.Errorf("cannot add census to database: %w", err)
	}
	v.backgroundQueue.Store(censusID.String(), CensusInfo{})
	go func() {
		startTime := time.Now()
		log.Debugw("building Airstack based census", "censusID", censusID)
		// get holders for each token
		var holders [][]string
		var err error
		v.trackStepProgress(censusID, 1, 3, func(progress chan int) {
			holders, err = v.getTokenHoldersFromAirstack(tokens, censusID, progress)
		})
		if err != nil {
			log.Warnw("failed to build census from NFT, cannot get holders", "err", err.Error())
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		// create census from token holders
		var participants []*FarcasterParticipant
		v.trackStepProgress(censusID, 2, 3, func(progress chan int) {
			participants, _, err = v.processCensusRecords(holders, progress)
		})
		if err != nil {
			log.Warnw("failed to build census from NFT", "err", err.Error())
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		var ci *CensusInfo
		v.trackStepProgress(censusID, 3, 3, func(progress chan int) {
			if tokenType == ERC20type {
				ci, err = CreateCensus(v.cli, participants, FrameCensusTypeERC20, progress)
			} else if tokenType == NFTtype {
				ci, err = CreateCensus(v.cli, participants, FrameCensusTypeNFT, progress)
			}
		})
		if err != nil {
			log.Errorw(err, "failed to create census")
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}

		// since each participant can have multiple signers, we need to get the unique usernames
		uniqueParticipantsMap := make(map[string]*big.Int)
		totalWeight := new(big.Int).SetUint64(0)
		for _, p := range participants {
			if _, ok := uniqueParticipantsMap[p.Username]; ok {
				// if the username is already in the map, continue
				continue
			}
			uniqueParticipantsMap[p.Username] = p.Weight
			totalWeight.Add(totalWeight, p.Weight)
		}
		uniqueParticipants := []string{}
		for k := range uniqueParticipantsMap {
			uniqueParticipants = append(uniqueParticipants, k)
		}
		ci.Usernames = uniqueParticipants
		ci.FromTotalAddresses = uint32(len(holders))
		log.Infow("census created from Airstack",
			"censusID", censusID.String(),
			"size", len(ci.Usernames),
			"totalWeight", totalWeight.String(),
			"duration", time.Since(startTime),
			"totalAddresses", ci.FromTotalAddresses,
			"participants", len(ci.Usernames),
		)
		// store the census info in the memory map
		v.backgroundQueue.Store(censusID.String(), *ci)
		// add participants to the census in the database
		if err := v.db.AddParticipantsToCensus(
			censusID,
			uniqueParticipantsMap,
			ci.FromTotalAddresses,
			totalWeight,
			ci.Url,
		); err != nil {
			log.Errorw(err, fmt.Sprintf("failed to add participants to census %s", censusID.String()))
		}
	}()
	data, err := json.Marshal(map[string]string{"censusId": censusID.String()})
	if err != nil {
		return nil, err
	}
	return data, nil
}

// censusWarpcastChannel helper method creates a new census from a Warpcast
// Channel. The process is async and returns the json encoded censusID. It
// updates the progress in the queue and the result when it's ready.
func (v *vocdoniHandler) censusWarpcastChannel(channelID string, authorFID uint64) ([]byte, error) {
	// create a censusID for the queue and store into it
	censusID, err := v.cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return nil, err
	}
	v.backgroundQueue.Store(censusID.String(), CensusInfo{})
	if err := v.db.AddCensus(censusID, authorFID); err != nil {
		return nil, fmt.Errorf("cannot add census to database: %w", err)
	}
	// run a goroutine to create the census, update the queue with the progress,
	// and update the queue result when it's ready
	go func() {
		internalCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var err error
		// get the fids of the users in the channel from neynar farcaster API, if
		// the channel does not exist, return a NotFound error
		var users []uint64
		v.trackStepProgress(censusID, 1, 3, func(progress chan int) {
			users, err = v.fcapi.ChannelFIDs(internalCtx, channelID, progress)
		})
		if err != nil {
			log.Errorw(err, "failed to get channel fids from farcaster API")
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		if len(users) == 0 {
			log.Errorw(fmt.Errorf("no valid participants found for the channel %s", channelID), "")
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: "no valid participants found for the channel"})
			return
		}
		// create the participants from the database users using the fids
		var participants []*FarcasterParticipant
		v.trackStepProgress(censusID, 2, 3, func(progress chan int) {
			participants = v.farcasterCensusFromFids(users, progress)
		})
		if len(participants) == 0 {
			log.Errorw(fmt.Errorf("no valid participant signers found for the channel %s", channelID), "")
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: "no valid participant signers found for the channel"})
			return
		}
		// create the census from the participants
		var censusInfo *CensusInfo
		v.trackStepProgress(censusID, 3, 3, func(progress chan int) {
			censusInfo, err = CreateCensus(v.cli, participants, FrameCensusTypeChannelGated, progress)
		})
		if err != nil {
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		uniqueParticipantsMap := make(map[string]*big.Int)
		for _, p := range participants {
			if _, ok := uniqueParticipantsMap[p.Username]; !ok {
				uniqueParticipantsMap[p.Username] = new(big.Int).SetUint64(1)
			}
		}
		// only return the username list if it's less than the maxUsersNamesToReturn
		if len(uniqueParticipantsMap) < maxUsersNamesToReturn {
			for username := range uniqueParticipantsMap {
				censusInfo.Usernames = append(censusInfo.Usernames, username)
			}
		}
		censusInfo.FromTotalAddresses = uint32(len(users))
		v.backgroundQueue.Store(censusID.String(), *censusInfo)
		// add participants to the census in the database
		if err := v.db.AddParticipantsToCensus(
			censusID,
			uniqueParticipantsMap,
			censusInfo.FromTotalAddresses,
			new(big.Int).SetUint64(uint64(len(uniqueParticipantsMap))),
			censusInfo.Url,
		); err != nil {
			log.Errorw(err, fmt.Sprintf("failed to add participants to census %s", censusID.String()))
		}
		log.Infow("census created from channel",
			"channelID", channelID,
			"participants", len(censusInfo.Usernames))
	}()
	// return the censusID to the client
	return json.Marshal(map[string]string{"censusId": censusID.String()})
}

// censusFollowers helper creates a new census from the followers of a user.
// The process is async and returns the json encoded censusID. It updates the
// progress in the queue and the result when it's ready. If something fails
// during the process, it returns an error or the error is stored in the queue
// if it's async.
func (v *vocdoniHandler) censusFollowers(userFID uint64) ([]byte, error) {
	// create a censusID for the queue and store into it
	censusID, err := v.cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return nil, err
	}
	// store the censusID in the database and the queue
	if err := v.db.AddCensus(censusID, userFID); err != nil {
		return nil, fmt.Errorf("cannot add census to database: %w", err)
	}
	v.backgroundQueue.Store(censusID.String(), CensusInfo{})
	// run a goroutine to create the census, update the queue with the progress,
	// and update the queue result when it's ready
	go func() {
		internalCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		users, err := v.fcapi.UserFollowers(internalCtx, userFID)
		if err != nil {
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		// include poll author in the census
		users = append(users, userFID)
		// create the participants from the database users using the fids
		var participants []*FarcasterParticipant
		v.trackStepProgress(censusID, 1, 2, func(progress chan int) {
			participants = v.farcasterCensusFromFids(users, progress)
		})
		if len(participants) == 0 {
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: "no valid participants"})
			return
		}
		// create the census from the participants
		var censusInfo *CensusInfo
		v.trackStepProgress(censusID, 2, 2, func(progress chan int) {
			censusInfo, err = CreateCensus(v.cli, participants, FrameCensusTypeFollowers, progress)
		})
		if err != nil {
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		uniqueParticipantsMap := make(map[string]*big.Int)
		totalWeight := new(big.Int).SetUint64(0)
		for _, p := range participants {
			if _, ok := uniqueParticipantsMap[p.Username]; ok {
				// if the username is already in the map, continue
				continue
			}
			uniqueParticipantsMap[p.Username] = p.Weight
			totalWeight.Add(totalWeight, p.Weight)
		}
		// only return the username list if it's less than the maxUsersNamesToReturn
		if len(uniqueParticipantsMap) < maxUsersNamesToReturn {
			for u := range uniqueParticipantsMap {
				censusInfo.Usernames = append(censusInfo.Usernames, u)
			}
		}
		// store the census info in the database
		if err := v.db.AddParticipantsToCensus(
			censusID,
			uniqueParticipantsMap,
			uint32(len(users)),
			totalWeight,
			censusInfo.Url,
		); err != nil {
			log.Errorw(err, fmt.Sprintf("failed to add participants to census %s", censusID.String()))
		}

		censusInfo.FromTotalAddresses = uint32(len(users))
		v.backgroundQueue.Store(censusID.String(), *censusInfo)
		log.Infow("census created from user followers",
			"fid", userFID,
			"participants", len(censusInfo.Usernames))
	}()
	// return the censusID to the client
	return json.Marshal(map[string]string{"censusId": censusID.String()})
}

// censusAlfafrensChannel creates a new census from an AlfaFrens Channel.
func (v *vocdoniHandler) censusAlfafrensChannel(censusID types.HexBytes, ownerFID uint64) ([]byte, error) {
	v.backgroundQueue.Store(censusID.String(), CensusInfo{})
	if err := v.db.AddCensus(censusID, ownerFID); err != nil {
		return nil, fmt.Errorf("cannot add census to database: %w", err)
	}
	// get the channel address from the alfafrens API
	channelAddr, err := alfafrens.ChannelByFid(ownerFID)
	if err != nil {
		return nil, fmt.Errorf("cannot get alfafrens channel address for user %d: %w", ownerFID, err)
	}
	// run a goroutine to create the census, update the queue with the progress,
	// and update the queue result when it's ready
	go func() {
		var err error
		var users []uint64
		// get the fids of the users in the channel from neynar farcaster API, if
		// the channel does not exist, return a NotFound error
		v.trackStepProgress(censusID, 1, 3, func(progress chan int) {
			progress <- 10
			users, err = alfafrens.ChannelFids(channelAddr)
			progress <- 100
		})
		if err != nil {
			log.Errorw(err, "failed to get channel fids from farcaster API")
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		if len(users) == 0 {
			log.Errorw(fmt.Errorf("no valid participants found for alfafrens channel %s", channelAddr.String()), "")
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: "no valid participants found for the channel"})
			return
		}
		// create the participants from the database users using the fids
		var participants []*FarcasterParticipant
		v.trackStepProgress(censusID, 1, 2, func(progress chan int) {
			participants = v.farcasterCensusFromFids(users, progress)
		})
		if len(participants) == 0 {
			log.Errorw(fmt.Errorf("no valid participant signers found for alfafrens channel %s", channelAddr.String()), "")
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: "no valid participant signers found for the channel"})
			return
		}
		// create the census from the participants
		var censusInfo *CensusInfo
		v.trackStepProgress(censusID, 2, 2, func(progress chan int) {
			censusInfo, err = CreateCensus(v.cli, participants, FrameCensusTypeAlfaFrensChannel, progress)
		})
		if err != nil {
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		uniqueParticipantsMap := make(map[string]*big.Int)
		for _, p := range participants {
			if _, ok := uniqueParticipantsMap[p.Username]; !ok {
				uniqueParticipantsMap[p.Username] = new(big.Int).SetUint64(1)
			}
		}
		for username := range uniqueParticipantsMap {
			censusInfo.Usernames = append(censusInfo.Usernames, username)
		}
		censusInfo.FromTotalAddresses = uint32(len(users))
		v.backgroundQueue.Store(censusID.String(), *censusInfo)
		// add participants to the census in the database
		if err := v.db.AddParticipantsToCensus(
			censusID,
			uniqueParticipantsMap,
			censusInfo.FromTotalAddresses,
			new(big.Int).SetUint64(uint64(len(uniqueParticipantsMap))),
			censusInfo.Url,
		); err != nil {
			log.Errorw(err, fmt.Sprintf("failed to add participants to census %s", censusID.String()))
		}
		log.Infow("census created for alfafrens channel",
			"channelID", channelAddr.String(),
			"participants", len(censusInfo.Usernames))
	}()
	// return the censusID to the client
	return json.Marshal(map[string]string{"censusId": censusID.String()})
}

func (v *vocdoniHandler) checkERC20ContractHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// TODO: It should receive CheckCensusSource instance
	return ctx.Send([]byte("ok"), http.StatusOK)
}
func (v *vocdoniHandler) checkNFTContractHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// TODO: It should receive CheckCensusSource instance
	return ctx.Send([]byte("ok"), http.StatusOK)
}

// getTokenHoldersFromAirstack retuns a list of token holders ans their balances given a list of tokens
// It fetches the information of the token holders by consuming the Airstack API
// The holders list balances is truncated to the number of decimals of the token (if any).
func (v *vocdoniHandler) getTokenHoldersFromAirstack(
	tokens []*CensusToken, censusID types.HexBytes, progress chan int,
) ([][]string, error) {
	holders := make([][]string, 0)
	processedTokens := 0
	totalTokens := len(tokens)
	totalHolders := 0
	for _, token := range tokens {
		tokenAddress := common.HexToAddress(token.Address)
		// Get the number of decimals for the token
		decimals, err := v.airstack.TokenDecimalsByToken(token.Address, token.Blockchain)
		if err != nil {
			log.Warnw("failed to fetch token details", "token", token.Address, "error", err)
		}

		tokenHolders, err := v.airstack.TokenBalances(tokenAddress, token.Blockchain)
		if err != nil {
			log.Warnw("failed to create census for token %s: %v", token.Address, err)
			v.backgroundQueue.Store(censusID.String(), CensusInfo{
				Error: fmt.Sprintf("cannot get token %s details: %v", token.Address, err),
			})
			return nil, err
		}

		for _, tokenHolder := range tokenHolders {
			holders = append(holders, []string{tokenHolder.Address.String(), helpers.TruncateDecimals(tokenHolder.Balance, uint32(decimals)).String()})
		}

		totalHolders += len(tokenHolders)

		// update the progress if the progress channel is provided
		// since the response time of GetTokenBalances is unknown, because it dependends on the total number
		// holders and cannot be known beforehand, update at least the progress between tokens
		if progress != nil {
			if currentProcessedTokens := processedTokens + 1; totalTokens >= int(currentProcessedTokens) {
				currentProgress := 100 * currentProcessedTokens / totalTokens
				progress <- currentProgress
			}
		}
	}
	log.Debugf("airstack total holders found: %d", totalHolders)
	return holders, nil
}

func (v *vocdoniHandler) farcasterCensusFromEthereumCSV(csv []byte, progress chan int) ([]*FarcasterParticipant, uint32, error) {
	records, err := ParseCSV(csv)
	if err != nil {
		return nil, 0, err
	}
	return v.processCensusRecords(records, progress)
}

// farcasterCensusFromFids creates a list of Farcaster participants from a list
// of FIDs. It queries the database to get the users signer keys and creates the
// participants from them. It returns the list of participants and a map of the
// FIDs that failed to get the users from the database or decoding the keys.
func (v *vocdoniHandler) farcasterCensusFromFids(fids []uint64, progress chan int) []*FarcasterParticipant {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// get participants from the users fids, quering the database and to get the
	// users public keys
	totalFids := len(fids)
	var wg sync.WaitGroup
	participants := []*FarcasterParticipant{}
	participantsCh := make(chan *FarcasterParticipant)
	concurrencyLimit := make(chan struct{}, 10)
	var processedFids atomic.Uint32
	// Start goroutines to consume data from channel
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
	// run database queries concurrently
	for i, fid := range fids {
		concurrencyLimit <- struct{}{}
		wg.Add(1)
		go func(idx int, fid uint64) {
			defer wg.Done()
			defer func() { <-concurrencyLimit }()
			// get the user from the database
			user, err := v.db.User(fid)
			if err != nil {
				if !errors.Is(err, mongo.ErrUserUnknown) {
					log.Warnw("error fetching user from database", "fid", fid, "error", err)
				}
				return
			}
			// create a participant for each signer of the user
			for _, signer := range user.Signers {
				signerBytes, err := hex.DecodeString(strings.TrimPrefix(signer, "0x"))
				if err != nil {
					log.Warnw("error decoding signer", "signer", signer, "err", err)
					return
				}
				participantsCh <- &FarcasterParticipant{
					PubKey:   signerBytes,
					Weight:   big.NewInt(1),
					Username: user.Username,
					FID:      fid,
				}
			}
			// update the progress if the progress channel is provided
			if progress != nil {
				// ensure the progress is updated only if the current fid is the
				// last one processed (its index is greater than the current
				// processed fids)
				if currentProcessedFids := processedFids.Add(1); uint32(idx) > currentProcessedFids {
					currentProgress := 100 * idx / totalFids
					progress <- currentProgress
				}
			}
		}(i, fid)
	}
	wg.Wait()
	close(participantsCh)
	close(concurrencyLimit)
	return participants
}

// trackStepProgress tracks the progress of a step in the census creation
// process. It updates the census progress in the queue. This method must
// envolve the steps actions secuentally in a goroutine to avoid blocking the
// main thread, while the progress is tracked in the main. This method creates
// its own channel and goroutine to track the progress of the current step. This
// channel is provided to the action function to update the progress and it's
// closed when the action function finishes. The action function is expected to
// update the progress channel with the progress of the step.
func (v *vocdoniHandler) trackStepProgress(censusID types.HexBytes, step, totalSteps int, action func(chan int)) {
	progress := make(chan int)
	chanClosed := atomic.Bool{}
	go func() {
		for {
			p, ok := <-progress
			if !ok {
				chanClosed.Store(true)
				return
			}
			censusInfo, ok := v.backgroundQueue.Load(censusID.String())
			if !ok {
				return
			}
			ci, ok := censusInfo.(CensusInfo)
			if !ok || ci.Root != nil {
				return
			}
			// calc partial progress of current step
			stepIndex := uint32(step - 1)
			partialStep := 100 / uint32(totalSteps)
			stepProgress := stepIndex*partialStep + uint32(p)/uint32(totalSteps)
			// update the census progress
			ci.Progress = stepProgress
			v.backgroundQueue.Store(censusID.String(), ci)
		}
	}()
	action(progress)
	if !chanClosed.Load() {
		close(progress)
	}
}

// processRecord processes a single record of a plain-text census and returns the corresponding Farcaster participants.
// The record is expected to be a string containing the address and the weight.
// Returns the list of participants and the total number of unique addresses available in the records.
func (v *vocdoniHandler) processCensusRecords(records [][]string, progress chan int) ([]*FarcasterParticipant, uint32, error) {
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
		// If the weight is not provided, set it to 1
		if weightRecord == "" {
			weightRecord = "1"
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
				return
			case <-logTicker.C:
				if progress == nil {
					return
				}
				progress <- int(100 * processedAddresses.Load() / uniqueAddressesCount)
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

			// find the addres on the map to get the weight
			// the weight is the sum of the weights of all the addresses of the user
			weight := new(big.Int).SetUint64(0)
			for _, addr := range user.Addresses {
				weightAddress, ok := addressMap[common.HexToAddress(addr).Hex()]
				if ok {
					weight = weight.Add(weight, weightAddress)
				}
			}

			for _, signer := range user.Signers {
				signerBytes, err := hex.DecodeString(strings.TrimPrefix(signer, "0x"))
				if err != nil {
					log.Warnw("error decoding signer", "signer", signer, "err", err)
					continue
				}
				participantsCh <- &FarcasterParticipant{
					PubKey:   signerBytes,
					Weight:   weight,
					Username: user.Username,
					FID:      user.UserID,
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
			log.Errorw(err, "error fetching users from Neynar API")
		}
		for _, userData := range usersData {
			// Add or update the user on the database
			dbUser, err := v.db.User(userData.FID)
			if err != nil {
				log.Debugw("adding new user to database", "fid", userData.FID)
				if err := v.db.AddUser(
					userData.FID,
					userData.Username,
					userData.Displayname,
					userData.VerificationsAddresses,
					userData.Signers,
					userData.CustodyAddress,
					0,
				); err != nil {
					return nil, 0, err
				}
			} else {
				log.Debugw("updating user on database", "fid", userData.FID)
				dbUser.Addresses = userData.VerificationsAddresses
				dbUser.Username = userData.Username
				dbUser.Signers = userData.Signers
				dbUser.CustodyAddress = userData.CustodyAddress
				if err := v.db.UpdateUser(dbUser); err != nil {
					return nil, 0, err
				}
			}
			// find the addres on the map to get the weight
			// the weight is the sum of the weights of all the addresses of the user
			weight := new(big.Int).SetUint64(0)
			for _, addr := range userData.VerificationsAddresses {
				weightAddress, ok := addressMap[common.HexToAddress(addr).Hex()]
				if ok {
					weight = weight.Add(weight, weightAddress)
				}
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
					FID:      userData.FID,
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

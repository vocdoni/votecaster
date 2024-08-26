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
	Root                      types.HexBytes `json:"root"`
	Url                       string         `json:"uri"`
	Size                      uint64         `json:"size"`
	Usernames                 []string       `json:"usernames,omitempty"`
	FromTotalAddresses        uint32         `json:"fromTotalAddresses,omitempty"`
	FarcasterParticipantCount uint32         `json:"farcasterParticipantCount,omitempty"`

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
	PubKey      []byte   `json:"pubkey"`
	Weight      *big.Int `json:"weight"`
	Username    string   `json:"username"`
	FID         uint64   `json:"fid"`
	Delegations uint32   `json:"delegations"`
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
		totalParticipants := uint32(0) // including delegations
		for _, p := range participants {
			if _, ok := uniqueParticipantsMap[p.Username]; ok {
				// if the username is already in the map, continue
				continue
			}
			uniqueParticipantsMap[p.Username] = p.Weight
			totalWeight.Add(totalWeight, p.Weight)
			totalParticipants += p.Delegations + 1
		}
		uniqueParticipants := []string{}
		for k := range uniqueParticipantsMap {
			uniqueParticipants = append(uniqueParticipants, k)
		}
		ci.Usernames = uniqueParticipants
		ci.FarcasterParticipantCount = uint32(len(uniqueParticipants))
		ci.FromTotalAddresses = totalCSVaddresses
		log.Infow("census created from CSV",
			"censusID", censusID.String(),
			"size", len(ci.Usernames),
			"totalWeight", totalWeight.String(),
			"duration", time.Since(startTime),
			"fromTotalAddresses", totalCSVaddresses,
			"fromTotalParticipants", totalParticipants,
		)

		// store the census info in the map
		v.backgroundQueue.Store(censusID.String(), *ci)

		// add participants to the census in the database
		if err := v.db.AddParticipantsToCensus(censusID, uniqueParticipantsMap, ci.FromTotalAddresses, totalParticipants, totalWeight, ci.Url); err != nil {
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
	data, err := v.censusWarpcastChannel(channelID, userFID, nil)
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
	data, err := v.censusFollowers(userFID, nil)
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
		CommunityID string `json:"communityID"`
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
	// check if the community is ready (soft check, if it fails, continue)
	ready, _, err := v.CommunityStatus(community)
	if err != nil {
		log.Warnw("error getting community status", "err", err, "community", community.ID)
	}
	// if the community is not ready, return a PreconditionFailed error
	if !ready {
		return ctx.Send([]byte("community not ready"), http.StatusPreconditionFailed)
	}
	// getting the delegations of the community to build the census taking into
	// account them
	delegations, err := v.db.FinalDelegationsByCommunity(req.CommunityID)
	if err != nil {
		return err
	}
	// check the type to create it from the correct source (channel, airstak
	// (nft/erc20) or user followers) and in the correct way (async or sync)
	switch community.Census.Type {
	case mongo.TypeCommunityCensusFollowers:
		// if the census type is followers, create the census from the users who
		// follow the user, the process is async so return add the censusID to the
		// queue and return it to the client
		data, err := v.censusFollowers(userFID, delegations)
		if err != nil {
			log.Warnf("error creating census for the user: %d: %v", userFID, err)
			return ctx.Send([]byte("error creating user followers census"), http.StatusInternalServerError)
		}
		return ctx.Send(data, http.StatusOK)
	case mongo.TypeCommunityCensusChannel:
		// if the census type is a channel, create the census from the users who
		// follow the channel, the process is async so return add the censusID
		// to the queue and return it to the client
		data, err := v.censusWarpcastChannel(community.Census.Channel, userFID, delegations)
		if err != nil {
			log.Warnf("error creating census for the chanel: %s: %v", community.Census.Channel, err)
			return ctx.Send([]byte("error creating channel census"), http.StatusInternalServerError)
		}
		return ctx.Send(data, http.StatusOK)
	case mongo.TypeCommunityCensusNFT, mongo.TypeCommunityCensusERC20:
		// create the census from the token holders
		data, err := v.tokenBasedCensus(community.Census.Strategy, community.Census.Type, userFID, delegations)
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

// tokenBasedCensusBlockchains returns the supported blockchains for token
// based censuses, it queries the census3 API to get the supported blockchains.
func (v *vocdoniHandler) tokenBasedCensusBlockchains(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	info, err := v.census3.Info()
	if err != nil {
		return ctx.Send([]byte(fmt.Sprintf("error getting blockchains: %v", err)), http.StatusInternalServerError)
	}
	var blockchains []string
	for _, b := range info.SupportedChains {
		blockchains = append(blockchains, b.ShortName)
	}
	data, err := json.Marshal(map[string][]string{"blockchains": blockchains})
	if err != nil {
		return ctx.Send([]byte(fmt.Sprintf("error encoding blockchains: %v", err)), http.StatusInternalServerError)
	}
	return ctx.Send(data, http.StatusOK)
}

const (
	MAXNFTTokens   = 3
	MAXERC20Tokens = 1
)

// tokenBasedCensus method creates a new census from the token holders of a
// group of NFTs or a single ERC20 token. The census is created by the census
// strategy ID in Census3 service. The process is async and returns the json
// encoded censusID. It updates the progress in the queue and the result when
// it's ready.
func (v *vocdoniHandler) tokenBasedCensus(strategyID uint64, tokenType string, createdByFID uint64, delegations []*mongo.Delegation) ([]byte, error) {
	if v.census3 == nil {
		return nil, fmt.Errorf("census3 client not available")
	}
	censusID, err := v.cli.NewCensus(api.CensusTypeWeighted)
	if err != nil {
		return nil, err
	}
	if err := v.db.AddCensus(censusID, createdByFID); err != nil {
		return nil, fmt.Errorf("cannot add census to database: %w", err)
	}
	v.backgroundQueue.Store(censusID.String(), CensusInfo{})
	log.Debugw("building token based census", "censusID", censusID)
	go func() {
		startTime := time.Now()
		// get holders for each token
		var holders [][]string
		var err error
		v.trackStepProgress(censusID, 1, 3, func(progress chan int) {
			log.Debugw("getting holders from census3", "strategyID", strategyID)
			rawHolders, err := v.census3.AllHoldersByStrategy(strategyID)
			if err != nil {
				log.Warnw("failed to build token based census, cannot get holders", "err", err.Error())
				v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
				return
			}
			log.Debugw("holders received from census3", "count", len(rawHolders))
			for address, balance := range rawHolders {
				holders = append(holders, []string{address.Hex(), balance.String()})
			}
		})
		if err != nil {
			log.Warnw("failed to build census, cannot get holders", "err", err.Error())
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		// create census from token holders
		var participants []*FarcasterParticipant
		v.trackStepProgress(censusID, 2, 3, func(progress chan int) {
			log.Debugw("processing holders", "count", len(holders))
			participants, _, err = v.processCensusRecords(holders, delegations, progress)
		})
		if err != nil {
			log.Warnw("failed to build census", "err", err.Error())
			v.backgroundQueue.Store(censusID.String(), CensusInfo{Error: err.Error()})
			return
		}
		var ci *CensusInfo
		v.trackStepProgress(censusID, 3, 3, func(progress chan int) {
			if tokenType == mongo.TypeCommunityCensusERC20 {
				ci, err = CreateCensus(v.cli, participants, FrameCensusTypeERC20, progress)
			} else if tokenType == mongo.TypeCommunityCensusNFT {
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
		totalParticipants := uint32(0) // including delegations
		for _, p := range participants {
			if _, ok := uniqueParticipantsMap[p.Username]; ok {
				// if the username is already in the map, continue
				continue
			}
			uniqueParticipantsMap[p.Username] = p.Weight
			totalWeight.Add(totalWeight, p.Weight)
			totalParticipants += p.Delegations + 1
		}
		uniqueParticipants := []string{}
		for k := range uniqueParticipantsMap {
			uniqueParticipants = append(uniqueParticipants, k)
		}
		ci.Usernames = uniqueParticipants
		ci.FromTotalAddresses = uint32(len(holders))
		ci.FarcasterParticipantCount = uint32(len(uniqueParticipants))
		log.Infow("token census based created",
			"censusID", censusID.String(),
			"size", len(ci.Usernames),
			"totalWeight", totalWeight.String(),
			"duration", time.Since(startTime),
			"totalAddresses", ci.FromTotalAddresses,
			"participants", len(ci.Usernames),
			"totalParticipants", totalParticipants,
		)
		// store the census info in the memory map
		v.backgroundQueue.Store(censusID.String(), *ci)
		// add participants to the census in the database
		if err := v.db.AddParticipantsToCensus(
			censusID,
			uniqueParticipantsMap,
			ci.FromTotalAddresses,
			totalParticipants,
			totalWeight,
			ci.Url,
		); err != nil {
			log.Errorw(err, fmt.Sprintf("failed to add participants to census %s", censusID.String()))
		}
	}()

	// return the censusID to the client
	data, err := json.Marshal(map[string]string{"censusId": censusID.String()})
	if err != nil {
		return nil, err
	}
	return data, nil
}

// censusWarpcastChannel helper method creates a new census from a Warpcast
// Channel. The process is async and returns the json encoded censusID. It
// updates the progress in the queue and the result when it's ready.
func (v *vocdoniHandler) censusWarpcastChannel(channelID string, authorFID uint64, delegations []*mongo.Delegation) ([]byte, error) {
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
			participants = v.farcasterCensusFromFids(users, delegations, progress)
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
		totalParticipants := uint32(0) // including delegations
		for _, p := range participants {
			if _, ok := uniqueParticipantsMap[p.Username]; !ok {
				uniqueParticipantsMap[p.Username] = new(big.Int).SetUint64(1)
				totalParticipants += p.Delegations + 1
			}
		}
		// only return the username list if it's less than the maxUsersNamesToReturn
		if len(uniqueParticipantsMap) < maxUsersNamesToReturn {
			for username := range uniqueParticipantsMap {
				censusInfo.Usernames = append(censusInfo.Usernames, username)
			}
		}
		censusInfo.FromTotalAddresses = uint32(len(users))
		censusInfo.FarcasterParticipantCount = uint32(len(uniqueParticipantsMap))
		v.backgroundQueue.Store(censusID.String(), *censusInfo)
		// add participants to the census in the database
		if err := v.db.AddParticipantsToCensus(
			censusID,
			uniqueParticipantsMap,
			censusInfo.FromTotalAddresses,
			totalParticipants,
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
func (v *vocdoniHandler) censusFollowers(userFID uint64, delegations []*mongo.Delegation) ([]byte, error) {
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
			participants = v.farcasterCensusFromFids(users, delegations, progress)
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
		totalParticipants := uint32(0) // including delegations
		for _, p := range participants {
			if _, ok := uniqueParticipantsMap[p.Username]; ok {
				// if the username is already in the map, continue
				continue
			}
			uniqueParticipantsMap[p.Username] = p.Weight
			totalWeight.Add(totalWeight, p.Weight)
			totalParticipants += p.Delegations + 1
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
			totalParticipants,
			totalWeight,
			censusInfo.Url,
		); err != nil {
			log.Errorw(err, fmt.Sprintf("failed to add participants to census %s", censusID.String()))
		}

		censusInfo.FromTotalAddresses = uint32(len(users))
		censusInfo.FarcasterParticipantCount = uint32(len(uniqueParticipantsMap))
		v.backgroundQueue.Store(censusID.String(), *censusInfo)
		log.Infow("census created from user followers",
			"fid", userFID,
			"participants", len(censusInfo.Usernames),
			"totalParticipants", totalParticipants,
		)
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
			participants = v.farcasterCensusFromFids(users, nil, progress)
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
		censusInfo.FarcasterParticipantCount = uint32(len(uniqueParticipantsMap))
		v.backgroundQueue.Store(censusID.String(), *censusInfo)
		// add participants to the census in the database
		if err := v.db.AddParticipantsToCensus(
			censusID,
			uniqueParticipantsMap,
			censusInfo.FromTotalAddresses,
			uint32(len(uniqueParticipantsMap)),
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

func (v *vocdoniHandler) farcasterCensusFromEthereumCSV(csv []byte, progress chan int) ([]*FarcasterParticipant, uint32, error) {
	records, err := ParseCSV(csv)
	if err != nil {
		return nil, 0, err
	}
	return v.processCensusRecords(records, nil, progress)
}

// farcasterCensusFromFids creates a list of Farcaster participants from a list
// of FIDs. It queries the database to get the users signer keys and creates the
// participants from them. It returns the list of participants and a map of the
// FIDs that failed to get the users from the database or decoding the keys.
func (v *vocdoniHandler) farcasterCensusFromFids(fids []uint64, delegations []*mongo.Delegation, progress chan int) []*FarcasterParticipant {
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
			// by default, a user has not delegated weight and has a weight of
			// 1. If the user has the vote delegated, the weight is 0. If the
			// user has votes delegations, the weight is the number of
			// delegations. The final user weight is the sum of the user weight
			// and the delegated weight, if that sum is 0, the user is not
			// included in the census.
			userWeight := int64(1)
			delegatedWeight := int64(0)
			for _, delegation := range delegations {
				if delegation.From == fid {
					userWeight = 0
				}
				if delegation.To == fid {
					delegatedWeight++
				}
			}
			// if the final weight is 0, the user is not included in the census
			finalWeight := userWeight + delegatedWeight
			if finalWeight == 0 {
				return
			}
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
				// send the participant to the channel
				safeSendParticipant(participantsCh, &FarcasterParticipant{
					PubKey:   signerBytes,
					Weight:   big.NewInt(finalWeight),
					Username: user.Username,
					FID:      fid,
				})
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

// findWeightAndSignersForCensusRecord is a helper function for the processCensusRecords method.
// It finds the final weight, including delegations, and the signers of a user.
// It creates a FarcasterParticipant for each signer of the user and sends it to the participants channel.
// If a participant has weight 0, it is not sent to the channel.
func findWeightAndSignersForCensusRecord(user *mongo.User, addressMap map[string]*big.Int, db *mongo.MongoStorage, delegations []*mongo.Delegation, participantsCh chan *FarcasterParticipant) {
	if user == nil || db == nil || participantsCh == nil {
		return
	}
	// find the addres on the map to get the weight
	// the weight is the sum of the weights of all the addresses of the user
	userWeight := new(big.Int).SetUint64(0)
	for _, addr := range user.Addresses {
		weightAddress, ok := addressMap[helpers.NormalizeAddressString(addr)]
		if ok {
			userWeight = userWeight.Add(userWeight, weightAddress)
		}
	}
	// by default, a user has not delegated weight and has a weight is
	// the sum of weights of all addresses of the user. If the user has
	// the vote delegated, the weight is 0. If the
	// user has votes delegations, the weight is the number of
	// delegations. The final user weight is the sum of the user weight
	// and the delegated weight, if that sum is 0, the user is not
	// included in the census.
	delegatedWeight := big.NewInt(0)
	delegationsCount := uint32(0)
	for _, delegation := range delegations {
		// if the user has delegated its vote, assign weight 0
		if delegation.From == user.UserID {
			log.Debugw("found delegation for user (outgoing)", "fid", user.Username, "delegated to", delegation.To)
			userWeight = big.NewInt(0)
			continue
		}
		// if the user has votes delegated to him, sum them on deleagatedWeight
		if delegation.To == user.UserID {
			log.Debugw("found delegation for user (incoming)", "fid", user.Username, "delegated from", delegation.From)
			delegator, err := db.User(delegation.From)
			if err != nil {
				log.Warnw("error fetching user from database", "fid", delegation.From, "error", err)
				continue
			}
			partialDelegatedWeight := big.NewInt(0)
			// sum the weight of all the addresses of the delegator
			for _, addr := range delegator.Addresses {
				weightAddress, ok := addressMap[helpers.NormalizeAddressString(addr)]
				if ok {
					partialDelegatedWeight = partialDelegatedWeight.Add(partialDelegatedWeight, weightAddress)
				}
			}
			// if the weight is 0, the delegator is not included in the census and the delegation is ignored
			if partialDelegatedWeight.Cmp(big.NewInt(0)) != 0 {
				delegationsCount++
				delegatedWeight = delegatedWeight.Add(delegatedWeight, partialDelegatedWeight)
			} else {
				log.Warnw("delegator has no weight, skiping...", "fid", delegation.From, "address", delegator.Addresses)
			}
		}
	}
	// if the final weight is 0, the user is not included in the census
	finalWeight := userWeight.Add(userWeight, delegatedWeight)
	if finalWeight.Cmp(big.NewInt(0)) == 0 {
		return
	}

	for _, signer := range user.Signers {
		signerBytes, err := hex.DecodeString(strings.TrimPrefix(signer, "0x"))
		if err != nil {
			log.Warnw("error decoding signer", "signer", signer, "err", err)
			continue
		}
		safeSendParticipant(participantsCh, &FarcasterParticipant{
			PubKey:      signerBytes,
			Weight:      finalWeight,
			Username:    user.Username,
			FID:         user.UserID,
			Delegations: delegationsCount,
		})
	}
}

func safeSendParticipant(ch chan *FarcasterParticipant, value *FarcasterParticipant) {
	defer func() {
		if r := recover(); r != nil {
			log.Warn("attempted to send participant on a closed channel")
		}
	}()
	ch <- value
}

// processRecord processes a single record of a plain-text census and returns the corresponding Farcaster participants.
// The record is expected to be a string containing the address and the weight.
// Returns the list of participants and the total number of unique addresses available in the records.
func (v *vocdoniHandler) processCensusRecords(records [][]string, delegations []*mongo.Delegation, progress chan int) ([]*FarcasterParticipant, uint32, error) {
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
		if weightRecord == "0" {
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
	var totalProcessedAddresses atomic.Uint32

	// Collect all addresses to process
	addresses := make([]string, 0, len(addressMap))
	for address := range addressMap {
		addresses = append(addresses, address)
	}

	// Defer closing the channels
	defer func() {
		close(participantsCh)
		close(pendingAddressesCh)
	}()

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
				progress <- int(100 * totalProcessedAddresses.Load() / uniqueAddressesCount)
			}
		}
	}()

	// Fetch users by addresses in bulk
	log.Infow("fetching users from database", "count", len(addresses))
	startTime := time.Now()
	batchSize := 10000
	for i := 0; i < len(addresses); i += batchSize {
		end := i + batchSize
		if end > len(addresses) {
			end = len(addresses)
		}
		batchAddresses := addresses[i:end]
		usersByAddress, err := v.db.UserByAddressBulk(batchAddresses)
		if err != nil {
			return nil, 0, fmt.Errorf("error fetching users from database: %w", err)
		}

		for _, addr := range batchAddresses {
			// Process the results of this batch, for each address check if it was found in the database
			concurrencyLimit <- struct{}{}
			wg.Add(1)
			go func(addr string) {
				defer wg.Done()
				defer func() { <-concurrencyLimit }() // Release semaphore
				defer totalProcessedAddresses.Add(1)

				user, ok := usersByAddress[helpers.NormalizeAddressString(addr)]
				if !ok || user == nil {
					pendingAddressesCh <- addr
					return
				}

				// If the user has no signers, add it to the pending addresses
				if len(user.Signers) == 0 {
					pendingAddressesCh <- addr
					return
				}

				findWeightAndSignersForCensusRecord(user, addressMap, v.db, delegations, participantsCh)
				processedAddresses.Add(1)
			}(addr)
		}
		wg.Wait()
	}
	log.Infow("users fetched from database", "count", processedAddresses.Load(), "elapsed (s)", time.Since(startTime).Seconds())

	// Fetch the remaining users from the Neynar API. Only if the number of cenus addresses is less than 5000
	if len(pendingAddresses) < 5000 {
		count := 0
		for i := 0; i < len(pendingAddresses); i += neynar.MaxAddressesPerRequest {
			// Fetch the user data from the farcaster API
			ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			to := i + neynar.MaxAddressesPerRequest
			if to > len(pendingAddresses) {
				to = len(pendingAddresses)
			}
			log.Debugw("fetching users from neynar", "from", i, "to", to, "total", len(pendingAddresses))
			usersData, err := v.fcapi.UserDataByVerificationAddress(ctx2, pendingAddresses[i:to])
			if err != nil {
				if errors.Is(err, farcasterapi.ErrNoDataFound) {
					break
				}
				log.Errorw(err, "error fetching users from Neynar API")
			}
			log.Debugw("users found on neynar", "count", len(usersData))
			for _, userData := range usersData {
				// Add or update the user on the database
				dbUser, err := v.db.User(userData.FID)
				if err != nil {
					log.Debugw("adding new user to database", "fid", userData.FID)
					if err := v.db.AddUser(
						userData.FID,
						userData.Username,
						userData.Displayname,
						helpers.NormalizeAddressStringSlice(userData.VerificationsAddresses),
						userData.Signers,
						helpers.NormalizeAddressString(userData.CustodyAddress),
						0,
					); err != nil {
						return nil, 0, err
					}
				} else {
					log.Debugw("updating user on database", "fid", userData.FID)
					dbUser.Addresses = helpers.NormalizeAddressStringSlice(userData.VerificationsAddresses)
					dbUser.Username = userData.Username
					dbUser.Signers = userData.Signers
					dbUser.CustodyAddress = helpers.NormalizeAddressString(userData.CustodyAddress)
					if err := v.db.UpdateUser(dbUser); err != nil {
						return nil, 0, err
					}
				}

				findWeightAndSignersForCensusRecord(dbUser, addressMap, v.db, delegations, participantsCh)
				count++
			}
			processedAddresses.Add(uint32(to - i))
		}
		if len(pendingAddresses) > 0 {
			log.Infow("users found on neynar", "count", count, "ratio", fmt.Sprintf("%.2f%%", 100*float64(count)/float64(len(pendingAddresses))))
		}
	} else {
		log.Warnf("skipping fetching users from Neynar API due to the number of pending addresses %d", len(pendingAddresses))
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

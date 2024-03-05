package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vocdoni/vote-frame/imageframe"
	"github.com/vocdoni/vote-frame/mongo"
	"github.com/vocdoni/vote-frame/shortener"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

const (
	ElectionSourceWebApp = "farcaster.vote"
	ElectionSourceBot    = "bot"
)

func (v *vocdoniHandler) election(electionID types.HexBytes) (*api.Election, error) {
	electionCached, ok := v.electionLRU.Get(electionID.String())
	if ok {
		return electionCached, nil
	}
	election, err := v.cli.Election(electionID)
	if err != nil {
		return nil, ErrElectionUnknown
	}
	evicted := v.electionLRU.Add(electionID.String(), election)
	log.Debugw("added election to cache", "electionID", electionID, "evicted", evicted)
	return election, nil
}

func (v *vocdoniHandler) createElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	var req *ElectionCreateRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return fmt.Errorf("failed to unmarshal election request: %w", err)
	}

	// use the request census or use the one hardcoded for all farcaster users
	census := req.Census
	if census == nil {
		census = v.defaultCensus
	}

	if req.Duration == 0 {
		req.Duration = time.Hour * 24
	} else {
		req.Duration *= time.Hour
		if req.Duration > maxElectionDuration {
			return fmt.Errorf("election duration too long")
		}
	}

	electionID, err := v.createAndSaveElectionAndProfile(&req.ElectionDescription, census, req.Profile, false, ElectionSourceWebApp)
	if err != nil {
		return fmt.Errorf("failed to create election: %v", err)
	}
	ctx.Writer.Header().Set("Content-Type", "application/json")
	return ctx.Send([]byte(electionID.String()), http.StatusOK)
}

func (v *vocdoniHandler) showElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	electionIDbytes, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	// check if the election is finished and if so, send the final results
	if v.checkIfElectionFinishedAndHandle(electionIDbytes, ctx) {
		return nil
	}

	log.Infow("received show election request", "electionID", ctx.URLParam("electionID"))

	// create a PNG image with the election description
	election, err := v.election(electionID)
	if err != nil {
		return fmt.Errorf("failed to fetch election: %w", err)
	}
	png, err := imageframe.QuestionImage(election)
	if err != nil {
		return fmt.Errorf("failed to generate image: %v", err)
	}
	// send the response
	response := strings.ReplaceAll(frame(frameVote), "{image}", v.addImageToCache(png, electionIDbytes))
	response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])
	response = strings.ReplaceAll(response, "{processID}", ctx.URLParam("electionID"))

	r := election.Metadata.Questions[0].Choices
	for i := 0; i < 4; i++ {
		if len(r) > i {
			opt := ""
			switch i {
			case 0:
				opt = "1️⃣"
			case 1:
				opt = "2️⃣"
			case 2:
				opt = "3️⃣"
			case 3:
				opt = "4️⃣"
			}
			response = strings.ReplaceAll(response, fmt.Sprintf("{option%d}", i), opt)
			continue
		}
		response = strings.ReplaceAll(response, fmt.Sprintf("{option%d}", i), "")
	}

	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func (v *vocdoniHandler) checkElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	if electionID == nil {
		return ctx.Send(nil, http.StatusNoContent)
	}
	_, ok := v.electionLRU.Get(fmt.Sprintf("%x", electionID))
	if !ok {
		return ctx.Send(nil, http.StatusNoContent)
	}
	frameUrl := fmt.Sprintf("%s/%x", serverURL, electionID)
	resultURL, err := shortener.ShortURL(ctx.Request.Context(), frameUrl)
	if err != nil {
		resultURL = fmt.Sprintf("%s/%x", serverURL, electionID)
	}
	body, err := json.Marshal(map[string]string{"url": resultURL})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	return ctx.Send(body, http.StatusOK)
}

// votersForElection returns the list of voters for the given election.
func (v *vocdoniHandler) votersForElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	usernames, err := v.db.VotersOfElection(electionID)
	if err != nil {
		return fmt.Errorf("failed to get voters of election: %w", err)
	}
	data, err := json.Marshal(map[string][]string{"voters": usernames})
	if err != nil {
		return fmt.Errorf("failed to marshal voters: %w", err)
	}
	return ctx.Send(data, http.StatusOK)
}

func newElectionDescription(description *ElectionDescription, census *CensusInfo) *api.ElectionDescription {
	choices := []api.ChoiceMetadata{}

	for i, choice := range description.Options {
		choices = append(choices, api.ChoiceMetadata{
			Title: map[string]string{"default": choice},
			Value: uint32(i),
		})
	}

	size := census.Size
	if size > uint64(maxElectionSize) {
		size = uint64(maxElectionSize)
	}

	return &api.ElectionDescription{
		Title:       map[string]string{"default": description.Question},
		Description: map[string]string{"default": "this is a farcaster frame poll"},
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
	electionID, err := cli.NewElection(newElectionDescription(description, census), false)
	if err != nil {
		return nil, err
	}
	return electionID, nil
}

// waitForElection waits until the election is created or returns an error.
func waitForElection(cli *apiclient.HTTPclient, electionID types.HexBytes) (*api.Election, error) {
	// Wait until the election is created
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*40)
	defer cancel()
	election := &api.Election{}
	startTime := time.Now()
	var err error
	for {
		election, err = cli.Election(electionID)
		if err != nil {
			// Return an error if the received error is not a '404 - Not found'
			// error which means that the election has not yet been created.
			if !strings.Contains(err.Error(), "API error: 404") {
				log.Warnw("failed to get election", "id", electionID, "err", err)
			}
		}
		if election != nil {
			break
		}
		select {
		case <-time.After(time.Second * 2):
			continue
		case <-ctx.Done():
			return nil, fmt.Errorf("election %x not created after %s: %w",
				electionID, time.Since(startTime).String(), ctx.Err())
		}
	}
	log.Infow("created new election", "id", election.ElectionID.String())
	return election, nil
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

// createAndSaveElectionAndProfile creates an election and saves it in the
// database. It receives a description of the election, a census, a profile and
// a wait flag. If the wait flag is true, it waits until the election is created
// and saved in the database.
func (v *vocdoniHandler) createAndSaveElectionAndProfile(desc *ElectionDescription,
	census *CensusInfo, profile *FarcasterProfile, wait bool, source string,
) (types.HexBytes, error) {
	// use the request census or use the one hardcoded for all farcaster users
	if census == nil {
		census = v.defaultCensus
	}
	// create the election
	electionID, err := createElection(v.cli, desc, census)
	if err != nil {
		return nil, fmt.Errorf("failed to create election: %v", err)
	}
	// create an inline function with the rest of the method logic
	backgroundProcess := func() error {
		election, err := waitForElection(v.cli, electionID)
		if err != nil {
			return fmt.Errorf("failed to create election: %w", err)
		}
		if err := v.saveElectionAndProfile(election, profile, source); err != nil {
			return fmt.Errorf("failed to save election and profile: %w", err)
		}
		return nil
	}
	// if wait flag is true, run the background process in the current goroutine
	// and return the resulting error, otherwise, run it in a new goroutine and
	// print the error if any.
	if wait {
		return electionID, backgroundProcess()
	} else {
		go func() {
			if err := backgroundProcess(); err != nil {
				log.Errorw(err, "failed to create election")
			}
		}()
	}
	return electionID, nil
}

// saveElectionAndProfile saves the election and the profile in the database.
func (v *vocdoniHandler) saveElectionAndProfile(election *api.Election, profile *FarcasterProfile, source string) error {
	// add the election to the LRU cache and the database
	v.electionLRU.Add(election.ElectionID.String(), election)
	if err := v.db.AddElection(election.ElectionID, profile.FID, source); err != nil {
		return fmt.Errorf("failed to add election to database: %w", err)
	}
	u, err := v.db.User(profile.FID)
	if err != nil {
		if !errors.Is(err, mongo.ErrUserUnknown) {
			return fmt.Errorf("failed to get user from database: %w", err)
		}
		if err := v.db.AddUser(profile.FID, profile.Username, profile.Verifications, []string{}, profile.Custody, 1); err != nil {
			return fmt.Errorf("failed to add user to database: %w", err)
		}
		return nil
	}
	u.Addresses = profile.Verifications
	u.Username = profile.Username
	u.ElectionCount++
	if err := v.db.UpdateUser(u); err != nil {
		return fmt.Errorf("failed to update user in database: %w", err)
	}
	return nil
}

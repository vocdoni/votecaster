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

	"github.com/vocdoni/vote-frame/features"
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
	// MaxUsersToNotify is the maximum number of users to notify in a single
	// election. If the census is larger than this number, the notification
	// will not be sent, but the election will still be created.
	MaxUsersToNotify = 1000
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
	// Get the user from the database to log the user creating the election and
	// check if the user has the required reputation to differents features.
	var accessProfile *mongo.UserAccessProfile
	fid, err := v.db.UserFromAuthToken(msg.AuthToken)
	if err != nil {
		log.Errorf("failed to get user from auth token %s: %v", msg.AuthToken, err)
	} else {
		user, err := v.db.User(fid)
		if err != nil {
			return fmt.Errorf("failed to get user from database: %w", err)
		}
		accessProfile, err = v.db.UserAccessProfile(fid)
		if err != nil {
			return fmt.Errorf("failed to get user access profile: %w", err)
		}
		// log the user creating the election for debugging purposes
		log.Infow("user creating election", "username", user.Username, "fid", fid)
	}
	// check if the user has enough reputation to notify voters
	if req.NotifyUsers && !features.IsAllowed(features.NOTIFY_USERS, accessProfile.Reputation) {
		return ctx.Send([]byte("user does not have enough reputation to notify voters"), http.StatusBadRequest)
	}
	// use the request census or use the one hardcoded for all farcaster users
	census := req.Census
	if census == nil {
		census = v.defaultCensus
	}
	// if no duration is provided, set it to 24 hours, otherwise, set it to the
	// provided duration in hours unless it is greater than the maximum allowed
	if req.Duration == 0 {
		req.Duration = time.Hour * 24
	} else {
		req.Duration *= time.Hour
		if req.Duration > maxElectionDuration {
			return fmt.Errorf("election duration too long")
		}
	}
	// create the election description
	req.ElectionDescription.UsersCount = uint32(len(census.Usernames))
	req.ElectionDescription.UsersCountInitial = uint32(census.FromTotalAddresses)
	// create the election and save it in the database
	electionID, err := v.createAndSaveElectionAndProfile(&req.ElectionDescription, census,
		req.Profile, false, req.NotifyUsers, req.NotificationText, ElectionSourceWebApp,
		req.CommunityID)
	if err != nil {
		return fmt.Errorf("failed to create election: %v", err)
	}
	// set the electionID for the census root previously stored on the database (if any).
	if req.Census != nil && req.Census.Root != nil {
		if err := v.db.SetElectionIdForCensusRoot(req.Census.Root, electionID); err != nil {
			log.Errorw(err, fmt.Sprintf("failed to set electionID for census root %s", req.Census.Root))
		}
	}
	// return the electionID
	ctx.SetResponseContentType("application/json")
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
	response := strings.ReplaceAll(frame(frameVote), "{image}", imageLink(png))
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
	census *CensusInfo, profile *FarcasterProfile, wait bool, notify bool,
	customText string, source string, communityID *uint64,
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
		if err := v.saveElectionAndProfile(election, profile, source, desc.UsersCount,
			desc.UsersCountInitial, census.TokenDecimals, communityID); err != nil {
			return fmt.Errorf("failed to save election and profile: %w", err)
		}
		if notify {
			if len(census.Usernames) > MaxUsersToNotify {
				return fmt.Errorf("census too large to notify users but election has been created successfully")
			}
			// set the notification deadline to 10 minutes before the election
			// ends if the election ends in less than 3 hours, otherwise, set it
			// to 3 hours before the election ends.
			expiration := election.EndDate
			if time.Until(expiration) < time.Hour*3 {
				expiration = expiration.Add(-time.Minute * 10)
			} else {
				expiration = expiration.Add(-time.Hour * 3)
			}
			frameURL := fmt.Sprintf("%s/%x", serverURL, electionID)
			if err := v.createNotifications(
				electionID,
				profile.FID,
				profile.DisplayName,
				census.Usernames,
				frameURL,
				customText,
				expiration,
			); err != nil {
				return fmt.Errorf("failed to create notifications: %w", err)
			}
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
func (v *vocdoniHandler) saveElectionAndProfile(
	election *api.Election,
	profile *FarcasterProfile,
	source string,
	usersCount, usersCountInitial, tokenDecimals uint32,
	communityID *uint64,
) error {
	if election == nil || election.Metadata == nil || len(election.Metadata.Questions) == 0 {
		return fmt.Errorf("invalid election")
	}
	var community *mongo.ElectionCommunity
	if communityID != nil {
		c, err := v.db.Community(*communityID)
		if err != nil {
			return fmt.Errorf("failed to get community from database: %w", err)
		}
		community = &mongo.ElectionCommunity{
			ID:   c.ID,
			Name: c.Name,
		}
	}
	// add the election to the LRU cache and the database
	v.electionLRU.Add(election.ElectionID.String(), election)
	if err := v.db.AddElection(
		election.ElectionID,
		profile.FID,
		source,
		election.Metadata.Title["default"],
		usersCount,
		usersCountInitial,
		tokenDecimals,
		community); err != nil {
		return fmt.Errorf("failed to add election to database: %w", err)
	}
	u, err := v.db.User(profile.FID)
	if err != nil {
		if !errors.Is(err, mongo.ErrUserUnknown) {
			return fmt.Errorf("failed to get user from database: %w", err)
		}
		if err := v.db.AddUser(profile.FID, profile.Username, profile.DisplayName, profile.Verifications, []string{}, profile.Custody, 1); err != nil {
			return fmt.Errorf("failed to add user to database: %w", err)
		}
		return nil
	}
	u.Addresses = profile.Verifications
	u.Username = profile.Username
	u.Displayname = profile.DisplayName
	u.ElectionCount++
	if err := v.db.UpdateUser(u); err != nil {
		return fmt.Errorf("failed to update user in database: %w", err)
	}
	return nil
}

// finalizeElectionsAtBackround checks for elections without results and finalizes them.
// Stores the final results as a static PNG image in the database. It must run in the background.
func (v *vocdoniHandler) finalizeElectionsAtBackround(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(120 * time.Second):
			electionIDs, err := v.db.ElectionsWithoutResults(ctx)
			if err != nil {
				log.Errorw(err, "failed to get elections without results")
				continue
			}
			if len(electionIDs) > 0 {
				log.Debugw("found elections without results", "count", len(electionIDs))
			}
			for _, electionID := range electionIDs {
				time.Sleep(5 * time.Second)
				electionIDbytes, err := hex.DecodeString(electionID)
				if err != nil {
					log.Errorw(err, "failed to decode electionID")
					continue
				}
				election, err := v.cli.Election(electionIDbytes)
				if err != nil {
					log.Errorw(err, "failed to get election")
					continue
				}
				electionMeta, err := v.db.Election(electionIDbytes)
				if err != nil {
					log.Errorw(err, "failed to get election")
					continue
				}
				if election.FinalResults {
					png, err := imageframe.ResultsImage(election, electionMeta.CensusERC20TokenDecimals)
					if err != nil {
						log.Errorw(err, "failed to generate results image")
						continue
					}
					if err := v.db.AddFinalResults(electionIDbytes, imageframe.FromCache(png)); err != nil {
						log.Errorw(err, "failed to add final results to database")
						continue
					}
					log.Infow("finalized election", "electionID", electionID)
				}
			}
		}
	}
}

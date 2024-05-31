package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/vocdoni/vote-frame/farcasterapi/warpcast"
	"github.com/vocdoni/vote-frame/features"
	"github.com/vocdoni/vote-frame/helpers"
	"github.com/vocdoni/vote-frame/imageframe"
	"github.com/vocdoni/vote-frame/mongo"
	"github.com/vocdoni/vote-frame/shortener"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
	"go.vocdoni.io/dvote/util"
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

	// if the poll is for a community, check if the user is an admin of the
	// community and if the community is disabled
	if req.CommunityID != nil {
		// check if the user is an admin of the community
		if !v.db.IsCommunityAdmin(fid, *req.CommunityID) {
			return fmt.Errorf("user is not an admin of the community")
		}
		// check if the community is disabled
		if v.db.IsCommunityDisabled(*req.CommunityID) {
			return fmt.Errorf("community is disabled")
		}
	}
	// if notifications are enabled, check if the poll is for a community
	if req.NotifyUsers {
		if req.CommunityID == nil {
			return ctx.Send([]byte("notifications are only available for community polls"), http.StatusBadRequest)
		}
		// check if the user has enough reputation to notify voters
		if !features.IsAllowed(features.NOTIFY_USERS, accessProfile.Reputation) {
			return ctx.Send([]byte("user does not have enough reputation to notify voters"), http.StatusBadRequest)
		}
		// check if the community allows notifications
		if !v.db.CommunityAllowNotifications(*req.CommunityID) {
			return fmt.Errorf("community does not allow notifications")
		}
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
	uniqueUsernames := map[string]bool{}
	for _, u := range census.Usernames {
		uniqueUsernames[u] = true
	}
	req.ElectionDescription.UsersCount = uint32(len(uniqueUsernames))
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
	response = strings.ReplaceAll(response, "{state}", ctx.URLParam("electionID"))

	r := election.Metadata.Questions[0].Choices
	for i := 0; i < 4; i++ {
		if len(r) > i {
			opt := ""
			switch i {
			case 0:
				opt = "A"
			case 1:
				opt = "B"
			case 2:
				opt = "C"
			case 3:
				opt = "D"
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
	// get current voters of the election
	voters, err := v.db.VotersOfElection(electionID)
	if err != nil && !errors.Is(err, mongo.ErrElectionUnknown) {
		return fmt.Errorf("failed to get voters of election: %w", err)
	}
	// get the usernames of the voters and create an index for faster access
	// to the voters to calculate the remaining usernames
	votersUsernames := []string{}
	for _, u := range voters {
		votersUsernames = append(votersUsernames, u.Username)
	}
	// send the response
	data, err := json.Marshal(ElectionVotersUsernames{
		Usernames: votersUsernames,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal voters: %w", err)
	}
	return ctx.Send(data, http.StatusOK)
}

// votersForElection returns the list of voters for the given election.
func (v *vocdoniHandler) remainingVotersForElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	// get current voters of the election
	voters, err := v.db.VotersOfElection(electionID)
	if err != nil && !errors.Is(err, mongo.ErrElectionUnknown) {
		return fmt.Errorf("failed to get voters of election: %w", err)
	}
	// get the census of the election
	census, err := v.db.CensusFromElection(electionID)
	if err != nil {
		if errors.Is(err, mongo.ErrElectionUnknown) {
			return ctx.Send([]byte("census not found"), http.StatusNotFound)
		}
		return fmt.Errorf("failed to get census from election: %w", err)
	}
	// create an index for faster access to the voters to calculate the remaining usernames
	votersIndex := make(map[string]bool)
	for _, u := range voters {
		votersIndex[u.Username] = true
	}
	// calculate the remaining usernames
	remainingUsernames := []string{}
	for username := range census.Participants {
		if _, ok := votersIndex[username]; !ok {
			remainingUsernames = append(remainingUsernames, username)
		}
	}
	slices.Sort(remainingUsernames)
	// send the response
	data, err := json.Marshal(ElectionVotersUsernames{
		Usernames: remainingUsernames,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal voters: %w", err)
	}
	return ctx.Send(data, http.StatusOK)
}

func (v *vocdoniHandler) electionFullInfo(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	dbElection, err := v.db.Election(electionID)
	if err != nil {
		return fmt.Errorf("could not fetch election %x: %w", electionID, err)
	}

	var username, displayname string
	user, err := v.db.User(dbElection.UserID)
	if err != nil {
		log.Warnw("failed to fetch user", "error", err)
		username = "unknown"
	} else {
		username = user.Username
		displayname = user.Displayname
	}
	census, err := v.db.CensusFromElection(electionID)
	if err != nil {
		log.Warnw("census not found for community election", "electionID", hex.EncodeToString(electionID))
		census = &mongo.Census{
			TotalWeight:  "0",
			Participants: make(map[string]string),
		}
	}

	// Fetch results from the database to return them in the response
	finalized := false
	results, err := v.db.Results(electionID)
	if err == nil {
		finalized = results.Finalized
	}

	if !finalized { // election is not finalized, so we need to fetch the results from the Vochain API and update the database
		results, err = v.updateAndFetchResultsFromDatabase(electionID, nil)
		if err != nil {
			return fmt.Errorf("failed to update/fetch results: %w", err)
		}
	}

	// Fetch participants
	participantFIDS := []uint64{}
	participants, err := v.db.VotersOfElection(electionID)
	if err != nil {
		log.Warnw("failed to fetch participants", "error", err)
	} else {
		for _, p := range participants {
			participantFIDS = append(participantFIDS, p.UserID)
		}
	}

	electionInfo := &ElectionInfo{
		CreatedTime:             dbElection.CreatedTime,
		ElectionID:              dbElection.ElectionID,
		LastVoteTime:            dbElection.LastVoteTime,
		EndTime:                 dbElection.EndTime,
		Question:                dbElection.Question,
		CastedVotes:             dbElection.CastedVotes,
		CensusParticipantsCount: uint64(dbElection.FarcasterUserCount),
		Turnout:                 helpers.CalculateTurnout(census.TotalWeight, dbElection.CastedWeight),
		FID:                     dbElection.UserID,
		Username:                username,
		Displayname:             displayname,
		TotalWeight:             census.TotalWeight,
		CastedWeight:            dbElection.CastedWeight,
		Participants:            participantFIDS,
		Choices:                 results.Choices,
		Votes:                   results.Votes,
		Finalized:               results.Finalized,
		Community:               dbElection.Community,
	}

	jresponse, err := json.Marshal(map[string]any{
		"poll": electionInfo,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.SetResponseContentType("application/json")
	return ctx.Send(jresponse, http.StatusOK)
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
			desc.UsersCountInitial, communityID); err != nil {
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
	usersCount, usersCountInitial uint32,
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
		election.EndDate,
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
		case <-time.After(60 * time.Second):
			electionIDs, err := v.db.ElectionsWithoutResults()
			if err != nil {
				log.Errorw(err, "failed to get elections without results")
				continue
			}
			for _, electionID := range electionIDs {
				electionIDbytes, err := hex.DecodeString(electionID)
				if err != nil {
					log.Errorw(err, fmt.Sprintf("failed to decode electionID: %s", electionID))
					continue
				}
				election, err := v.cli.Election(electionIDbytes)
				if err != nil {
					log.Errorw(err, fmt.Sprintf("failed to get election from API: %s", electionID))
					continue
				}
				if election.FinalResults {
					electiondb, err := v.db.Election(electionIDbytes)
					if err != nil {
						log.Errorw(err, fmt.Sprintf("failed to get election from database: %x", electionIDbytes))
						continue
					}
					if _, err = v.finalizeElectionResults(election, electiondb); err != nil {
						log.Errorw(err, fmt.Sprintf("failed to finalize election results: %x", electionIDbytes))
					}
				}
			}
		}
	}
}

// remindersHandler returns the remindable voters and the number of already
// reminded voters of an election. It requires the user to be the owner of the
// election.
func (v *vocdoniHandler) remindersHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the authenticated user from the token
	token := msg.AuthToken
	if token == "" {
		return fmt.Errorf("missing auth token header")
	}
	auth, err := v.db.UpdateActivityAndGetData(token)
	if err != nil {
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusNotFound)
	}
	// get the election id from the url params
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	// check that the user is the owner of the election
	election, err := v.db.Election(electionID)
	if err != nil {
		if err == mongo.ErrElectionUnknown {
			return ctx.Send([]byte("election not found"), http.StatusNotFound)
		}
		return fmt.Errorf("failed to get election: %w", err)
	}
	// check if the election is a community election and if the user is an admin
	if election.Community == nil || election.Community.ID == 0 {
		return fmt.Errorf("election is not a community election")
	}
	if !v.db.IsCommunityAdmin(auth.UserID, election.Community.ID) {
		return ctx.Send([]byte("user is not an admin of the community"), http.StatusForbidden)
	}
	// get the remindable users and the number of alredy reminded users from the
	// database
	remindableUsers, remindersSent, err := v.db.RemindersOfElection(electionID)
	if err != nil {
		return fmt.Errorf("failed to get voters of election: %w", err)
	}
	// if the reminders have not been populated yet, populate them and retry
	if len(remindableUsers) == 0 && remindersSent == 0 {
		return fmt.Errorf("failed to get remindable voters")
	}
	// get the census to include the voters weight
	census, err := v.db.CensusFromElection(electionID)
	if err != nil {
		return fmt.Errorf("failed to get census from election: %w", err)
	}
	reminableWeights := map[uint64]string{}
	for fid := range remindableUsers {
		if weight, ok := census.Participants[remindableUsers[fid]]; ok {
			reminableWeights[fid] = weight
		}
	}
	// remove the authenticated user from the remindable users
	delete(remindableUsers, auth.UserID)
	delete(reminableWeights, auth.UserID)
	// get the maximum number of direct messages the user can send by reputation
	maxDMs := v.db.MaxDirectMessages(auth.UserID, maxDirectMessages)
	if uint32(len(remindableUsers)) < maxDMs {
		maxDMs = uint32(len(remindableUsers))
	}
	// encode results
	res, err := json.Marshal(&Reminders{
		RemindableVoters:       remindableUsers,
		RemindableVotersWeight: reminableWeights,
		AlreadySent:            uint32(remindersSent),
		MaxReminders:           maxDMs,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal reminders: %w", err)
	}
	return ctx.Send(res, http.StatusOK)
}

// sendRemindersHandler sends reminders to the voters of an election. It requires
// the user to be the owner of the election. The reminders can be of two types,
// one for a ranked list of n users by weight and another for a single choice of
// users. The reminders are sent in background. The request body must contain the
// type of reminder, the number of users to remind (for the ranked list of users),
// the content of the reminder and the list of users to remind (for the single
// choice of users).
func (v *vocdoniHandler) sendRemindersHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the authenticated user from the token
	token := msg.AuthToken
	if token == "" {
		return fmt.Errorf("missing auth token header")
	}
	auth, err := v.db.UpdateActivityAndGetData(token)
	if err != nil {
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusNotFound)
	}
	// get access profile to use the warpcast api key of the current user
	accessProfile, err := v.db.UserAccessProfile(auth.UserID)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusInternalServerError)
	}
	// check if the user has a configured warpcast api key
	if accessProfile == nil || accessProfile.WarpcastAPIKey == "" {
		return ctx.Send([]byte("no warpcast api key configured"), http.StatusBadRequest)
	}
	// init warpcast client to send the reminders with the user warpcast api key
	warpcastClient := warpcast.NewWarpcastAPI()
	if err := warpcastClient.SetFarcasterUser(auth.UserID, accessProfile.WarpcastAPIKey); err != nil {
		log.Warnw("failed to initialize warpcast client", "error", err)
		return ctx.Send([]byte("failed to initialize warpcast client: "+err.Error()), http.StatusInternalServerError)
	}
	// get the election id from the url params
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	// check that the user is the owner of the election
	election, err := v.db.Election(electionID)
	if err != nil {
		if err == mongo.ErrElectionUnknown {
			return ctx.Send([]byte("election not found"), http.StatusNotFound)
		}
		return fmt.Errorf("failed to get election: %w", err)
	}
	// check if the election is a community election and if the user is an admin
	if election.Community == nil || election.Community.ID == 0 {
		return fmt.Errorf("election is not a community election")
	}
	if !v.db.IsCommunityAdmin(auth.UserID, election.Community.ID) {
		return ctx.Send([]byte("user is not an admin of the community"), http.StatusForbidden)
	}
	// decode the reminders request from the body, there are two types of
	// reminders, one for ranked list of n users by weight and another for
	// single choice of n users
	req := &ReminderRequest{}
	if err := json.Unmarshal(msg.Data, req); err != nil {
		return fmt.Errorf("failed to unmarshal reminders request: %w", err)
	}
	usersToRemind := map[uint64]string{}
	switch req.Type {
	case RankedRemindersType:
		// if the reminder is for a ranked list of users, get the number of users
		// to remind from the request, and get the list of users to remind by weight
		// from the database limited to that number
		if req.NumberOfUsers == 0 {
			return ctx.Send([]byte("missing number of users to remind"), http.StatusBadRequest)
		}
		participants, err := v.db.ParticipantsByWeight(electionID, req.NumberOfUsers)
		if err != nil {
			return fmt.Errorf("failed to get participants by weight: %w", err)
		}
		for username := range participants {
			user, err := v.db.UserByUsername(username)
			if err != nil {
				return fmt.Errorf("failed to get user by username: %w", err)
			}
			usersToRemind[user.UserID] = username
		}
	case IndividualRemindersType:
		// if the reminder is for a individual users, get the list of users fids to
		// remind from the request
		if len(req.Users) == 0 {
			return ctx.Send([]byte("no users to remind"), http.StatusBadRequest)
		}
		usersToRemind = req.Users
	default:
		return ctx.Send([]byte("invalid reminder type"), http.StatusBadRequest)
	}
	// get the remindable users to check if the users to remind are remindable
	remindableUsers, alreadySent, err := v.db.RemindersOfElection(electionID)
	if err != nil {
		return fmt.Errorf("failed to get voters of election: %w", err)
	}
	maxDMs := v.db.MaxDirectMessages(auth.UserID, maxDirectMessages)
	if uint32(len(remindableUsers)) > maxDMs {
		msg := fmt.Sprintf("too many users to remind, by your reputation you only can sent %d reminds", maxDMs)
		return ctx.Send([]byte(msg), http.StatusBadRequest)
	}
	if alreadySent >= maxDMs {
		msg := fmt.Sprintf("you have already sent the maximum number of reminders (%d)", maxDMs)
		return ctx.Send([]byte(msg), http.StatusBadRequest)
	}
	// send the reminders to the users in background
	taskID := util.RandomHex(16)
	v.backgroundQueue.Store(taskID, RemindersStatus{
		Total:      len(usersToRemind),
		ElectionID: election.ElectionID,
	})
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// get the status of the task from the background queue
		s, _ := v.backgroundQueue.Load(taskID)
		currentStatus := s.(RemindersStatus)
		// iterate over the list of users to remind, check if the user is remindable
		// and send the reminder to the user, store the reminded users in a list
		remindsSent := map[uint64]string{}
		for fid, username := range usersToRemind {
			// check if the user is remindable
			if _, ok := remindableUsers[fid]; !ok {
				continue
			}
			// send the reminder to the user
			log.Debugw("sending direct message reminder",
				"content", string(req.Content),
				"to", fid,
				"from", auth.UserID)
			if err := warpcastClient.DirectMessage(ctx, req.Content, fid); err != nil {
				log.Warnw("failed to send direct notification", "error", err, "fid", fid, "username", username)
				currentStatus.Fails[username] = err.Error()
				v.backgroundQueue.Store(taskID, currentStatus)
				continue
			}
			remindsSent[fid] = username
			// update the status of the task
			currentStatus.AlreadySent++
			v.backgroundQueue.Store(taskID, currentStatus)
		}
		// update the already reminded users and the remindable users in the
		// database with the list of reminded users
		if err := v.db.RemindersSent(electionID, remindsSent); err != nil {
			log.Warnf("failed to update reminders: %v", err)
		}
		currentStatus.Completed = true
		v.backgroundQueue.Store(taskID, currentStatus)
	}()
	res, err := json.Marshal(&ReminderResponse{
		QueueID: taskID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal reminders response: %w", err)
	}
	return ctx.Send(res, http.StatusOK)
}

func (v *vocdoniHandler) remindersQueueHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the authenticated user from the token and check if the user is logged in
	token := msg.AuthToken
	if token == "" {
		return fmt.Errorf("missing auth token header")
	}
	if _, err := v.db.UpdateActivityAndGetData(token); err != nil {
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusNotFound)
	}
	// get the election id from the url params
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	// get the queue id from the url params
	queueID := ctx.URLParam("queueID")
	if queueID == "" {
		return ctx.Send([]byte("missing queueID"), http.StatusBadRequest)
	}
	// get the status of the reminders task from the background queue
	status, ok := v.backgroundQueue.Load(queueID)
	if !ok {
		return ctx.Send([]byte("task not found"), http.StatusNotFound)
	}
	currentStatus := status.(RemindersStatus)
	// check if the election match the task
	if currentStatus.ElectionID != hex.EncodeToString(electionID) {
		return ctx.Send([]byte("task does not match the election"), http.StatusBadRequest)
	}
	// check if the task is completed and remove it from the queue
	if currentStatus.Completed {
		v.backgroundQueue.Delete(queueID)
	}
	// encode the status of the task
	res, err := json.Marshal(currentStatus)
	if err != nil {
		return fmt.Errorf("failed to marshal reminders response: %w", err)
	}
	return ctx.Send(res, http.StatusOK)
}

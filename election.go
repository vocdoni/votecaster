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

	"github.com/vocdoni/vote-frame/airstack"
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
	"go.vocdoni.io/dvote/vochain/transaction/proofs/farcasterproof"
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

	// get the user count from different sources (fallback to the total number of addresses)
	req.ElectionDescription.UsersCount = census.FarcasterParticipantCount
	if req.ElectionDescription.UsersCount == 0 {
		req.ElectionDescription.UsersCount = uint32(len(census.Usernames))
		// if no username list is provided, set the users count to the total number
		// this happens in the case of all farcaster users poll and my followers
		if req.ElectionDescription.UsersCount == 0 {
			req.ElectionDescription.UsersCount = uint32(census.FromTotalAddresses)
		}
	}

	// set from total addresses, this information provides the initial number of
	// potential voters in the election, but not all of them might have farcaster account
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
		log.Debugw("set census root to election", "root", req.Census.Root.String(), "election", electionID.String())
		if err := v.db.SetElectionIdForCensusRoot(req.Census.Root, electionID); err != nil {
			log.Errorw(err, fmt.Sprintf("failed to set electionID for census root %s", req.Census.Root))
		}
	}
	// return the electionID
	ctx.SetResponseContentType("application/json")
	return ctx.Send([]byte(electionID.String()), http.StatusOK)
}

func (v *vocdoniHandler) showElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// validate the frame package to airstack
	if v.airstack != nil {
		airstack.ValidateFrameMessage(msg.Data, v.airstack.ApiKey())
	}
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
	// get the election from the cache or the API
	election, err := v.election(electionID)
	if err != nil {
		return fmt.Errorf("failed to fetch election: %w", err)
	}
	// unpack the frame data from the message body
	packet := &FrameSignaturePacket{}
	if err := json.Unmarshal(msg.Data, packet); err != nil {
		return fmt.Errorf("failed to unmarshal frame signature packet: %w", err)
	}

	// check if the user is eligible to vote and extract the vote data
	vote, voteErr := extractVoteDataAndCheckIfEligible(packet, electionID, election.Census.CensusRoot, v.cli)
	if voteErr != nil {
		// check if the user has delegated their vote, if so, return an error
		dbElection, _ := v.db.Election(electionIDbytes)
		if dbElection != nil && dbElection.Community != nil {
			delegations, err := v.db.DelegationsByCommunityFrom(dbElection.Community.ID, uint64(packet.UntrustedData.FID), false)
			if err != nil {
				log.Warnw("failed to fetch delegations", "error", err)
			}
			if len(delegations) > 0 {
				if response, err := handleVoteError(ErrVoteDelegated, vote, electionIDbytes); err != nil {
					ctx.SetResponseContentType("text/html; charset=utf-8")
					return ctx.Send([]byte(response), http.StatusOK)
				}
			}
		}

		// handle the error (if any)
		if response, err := handleVoteError(voteErr, vote, electionIDbytes); err != nil {
			ctx.SetResponseContentType("text/html; charset=utf-8")
			return ctx.Send([]byte(response), http.StatusOK)
		}
	}

	// get the election metadata (question, title, etc.)
	metadata := helpers.UnpackMetadata(election.Metadata)
	png, err := imageframe.QuestionImage(election)
	if err != nil {
		return fmt.Errorf("failed to generate image: %v", err)
	}

	// create the frame state with the electionID, required to verify the vote
	state, err := json.Marshal(&farcasterproof.FarcasterState{
		ProcessID: electionIDbytes,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal farcaster state: %w", err)
	}

	// send the response
	response := strings.ReplaceAll(frame(frameVote), "{image}", imageLink(png))
	response = strings.ReplaceAll(response, "{title}", metadata.Title["default"])
	response = strings.ReplaceAll(response, "{processID}", ctx.URLParam("electionID"))
	response = strings.ReplaceAll(response, "{state}", string(state))

	r := metadata.Questions[0].Choices
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

	// Extract the total participants from the census, try several sources
	totalParticipants := census.FromTotalParticipants
	if totalParticipants == 0 {
		totalParticipants = uint32(len(census.Participants))
		if totalParticipants == 0 {
			totalParticipants = uint32(dbElection.FarcasterUserCount)
		}
		if totalParticipants == 0 {
			totalParticipants = uint32(census.FromTotalAddresses)
		}
	}

	electionInfo := &ElectionInfo{
		CreatedTime:             dbElection.CreatedTime,
		ElectionID:              dbElection.ElectionID,
		LastVoteTime:            dbElection.LastVoteTime,
		EndTime:                 dbElection.EndTime,
		Question:                dbElection.Question,
		CastedVotes:             dbElection.CastedVotes,
		CensusParticipantsCount: uint64(totalParticipants),
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
		Name:        map[string]string{"default": "Farcaster frame proxy"},
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
	customText string, source string, communityID *string,
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
	communityID *string,
) error {
	if election == nil || election.Metadata == nil {
		return fmt.Errorf("invalid election")
	}
	metadata := helpers.UnpackMetadata(election.Metadata)
	if len(metadata.Questions) == 0 {
		return fmt.Errorf("invalid election metadata")
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
		metadata.Title["default"],
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
	if election.Community == nil || election.Community.ID == "" {
		return fmt.Errorf("election is not a community election")
	}
	if !v.db.IsCommunityAdmin(auth.UserID, election.Community.ID) && auth.UserID != v.adminFID {
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
	maxDMs := v.MaxDirectMessages(auth.UserID, maxDirectMessages)
	if uint64(len(remindableUsers)) < maxDMs {
		maxDMs = uint64(len(remindableUsers))
	}
	// encode results
	res, err := json.Marshal(&Reminders{
		RemindableVoters:       remindableUsers,
		RemindableVotersWeight: reminableWeights,
		AlreadySent:            remindersSent,
		MaxReminders:           maxDMs,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal reminders: %w", err)
	}
	return ctx.Send(res, http.StatusOK)
}

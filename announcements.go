package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/vocdoni/vote-frame/farcasterapi/warpcast"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/util"
)

const (
	DefaultAnnouncementTimeSpan = 10 * time.Minute
)

// communityUserProfiles returns a map of user fids and usernames of DAOs
// communities, which census is based on NFTs or ERC20 tokens. The function
// fetches the holders of the community census addresses using census3 and
// then fetches the user profiles from the database based on the addresses
// fetched.
func (v *vocdoniHandler) communityUserProfiles(community *mongo.Community) (map[uint64]string, error) {
	if community.Census.Type != mongo.TypeCommunityCensusNFT &&
		community.Census.Type != mongo.TypeCommunityCensusERC20 {
		return nil, fmt.Errorf("unsupported community census type: %s", community.Census.Type)
	}
	if len(community.Census.Addresses) == 0 {
		return nil, fmt.Errorf("empty community census addresses")
	}
	chainIDs := map[string]uint64{}
	for _, contract := range community.Census.Addresses {
		name := contract.Blockchain
		if name == "ethereum" {
			name = "eth"
		}
		chainID, ok := v.comhub.Census3ChainID(name)
		if !ok {
			log.Warnf("invalid blockchain alias %s for community %s", name, community.ID)
			continue
		}
		chainIDs[contract.Blockchain] = chainID
	}

	// create two goroutines, one to fetch holders from census3 and another
	// to fetch user profiles from the database based on the addresses fetched
	// create a channel to communicate the fetched holders, a list to store the
	// final results and a waitgroup to wait for the goroutines to finish
	communityUsers := make(map[uint64]string)
	holderAddrsCh := make(chan string)
	waiter := sync.WaitGroup{}
	// create a list to store the background errors and a mutex to protect it
	var errsMtx sync.Mutex
	var backgroundErrs []error

	// fetch user profiles from the database based on the addresses fetched
	waiter.Add(1)
	go func() {
		defer waiter.Done()
		for addr := range holderAddrsCh {
			user, err := v.db.UserByAddress(addr)
			if err != nil {
				if !errors.Is(err, mongo.ErrUserUnknown) {
					errsMtx.Lock()
					backgroundErrs = append(backgroundErrs, fmt.Errorf("failed to get user by address (%s): %w", addr, err))
					errsMtx.Unlock()
				}
				continue
			}
			communityUsers[user.UserID] = user.Username
		}
	}()
	// fetch holders from the community census addresses using census3
	waiter.Add(1)
	go func() {
		defer waiter.Done()
		// close the channel when the goroutine finishes to signal the other goroutine
		// to finish
		defer close(holderAddrsCh)
		for _, contractAddr := range community.Census.Addresses {
			chainID, ok := chainIDs[contractAddr.Blockchain]
			if !ok {
				errsMtx.Lock()
				backgroundErrs = append(backgroundErrs, fmt.Errorf("missing chain id for blockchain: %s", contractAddr.Blockchain))
				errsMtx.Unlock()
				continue
			}
			tokenInfo, err := v.census3.Token(contractAddr.Address, chainID, "")
			if err != nil {
				errsMtx.Lock()
				backgroundErrs = append(backgroundErrs, fmt.Errorf("failed to get token info: %w", err))
				errsMtx.Unlock()
				continue
			}
			holdersQueueID, err := v.census3.HoldersByStrategy(tokenInfo.DefaultStrategy, false)
			if err != nil {
				errsMtx.Lock()
				backgroundErrs = append(backgroundErrs, fmt.Errorf("failed to get holders queue id: %w", err))
				errsMtx.Unlock()
				continue
			}
			for {
				holders, finished, err := v.census3.HoldersByStrategyQueue(
					tokenInfo.DefaultStrategy, holdersQueueID)
				if err != nil {
					errsMtx.Lock()
					backgroundErrs = append(backgroundErrs, fmt.Errorf("failed to get holders by strategy queue: %w", err))
					errsMtx.Unlock()
					break
				}
				if finished {
					for holderAddr := range holders {
						holderAddrsCh <- holderAddr.String()
					}
					break
				}
			}
		}
	}()
	// wait for the goroutines to finish
	waiter.Wait()
	// check if there were any background errors
	if len(backgroundErrs) > 0 {
		return nil, fmt.Errorf("failed to get user profiles: %v", backgroundErrs)
	}
	if len(communityUsers) == 0 {
		return nil, fmt.Errorf("no users in the community")
	}
	return communityUsers, nil
}

// usersToAnnounceHandler returns a list of users to announce in a community.
// The list will contain the user fids and usernames of the users in the
// community.
func (v *vocdoniHandler) usersToAnnounceHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	return ctx.Send([]byte("not implemented"), http.StatusNotImplemented)
}

// sendAnnouncementsHandler sends an announcement for the commununity requested
// with the content and to the users specified in the request. The announcement
// is sent via warpcast api, using the api key of the user that sends the
// announcement. The user must be an admin of the community to send the
// announcement. The users to send the announcement to must be part of the
// community. The announcement is sent in background and the status of the task
// can be checked with the queueID returned in the response.
func (v *vocdoniHandler) sendAnnouncementsHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get community id from the URL
	communityID, _, _, err := v.parseCommunityIDFromURL(ctx)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusBadRequest)
	}
	// get the authenticated user from the token
	token := msg.AuthToken
	if token == "" {
		return fmt.Errorf("missing auth token header")
	}
	auth, err := v.db.UpdateActivityAndGetData(token)
	if err != nil {
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusNotFound)
	}
	// get the announcement request from the body and validate it
	req := AnnouncementRequest{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return ctx.Send([]byte("failed to unmarshal announcement request: "+err.Error()), http.StatusBadRequest)
	}
	if req.Content == "" {
		return ctx.Send([]byte("missing content in announcement request"), http.StatusBadRequest)
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
	// get the community from the database
	dbCommunity, err := v.db.Community(communityID)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusInternalServerError)
	}
	if dbCommunity == nil {
		return ctx.Send([]byte("community not found"), http.StatusNotFound)
	}
	// check if the user is admin of the community
	isAdmin := false
	for _, admin := range dbCommunity.Admins {
		if admin == auth.UserID {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		return ctx.Send([]byte("user is not an admin of the community"), http.StatusForbidden)
	}
	// check if the last community announcement is older than the default time span
	if dbCommunity.LastAnnouncement.Add(DefaultAnnouncementTimeSpan).After(time.Now()) {
		return ctx.Send([]byte("last announcement was less than 24 hours ago"), http.StatusBadRequest)
	}
	// init warpcast client to send the reminders with the user warpcast api key
	warpcastClient := warpcast.NewWarpcastAPI()
	if err := warpcastClient.SetFarcasterUser(auth.UserID, accessProfile.WarpcastAPIKey); err != nil {
		return ctx.Send([]byte("failed to initialize warpcast client: "+err.Error()), http.StatusInternalServerError)
	}
	// init the background queue to store the status of the announcement task
	taskID := util.RandomHex(16)
	v.backgroundQueue.Store(taskID, AnnouncementStatus{
		CommunityID: communityID,
		Fails:       make(map[string]string),
	})
	// send the announcement to all the users of the community
	go func() {
		// get the status of the task from the background queue
		s, _ := v.backgroundQueue.Load(taskID)
		currentStatus := s.(AnnouncementStatus)
		// create a context to cancel the task if needed
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// get the list of users of the community
		communityUsers, err := v.communityUserProfiles(dbCommunity)
		if err != nil {
			log.Warnw("failed to get community users", "error", err)
			currentStatus.Error = err.Error()
			v.backgroundQueue.Store(taskID, currentStatus)
			return
		}
		// update the total number of users to send the announcement to
		currentStatus.Total = len(communityUsers) - 1 // exclude the sender
		v.backgroundQueue.Store(taskID, currentStatus)
		// send the announcement to the users in the community
		for fid, username := range communityUsers {
			// skip the user that sends the announcement
			if fid == auth.UserID {
				continue
			}
			// send the announcement to the user via warpcast api
			if err := warpcastClient.DirectMessage(ctx, req.Content, fid); err != nil {
				log.Warnw("failed to send direct notification",
					"error", err,
					"fid", fid,
					"username", username)
				currentStatus.Fails[username] = err.Error()
				v.backgroundQueue.Store(taskID, currentStatus)
				continue
			}
			currentStatus.AlreadySent++
			v.backgroundQueue.Store(taskID, currentStatus)
		}
		currentStatus.Completed = true
		v.backgroundQueue.Store(taskID, currentStatus)
		// update the last announcement time of the community
		if err := v.db.SetCommunityLastAnnouncement(communityID, time.Now()); err != nil {
			log.Warnf("failed to update community last announcement: %v", err)
		}
	}()
	res, err := json.Marshal(&AnnouncementResponse{QueuedID: taskID})
	if err != nil {
		return ctx.Send([]byte("failed to marshal announcement response: "+err.Error()), http.StatusInternalServerError)
	}
	return ctx.Send(res, http.StatusOK)
}

// announcementsQueueHandler returns the status of the announcement task with
// the queueID specified in the URL. The status of the task contains the total
// number of users to send the announcement to, the number of users already
// sent the announcement, the list of users that failed to receive the
// announcement and the error message if the task failed and a flag to indicate
// if the task is completed (with success or failure).
func (v *vocdoniHandler) announcementsQueueHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the authenticated user from the token and check if the user is logged in
	token := msg.AuthToken
	if token == "" {
		return fmt.Errorf("missing auth token header")
	}
	if _, err := v.db.UpdateActivityAndGetData(token); err != nil {
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusNotFound)
	}
	// get community id from the URL
	communityID, _, _, err := v.parseCommunityIDFromURL(ctx)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusBadRequest)
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
	currentStatus := status.(AnnouncementStatus)
	// check if the community match the task
	if currentStatus.CommunityID != communityID {
		return ctx.Send([]byte("task does not match the community"), http.StatusBadRequest)
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

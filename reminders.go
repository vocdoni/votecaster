package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vocdoni/vote-frame/farcasterapi/warpcast"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/util"
)

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
	if election.Community == nil || election.Community.ID == "" {
		return fmt.Errorf("election is not a community election")
	}
	if !v.db.IsCommunityAdmin(auth.UserID, election.Community.ID) && auth.UserID != v.adminFID {
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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vocdoni/vote-frame/mongo"
	"github.com/vocdoni/vote-frame/reputation"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
)

func (v *vocdoniHandler) profileHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	token := msg.AuthToken
	if token == "" {
		return fmt.Errorf("missing auth token header")
	}
	auth, err := v.db.UpdateActivityAndGetData(token)
	if err != nil {
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusNotFound)
	}
	// get user data and access profile
	user, err := v.db.User(auth.UserID)
	if err != nil {
		return ctx.Send([]byte("user not found"), apirest.HTTPstatusNotFound)
	}
	accessprofile, err := v.db.UserAccessProfile(auth.UserID)
	if err != nil {
		return ctx.Send([]byte("could not get user access profile"), apirest.HTTPstatusInternalErr)
	}
	// Get the elections created by the user. If the user is not found, it
	// continues with an empty list.
	userElections, err := v.db.ElectionsByUser(auth.UserID, 16)
	if err != nil && !errors.Is(err, mongo.ErrElectionUnknown) {
		return fmt.Errorf("could not get user elections: %v", err)
	}

	// Get muted users by current user. If the user is not found, it continues
	// with an empty list.
	mutedUsers, err := v.db.ListNotificationMutedUsers(auth.UserID)
	if err != nil && !errors.Is(err, mongo.ErrUserUnknown) {
		return fmt.Errorf("could not get muted users: %v", err)
	}

	// get user delegations
	delegations, err := v.db.DelegationsFrom(auth.UserID)
	if err != nil {
		return fmt.Errorf("could not get user delegations: %v", err)
	}
	// get user reputation
	rep, err := v.db.DetailedUserReputation(auth.UserID)
	if err != nil {
		return fmt.Errorf("could not get user reputation: %v", err)
	}

	// Marshal the response
	data, err := json.Marshal(map[string]any{
		"user":               user,
		"reputation":         reputation.ReputationToAPIResponse(rep),
		"polls":              userElections,
		"mutedUsers":         mutedUsers,
		"delegations":        delegations,
		"warpcastApiEnabled": accessprofile.WarpcastAPIKey != "",
	})
	if err != nil {
		return fmt.Errorf("could not marshal response: %v", err)
	}
	return ctx.Send(data, apirest.HTTPstatusOK)
}

func (v *vocdoniHandler) muteUserHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the authenticated user
	token := msg.AuthToken
	if token == "" {
		return ctx.Send([]byte("missing auth token header"), apirest.HTTPstatusBadRequest)
	}
	auth, err := v.db.UpdateActivityAndGetData(token)
	if err != nil {
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusNotFound)
	}
	// parse the username from the request
	req := map[string]string{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return ctx.Send([]byte("could not parse request"), apirest.HTTPstatusBadRequest)
	}
	usernameToMute, ok := req["username"]
	if !ok {
		return ctx.Send([]byte("missing username"), apirest.HTTPstatusBadRequest)
	}
	// get the user to mute
	userToMute, err := v.db.UserByUsername(usernameToMute)
	if err != nil {
		return ctx.Send([]byte("user not found"), apirest.HTTPstatusNotFound)
	}
	// check if the user is already muted
	isMuted, err := v.db.IsUserNotificationMuted(auth.UserID, userToMute.UserID)
	if err != nil {
		return ctx.Send([]byte("could not check if user is muted"), apirest.HTTPstatusInternalErr)
	}
	// if the user is already muted, return an error
	if isMuted {
		return ctx.Send([]byte("user is already muted"), apirest.HTTPstatusBadRequest)
	}
	// mute the user
	if err := v.db.AddNotificationMutedUser(auth.UserID, userToMute.UserID); err != nil {
		return ctx.Send([]byte("could not mute user"), apirest.HTTPstatusInternalErr)
	}
	return ctx.Send([]byte("Ok"), apirest.HTTPstatusOK)
}

func (v *vocdoniHandler) unmuteUserHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the authenticated user
	token := msg.AuthToken
	if token == "" {
		return ctx.Send([]byte("missing auth token header"), apirest.HTTPstatusBadRequest)
	}
	auth, err := v.db.UpdateActivityAndGetData(token)
	if err != nil {
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusNotFound)
	}
	// get the muted username from the request
	mutedUsername := ctx.URLParam("username")
	if mutedUsername == "" {
		return ctx.Send([]byte("missing username"), apirest.HTTPstatusBadRequest)
	}
	// get the muted user from the database
	mutedUser, err := v.db.UserByUsername(mutedUsername)
	if err != nil {
		return ctx.Send([]byte("user not found"), apirest.HTTPstatusNotFound)
	}
	// check if the user is muted
	isMuted, err := v.db.IsUserNotificationMuted(auth.UserID, mutedUser.UserID)
	if err != nil {
		return ctx.Send([]byte("could not check if user is muted"), apirest.HTTPstatusInternalErr)
	}
	// if the user is not muted, return an error
	if !isMuted {
		return ctx.Send([]byte("user is not muted"), apirest.HTTPstatusBadRequest)
	}
	// unmute the user
	if err := v.db.DelNotificationMutedUser(auth.UserID, mutedUser.UserID); err != nil {
		return ctx.Send([]byte("could not mute user"), apirest.HTTPstatusInternalErr)
	}
	return ctx.Send([]byte("Ok"), apirest.HTTPstatusOK)
}

func (v *vocdoniHandler) delegateVoteHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the authenticated user
	token := msg.AuthToken
	if token == "" {
		return ctx.Send([]byte("missing auth token header"), apirest.HTTPstatusBadRequest)
	}
	auth, err := v.db.UpdateActivityAndGetData(token)
	if err != nil {
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusNotFound)
	}
	// parse the username from the request
	req := &mongo.Delegation{}
	if err := json.Unmarshal(msg.Data, req); err != nil {
		return ctx.Send([]byte("could not parse request"), apirest.HTTPstatusBadRequest)
	}
	// check if the required fields are present
	if req.To == 0 || req.CommuniyID == "" {
		return ctx.Send([]byte("missing required fields"), apirest.HTTPstatusBadRequest)
	}
	req.From = auth.UserID
	// check if the user is trying to delegate to themselves
	if req.From == req.To {
		return ctx.Send([]byte("cannot delegate to yourself"), apirest.HTTPstatusBadRequest)
	}
	// check if the user is trying to delegate to a non-existing user
	_, err = v.db.User(req.To)
	if err != nil {
		return ctx.Send([]byte("failied to get user to delegate to"), apirest.HTTPstatusInternalErr)
	}
	// check if the user is trying to delegate to a non-existing community
	_, err = v.db.Community(req.CommuniyID)
	if err != nil {
		return ctx.Send([]byte("failied to get community to delegate to"), apirest.HTTPstatusInternalErr)
	}
	// delegate the vote
	if _, err := v.db.SetDelegation(req); err != nil {
		return ctx.Send([]byte("could not delegate vote"), apirest.HTTPstatusInternalErr)
	}
	return ctx.Send([]byte("Ok"), apirest.HTTPstatusOK)
}

func (v *vocdoniHandler) removeVoteDelegationHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the authenticated user
	token := msg.AuthToken
	if token == "" {
		return ctx.Send([]byte("missing auth token header"), apirest.HTTPstatusBadRequest)
	}
	auth, err := v.db.UpdateActivityAndGetData(token)
	if err != nil {
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusNotFound)
	}
	// get the delegation ID from the request and retrieve the delegation from
	// the database
	delegationID := ctx.URLParam("delegationID")
	delegation, err := v.db.Delegation(delegationID)
	if err != nil {
		return ctx.Send([]byte("delegation not found"), apirest.HTTPstatusNotFound)
	}
	// check if the user is trying to remove a delegation that does not belong to
	// them
	if delegation.From != auth.UserID {
		return ctx.Send([]byte("delegation does not belong to user"), apirest.HTTPstatusBadRequest)
	}
	// remove the delegation
	if err := v.db.DeleteDelegation(delegationID); err != nil {
		return ctx.Send([]byte("could not remove delegation"), apirest.HTTPstatusInternalErr)
	}
	return ctx.Send([]byte("Ok"), apirest.HTTPstatusOK)
}

func (v *vocdoniHandler) profilePublicHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	var user *mongo.User
	var err error

	// Get the user by username if provided
	if handle := ctx.URLParam("userHandle"); handle != "" {
		user, err = v.db.UserByUsername(handle)
		if err != nil {
			return ctx.Send([]byte("user not found"), apirest.HTTPstatusNotFound)
		}
	}

	// Else get the user by FID
	if user == nil {
		if ctx.URLParam("fid") == "" {
			return ctx.Send([]byte("missing user handle or fid"), apirest.HTTPstatusBadRequest)
		}
		fid, err := strconv.ParseUint(ctx.URLParam("fid"), 10, 64)
		if err != nil {
			return ctx.Send([]byte("invalid fid"), apirest.HTTPstatusBadRequest)
		}
		user, err = v.db.User(fid)
		if err != nil {
			return ctx.Send([]byte("user not found"), apirest.HTTPstatusNotFound)
		}
	}

	// Get the elections created by the user. If the user is not found, it
	// continues with an empty list.
	userElections, err := v.db.ElectionsByUser(user.UserID, 16)
	if err != nil && !errors.Is(err, mongo.ErrElectionUnknown) {
		return fmt.Errorf("could not get user elections: %v", err)
	}

	// Get muted users by current user. If the user is not found, it continues
	// with an empty list.
	mutedUsers, err := v.db.ListNotificationMutedUsers(user.UserID)
	if err != nil && !errors.Is(err, mongo.ErrUserUnknown) {
		return fmt.Errorf("could not get muted users: %v", err)
	}

	// get user reputation
	rep, err := v.repUpdater.UserReputation(user.UserID, true)
	if err != nil {
		return fmt.Errorf("could not get user reputation: %v", err)
	}

	// Marshal the response
	data, err := json.Marshal(map[string]any{
		"user":       user,
		"reputation": rep,
		"polls":      userElections,
		"mutedUsers": mutedUsers,
	})
	if err != nil {
		return fmt.Errorf("could not marshal response: %v", err)
	}
	return ctx.Send(data, apirest.HTTPstatusOK)
}

func (v *vocdoniHandler) registerWarpcastApiKey(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	token := msg.AuthToken
	if token == "" {
		return fmt.Errorf("missing auth token header")
	}
	auth, err := v.db.UpdateActivityAndGetData(token)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusNotFound)
	}
	// decode the api key
	var apiKey WarpcastAPIKey
	if err := json.Unmarshal(msg.Data, &apiKey); err != nil {
		return ctx.Send([]byte("could not parse request"), apirest.HTTPstatusBadRequest)
	}
	// store the api key
	if err := v.db.SetWarpcastAPIKey(auth.UserID, apiKey.APIKey); err != nil {
		return ctx.Send([]byte("could not store api key: "+err.Error()), http.StatusInternalServerError)
	}
	return ctx.Send([]byte("ok"), apirest.HTTPstatusOK)
}

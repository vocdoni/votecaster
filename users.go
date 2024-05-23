package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vocdoni/vote-frame/mongo"
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
	// Get and update the user's reputation
	reputation, reputationData, err := v.db.UpdateAndGetReputationForUser(auth.UserID)
	if err != nil {
		return fmt.Errorf("could not update reputation: %v", err)
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

	// Marshal the response
	data, err := json.Marshal(map[string]any{
		"user":               user,
		"reputation":         reputation,
		"reputationData":     reputationData,
		"polls":              userElections,
		"mutedUsers":         mutedUsers,
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

	// Get and update the user's reputation
	reputation, reputationData, err := v.db.UpdateAndGetReputationForUser(user.UserID)
	if err != nil {
		return fmt.Errorf("could not update reputation: %v", err)
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

	// Marshal the response
	data, err := json.Marshal(map[string]any{
		"user":           user,
		"reputation":     reputation,
		"reputationData": reputationData,
		"polls":          userElections,
		"mutedUsers":     mutedUsers,
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

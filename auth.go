package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/vocdoni/vote-frame/farcasterauth"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
)

// authChannels is a map of existing authentication channels.
var authChannels sync.Map

// authLinkHandler creates an authentication channel and returns the URL and ID.
func (v *vocdoniHandler) authLinkHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	c := farcasterauth.New()
	resp, err := c.CreateChannel(serverURL)
	if err != nil {
		return fmt.Errorf("could not create authentication channel: %v", err)
	}
	data, err := json.Marshal(
		map[string]string{
			"url": resp.URL,
			"id":  resp.Nonce,
		},
	)
	if err != nil {
		return fmt.Errorf("could not marshal response: %v", err)
	}
	if _, ok := authChannels.Load(resp.Nonce); ok {
		return fmt.Errorf("channel already exists")
	}
	authChannels.Store(resp.Nonce, c)
	log.Debugw("authentication channel created", "id", resp.Nonce, "url", resp.URL)
	return ctx.Send(data, apirest.HTTPstatusOK)
}

// authVerifyHandler verifies the authentication channel and returns the auth token.
func (v *vocdoniHandler) authVerifyHandler(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	nonce := ctx.URLParam("id")
	if nonce == "" {
		return fmt.Errorf("missing id parameter")
	}
	c, ok := authChannels.Load(nonce)
	if !ok {
		return ctx.Send(nil, apirest.HTTPstatusNotFound)
	}
	resp, err := c.(*farcasterauth.Client).CheckStatus()
	if err != nil {
		return ctx.Send(nil, apirest.HTTPstatusNoContent)
	}
	defer authChannels.Delete(nonce)
	token, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("could not generate token: %v", err)
	}

	// Get and update the user's reputation
	reputation, reputationData, err := v.db.UpdateAndGetReputationForUser(resp.Fid)
	if err != nil {
		return fmt.Errorf("could not update reputation: %v", err)
	}

	// Get the elections created by the user. If the user is not found, it
	// continues with an empty list.
	userElections, err := v.db.ElectionsByUser(resp.Fid)
	if err != nil && !errors.Is(err, mongo.ErrElectionUnknown) {
		return fmt.Errorf("could not get user elections: %v", err)
	}

	// Get muted users by current user. If the user is not found, it continues
	// with an empty list.
	mutedUsers, err := v.db.ListNotificationMutedUsers(resp.Fid)
	if err != nil && !errors.Is(err, mongo.ErrUserUnknown) {
		return fmt.Errorf("could not get muted users: %v", err)
	}

	// Remove unnecessary fields
	resp.State = ""
	resp.Nonce = ""
	resp.Message = ""
	resp.Signature = ""

	// Marshal the response
	data, err := json.Marshal(map[string]any{
		"profile":        resp,
		"authToken":      token.String(),
		"reputation":     reputation,
		"reputationData": reputationData,
		"elections":      userElections,
		"mutedUsers":     mutedUsers,
	})
	if err != nil {
		return fmt.Errorf("could not marshal response: %v", err)
	}
	v.addAuthTokenFunc(resp.Fid, token.String())

	log.Infow("authentication completed", "username", resp.Username, "fid", resp.Fid, "reputation")
	return ctx.Send(data, apirest.HTTPstatusOK)
}

// authCheckHandler checks if the auth token is valid and updates the activity time.
func (v *vocdoniHandler) authCheckHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	token := msg.AuthToken
	if token == "" {
		return fmt.Errorf("missing auth token header")
	}
	auth, err := v.db.UpdateActivityAndGetData(token)
	if err != nil {
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusNotFound)
	}

	// Get and update the user's reputation
	reputation, reputationData, err := v.db.UpdateAndGetReputationForUser(auth.UserID)
	if err != nil {
		return fmt.Errorf("could not update reputation: %v", err)
	}

	// Get the elections created by the user. If the user is not found, it
	// continues with an empty list.
	userElections, err := v.db.ElectionsByUser(auth.UserID)
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
		"reputation":     reputation,
		"reputationData": reputationData,
		"elections":      userElections,
		"mutedUsers":     mutedUsers,
	})
	if err != nil {
		return fmt.Errorf("could not marshal response: %v", err)
	}
	log.Infow("authentication check completed, updated reputation", "fid", auth.UserID, "reputation", reputation)
	return ctx.Send(data, apirest.HTTPstatusOK)
}

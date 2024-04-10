package main

import (
	"encoding/json"
	"errors"
	"fmt"

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
		"polls":          userElections,
		"mutedUsers":     mutedUsers,
	})
	if err != nil {
		return fmt.Errorf("could not marshal response: %v", err)
	}
	return ctx.Send(data, apirest.HTTPstatusOK)
}

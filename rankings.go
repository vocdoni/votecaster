package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
)

func (v *vocdoniHandler) rankingByElectionsCreated(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	users, err := v.db.UsersByElectionNumber()
	if err != nil {
		return fmt.Errorf("failed to get ranking: %w", err)
	}
	jresponse, err := json.Marshal(map[string]any{
		"users": users,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.SetResponseContentType("application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

func (v *vocdoniHandler) rankingByVotes(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	users, err := v.db.UsersByVoteNumber()
	if err != nil {
		return fmt.Errorf("failed to get ranking: %w", err)
	}
	jresponse, err := json.Marshal(map[string]any{
		"users": users,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.SetResponseContentType("application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

func (v *vocdoniHandler) rankingOfElections(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	elections, err := v.db.ElectionsByVoteNumber()
	if err != nil {
		return fmt.Errorf("failed to get ranking: %w", err)
	}
	jresponse, err := json.Marshal(map[string]any{
		"polls": elections,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.SetResponseContentType("application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

func (v *vocdoniHandler) lastElectionsHandler(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	dbElections, err := v.db.LastCreatedElections(10)
	if err != nil {
		return fmt.Errorf("failed to get last elections: %w", err)
	}

	type Election struct {
		CreatedTime             time.Time `json:"createdTime"`
		ElectionID              string    `json:"electionId"`
		LastVoteTime            time.Time `json:"lastVoteTime"`
		Question                string    `json:"title"`
		CastedVotes             uint64    `json:"voteCount"`
		CensusParticipantsCount uint64    `json:"censusParticipantsCount"`
		Username                string    `json:"createdByUsername"`
		Displayname             string    `json:"createdByDisplayname"`
	}

	var elections []*Election

	for i := range dbElections {
		var username, displayname string
		user, err := v.db.User(dbElections[i].UserID)
		if err != nil {
			log.Warnw("failed to fetch user", "error", err)
			username = "unknown"
		} else {
			username = user.Username
			displayname = user.Displayname
		}
		elections = append(elections, &Election{
			dbElections[i].CreatedTime,
			dbElections[i].ElectionID,
			dbElections[i].LastVoteTime,
			dbElections[i].Question,
			dbElections[i].CastedVotes,
			uint64(dbElections[i].FarcasterUserCount),
			username,
			displayname,
		})
	}
	jresponse, err := json.Marshal(map[string]any{
		"polls": elections,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.SetResponseContentType("application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

func (v *vocdoniHandler) rankingByReputation(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	users, err := v.db.UserByReputation()
	if err != nil {
		return fmt.Errorf("failed to get ranking: %w", err)
	}
	jresponse, err := json.Marshal(map[string]any{
		"users": users,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.SetResponseContentType("application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

func (v *vocdoniHandler) electionsByCommunityHandler(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// TODO: we should limit and paginate this call
	id, err := strconv.ParseUint(ctx.URLParam("communityID"), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse community ID: %w", err)
	}

	dbElections, err := v.db.ElectionsByCommunity(id)
	if err != nil {
		return fmt.Errorf("failed to get elections for community %d: %w", id, err)
	}
	var elections []*ElectionInfo

	for i := range dbElections {
		var username, displayname string
		user, err := v.db.User(dbElections[i].UserID)
		if err != nil {
			log.Warnw("failed to fetch user", "error", err)
			username = "unknown"
		} else {
			username = user.Username
			displayname = user.Displayname
		}
		elections = append(elections, &ElectionInfo{
			CreatedTime:             dbElections[i].CreatedTime,
			EndTime:                 dbElections[i].EndTime,
			ElectionID:              dbElections[i].ElectionID,
			LastVoteTime:            dbElections[i].LastVoteTime,
			Question:                dbElections[i].Question,
			CastedVotes:             dbElections[i].CastedVotes,
			CensusParticipantsCount: uint64(dbElections[i].FarcasterUserCount),
			Username:                username,
			Displayname:             displayname,
			Finalized:               time.Now().After(dbElections[i].EndTime), // return true if EndTime is in the past
		})
	}
	jresponse, err := json.Marshal(map[string]any{
		"polls": elections,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.SetResponseContentType("application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

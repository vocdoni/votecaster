package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vocdoni/vote-frame/mongo"
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
		Turnout                 float32   `json:"turnout"`
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
			dbElections[i].Turnout,
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
	type Election struct {
		CreatedTime             time.Time         `json:"createdTime"`
		ElectionID              string            `json:"electionId"`
		LastVoteTime            time.Time         `json:"lastVoteTime"`
		Question                string            `json:"title"`
		CastedVotes             uint64            `json:"voteCount"`
		CensusParticipantsCount uint64            `json:"censusParticipantsCount"`
		Turnout                 float32           `json:"turnout"`
		Username                string            `json:"createdByUsername"`
		Displayname             string            `json:"createdByDisplayname"`
		TotalWeight             string            `json:"totalWeight"`
		Participants            map[string]string `json:"participants"`
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
		electionIDBytes, err := hex.DecodeString(dbElections[i].ElectionID)
		if err != nil {
			log.Errorw(err, "cannot decode electionID")
			continue
		}
		census, err := v.db.CensusFromElection(electionIDBytes)
		if err != nil {
			log.Warnw("census not found for community election", "electionID", dbElections[i].ElectionID)
			census = &mongo.Census{
				TotalWeight:  "0",
				Participants: make(map[string]string),
			}
		}
		elections = append(elections, &Election{
			dbElections[i].CreatedTime,
			dbElections[i].ElectionID,
			dbElections[i].LastVoteTime,
			dbElections[i].Question,
			dbElections[i].CastedVotes,
			uint64(dbElections[i].FarcasterUserCount),
			dbElections[i].Turnout,
			username,
			displayname,
			census.TotalWeight,
			census.Participants,
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

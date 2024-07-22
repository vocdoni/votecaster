package main

import (
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
	dbElections, err := v.db.ElectionsByVoteNumber()
	if err != nil {
		return fmt.Errorf("failed to get ranking: %w", err)
	}
	// decode the elections to the response format
	var elections []*RankedElection
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
		var community *Community
		if dbElections[i].Community != nil {
			dbCommunity, err := v.db.Community(dbElections[i].Community.ID)
			if err != nil {
				log.Warnw("failed to fetch community", "error", err)
			} else if dbCommunity != nil {
				community = &Community{
					ID:            dbCommunity.ID,
					Name:          dbCommunity.Name,
					LogoURL:       dbCommunity.ImageURL,
					GroupChatURL:  dbCommunity.GroupChatURL,
					Notifications: dbCommunity.Notifications,
					Channels:      dbCommunity.Channels,
				}
			}
		}

		elections = append(elections, &RankedElection{
			dbElections[i].CreatedTime,
			dbElections[i].ElectionID,
			dbElections[i].LastVoteTime,
			dbElections[i].Question,
			dbElections[i].CastedVotes,
			uint64(dbElections[i].FarcasterUserCount),
			username,
			displayname,
			community,
		})
	}
	// encode the response to json including pagination information
	jresponse, err := json.Marshal(RankedElections{
		Elections: elections,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.SetResponseContentType("application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

func (v *vocdoniHandler) latestElectionsHandler(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get optional parameters to paginate the results
	limit := maxPaginatedItems
	strLimit := ctx.Request.URL.Query().Get("limit")
	if strLimit != "" {
		// if the query has the limit parameter, list the first n elections
		n, err := strconv.Atoi(strLimit)
		if err != nil {
			return ctx.Send([]byte("invalid limit"), http.StatusBadRequest)
		}
		limit = int64(n)
		if limit > maxPaginatedItems {
			limit = maxPaginatedItems
		} else if limit < minPaginatedItems {
			limit = minPaginatedItems
		}
	}
	offset := int64(0)
	strOffset := ctx.Request.URL.Query().Get("offset")
	if strOffset != "" {
		// if the query has the offset parameter, list elections starting from
		// the n th election
		n, err := strconv.Atoi(strOffset)
		if err != nil {
			return ctx.Send([]byte("invalid offset"), http.StatusBadRequest)
		}
		offset = int64(n)
	}
	// get elections from the database
	dbElections, total, err := v.db.LatestElections(limit, offset)
	if err != nil {
		return fmt.Errorf("failed to get latest elections: %w", err)
	}
	// decode the elections to the response format
	var elections []*RankedElection
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
		var community *Community
		if dbElections[i].Community != nil {
			dbCommunity, err := v.db.Community(dbElections[i].Community.ID)
			if err != nil {
				log.Warnw("failed to fetch community", "error", err)
			} else if dbCommunity != nil {
				community = &Community{
					ID:            dbCommunity.ID,
					Name:          dbCommunity.Name,
					LogoURL:       dbCommunity.ImageURL,
					GroupChatURL:  dbCommunity.GroupChatURL,
					Notifications: dbCommunity.Notifications,
					Channels:      dbCommunity.Channels,
				}
			}
		}

		elections = append(elections, &RankedElection{
			dbElections[i].CreatedTime,
			dbElections[i].ElectionID,
			dbElections[i].LastVoteTime,
			dbElections[i].Question,
			dbElections[i].CastedVotes,
			uint64(dbElections[i].FarcasterUserCount),
			username,
			displayname,
			community,
		})
	}
	// encode the response to json including pagination information
	jresponse, err := json.Marshal(RankedElections{
		Elections: elections,
		Pagination: &Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
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
	// get community id from the URL
	communityID, _, _, err := v.parseCommunityIDFromURL(ctx)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusBadRequest)
	}

	dbElections, err := v.db.ElectionsByCommunity(communityID)
	if err != nil {
		return fmt.Errorf("failed to get elections for community %s: %w", communityID, err)
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
			FID:                     dbElections[i].UserID,
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

func (v *vocdoniHandler) rankingByPoints(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	_, onlyUsers := ctx.Request.URL.Query()["onlyUsers"]
	_, onlyCommunities := ctx.Request.URL.Query()["onlyCommunities"]
	if onlyUsers && onlyCommunities {
		return ctx.Send([]byte("invalid query parameters"), http.StatusBadRequest)
	}
	var err error
	var ranking []mongo.ReputationRanking
	switch {
	case onlyUsers:
		if ranking, err = v.db.ReputationRanking(true, false); err != nil {
			return fmt.Errorf("failed to get ranking: %w", err)
		}
	case onlyCommunities:
		if ranking, err = v.db.ReputationRanking(false, true); err != nil {
			return fmt.Errorf("failed to get ranking: %w", err)
		}
	default:
		log.Info("both")
		if ranking, err = v.db.ReputationRanking(true, true); err != nil {
			return fmt.Errorf("failed to get ranking: %w", err)
		}
	}
	for i, item := range ranking {
		if item.CommunityID != "" {
			community, err := v.db.Community(item.CommunityID)
			if err != nil {
				log.Warnw("failed to fetch community", "error", err)
			} else if community != nil {
				ranking[i].ImageURL = community.ImageURL
			}
		} else if item.UserID > 0 {
			user, err := v.db.User(item.UserID)
			if err != nil {
				log.Warnw("failed to fetch user", "error", err)
			} else if user != nil {
				ranking[i].ImageURL = user.Avatar
			}
		}
	}

	jresponse, err := json.Marshal(map[string]any{
		"points": ranking,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.SetResponseContentType("application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

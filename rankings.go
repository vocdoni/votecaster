package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vocdoni/vote-frame/mongo"
	"github.com/vocdoni/vote-frame/reputation"
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
	var rankingItems []mongo.ReputationRanking
	switch {
	case onlyUsers:
		if rankingItems, err = v.db.ReputationRanking(true, false); err != nil {
			return fmt.Errorf("failed to get ranking: %w", err)
		}
	case onlyCommunities:
		if rankingItems, err = v.db.ReputationRanking(false, true); err != nil {
			return fmt.Errorf("failed to get ranking: %w", err)
		}
	default:
		log.Info("both")
		if rankingItems, err = v.db.ReputationRanking(true, true); err != nil {
			return fmt.Errorf("failed to get ranking: %w", err)
		}
	}
	// iterate over ranking items and fetch the user or community image and
	// skip the disabled communities
	var ranking []mongo.ReputationRanking
	for _, item := range rankingItems {
		if item.CommunityID != "" {
			community, err := v.db.Community(item.CommunityID)
			if err != nil {
				log.Warnw("failed to fetch community", "error", err)
			} else if community != nil && !community.Disabled {
				item.ImageURL = community.ImageURL
				ranking = append(ranking, item)
			}
		} else if item.UserID > 0 {
			user, err := v.db.User(item.UserID)
			if err != nil {
				log.Warnw("failed to fetch user", "error", err)
			} else if user != nil {
				item.ImageURL = user.Avatar
				ranking = append(ranking, item)
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

func (v *vocdoniHandler) rankingByYieldRate(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	reps, err := v.db.CommunitiesReputationByParticipationAndCensusSize()
	if err != nil {
		return fmt.Errorf("failed to get ranking: %w", err)
	}
	var rankingItems []mongo.ReputationRanking
	for _, rep := range reps {
		// get community info from reputation community id
		community, err := v.db.Community(rep.CommunityID)
		if err != nil {
			log.Warnw("failed to fetch community", "error", err)
			continue
		}
		// skip disabled communities
		if community == nil || community.Disabled {
			continue
		}
		// get creator reputation
		creatorRep, err := v.db.DetailedUserReputation(community.Creator)
		if err != nil {
			log.Warnw("failed to fetch creator reputation", "error", err)
			continue
		}
		// calculate the yield rate for the community
		isDao := community.Census.Type == mongo.TypeCommunityCensusERC20 || community.Census.Type == mongo.TypeCommunityCensusNFT
		isChannel := community.Census.Type == mongo.TypeCommunityCensusChannel
		yieldRate := reputation.CommunityYieldRate(rep.Participation, float64(rep.CensusSize),
			float64(creatorRep.TotalReputation), isDao, isChannel)
		// add the community to the ranking
		rankingItems = append(rankingItems, mongo.ReputationRanking{
			CommunityID:      community.ID,
			CommunityName:    community.Name,
			ImageURL:         community.ImageURL,
			CommunityCreator: community.Creator,
			TotalPoints:      uint64(yieldRate),
		})
	}
	// encode the response to json
	res, err := json.Marshal(map[string]any{
		"yieldRates": rankingItems,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.SetResponseContentType("application/json")
	return ctx.Send(res, http.StatusOK)
}

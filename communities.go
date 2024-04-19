package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
)

// censusChannelOrAddresses gets the census channel or addresses based on the
// type of the census provided from the database. If the census provided is
// based on a channel, it gets the channel information from the farcaster API,
// and returns a nil for the addresses and the channel information. If the
// census provided is based on addresses, it converts the address from the
// database to the API format and returns them, with a nil for the channel
// information.
func (v *vocdoniHandler) censusChannelOrAddresses(ctx context.Context,
	dbCensus mongo.CommunityCensus,
) ([]*CensusAddress, *Channel, error) {
	var censusChannel *Channel
	var censusAddresses []*CensusAddress
	switch dbCensus.Type {
	case mongo.TypeCommunityCensusChannel:
		channel, err := v.fcapi.Channel(ctx, dbCensus.Channel)
		if err != nil {
			return nil, nil, err
		}
		censusChannel = &Channel{
			ID:          channel.ID,
			Name:        channel.Name,
			Description: channel.Description,
			Followers:   channel.Followers,
			ImageURL:    channel.Image,
			URL:         channel.URL,
		}
	case mongo.TypeCommunityCensusERC20, mongo.TypeCommunityCensusNFT:
		censusAddresses = []*CensusAddress{}
		if len(dbCensus.Addresses) > 0 {
			for _, addr := range dbCensus.Addresses {
				censusAddresses = append(censusAddresses, &CensusAddress{
					Address:    addr.Address,
					Blockchain: addr.Blockchain,
				})
			}
		}
	default:
		return nil, nil, fmt.Errorf("invalid census type")
	}
	return censusAddresses, censusChannel, nil
}

func (v *vocdoniHandler) listCommunitiesHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	var err error
	var dbCommunities []mongo.Community
	// check if the query has the byAdminFID or byAdminUsername parameters and
	// list communities by admin FID or username respectively, otherwise list
	// all communities
	byAdminFID := ctx.Request.URL.Query().Get("byAdminFID")
	byAdminUsername := ctx.Request.URL.Query().Get("byAdminUsername")
	switch {
	case byAdminFID != "":
		// if the query has the byAdminFID parameter, list communities by admin FID
		var adminFID int
		adminFID, err = strconv.Atoi(byAdminFID)
		if err != nil {
			return ctx.Send([]byte("invalid admin FID"), http.StatusBadRequest)
		}
		if dbCommunities, err = v.db.ListCommunitiesByAdminFID(uint64(adminFID)); err != nil {
			return ctx.Send([]byte("error listing communities"), http.StatusInternalServerError)
		}
	case byAdminUsername != "":
		// if the query has the byAdminUsername parameter, list communities by admin username
		if dbCommunities, err = v.db.ListCommunitiesByAdminUsername(byAdminUsername); err != nil {
			return ctx.Send([]byte("error listing communities"), http.StatusInternalServerError)
		}
	default:
		// otherwise, list all communities
		if dbCommunities, err = v.db.ListCommunities(); err != nil {
			return ctx.Send([]byte("error listing communities"), http.StatusInternalServerError)
		}
	}
	if len(dbCommunities) == 0 {
		return ctx.Send([]byte("no communities found"), http.StatusNotFound)
	}
	communities := CommunityList{
		Communities: []*Community{},
	}
	for _, c := range dbCommunities {
		// get admin profiles from the database
		admins := []*User{}
		for _, admin := range c.Admins {
			user, err := v.db.User(admin)
			if err != nil {
				if err == farcasterapi.ErrNoDataFound ||
					strings.Contains(err.Error(), "user unknown") {
					log.Warnw("community admin not found in the database",
						"err", err,
						"user", admin,
						"community", c.ID)
					admins = append(admins, &User{FID: admin})
					continue
				}
				return ctx.Send([]byte(err.Error()), http.StatusInternalServerError)
			}
			admins = append(admins, &User{
				FID:         user.UserID,
				Username:    user.Username,
				DisplayName: user.Displayname,
				Avatar:      user.Avatar,
			})
		}
		// get census channel or addresses based on the type
		cAddresses, cChannel, err := v.censusChannelOrAddresses(ctx.Request.Context(), c.Census)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if err == farcasterapi.ErrChannelNotFound {
				statusCode = http.StatusNotFound
			}
			return ctx.Send([]byte(err.Error()), statusCode)
		}
		// add community to the list
		communities.Communities = append(communities.Communities, &Community{
			ID:              c.ID,
			Name:            c.Name,
			LogoURL:         c.ImageURL,
			GroupChatURL:    c.GroupChatURL,
			Admins:          admins,
			Notifications:   c.Notifications,
			CensusType:      c.Census.Type,
			CensusAddresses: cAddresses,
			CensusChannel:   cChannel,
			Channels:        c.Channels,
		})
	}
	res, err := json.Marshal(communities)
	if err != nil {
		return ctx.Send([]byte("error encoding communities"), http.StatusInternalServerError)
	}
	return ctx.Send(res, http.StatusOK)
}

func (v *vocdoniHandler) communityHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get community id from the URL and parse to int
	communityID := ctx.URLParam("communityID")
	if communityID == "" {
		return ctx.Send([]byte("no community ID provided"), http.StatusBadRequest)
	}
	id, err := strconv.Atoi(communityID)
	if err != nil {
		return ctx.Send([]byte("invalid community ID"), http.StatusBadRequest)
	}
	// get the community from the database by its id
	dbCommunity, err := v.db.Community(uint64(id))
	if err != nil {
		return ctx.Send([]byte("error getting community"), http.StatusInternalServerError)
	}
	if dbCommunity == nil {
		return ctx.Send([]byte("community not found"), http.StatusNotFound)
	}
	// get admin profiles for the community
	admins := []*User{}
	for _, admin := range dbCommunity.Admins {
		user, err := v.db.User(admin)
		if err != nil {
			if err == farcasterapi.ErrNoDataFound {
				log.Warnw("community admin not found in the database",
					"err", err,
					"user", admin,
					"community", dbCommunity.ID)
				continue
			}
			return ctx.Send([]byte(err.Error()), http.StatusInternalServerError)
		}
		admins = append(admins, &User{
			FID:         user.UserID,
			Username:    user.Username,
			DisplayName: user.Displayname,
			Avatar:      user.Avatar,
		})
	}
	// get census channel or addresses based on the type
	cAddresses, cChannel, err := v.censusChannelOrAddresses(ctx.Request.Context(), dbCommunity.Census)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == farcasterapi.ErrChannelNotFound {
			statusCode = http.StatusNotFound
		}
		return ctx.Send([]byte(err.Error()), statusCode)
	}
	// encode the community
	res, err := json.Marshal(Community{
		ID:              dbCommunity.ID,
		Name:            dbCommunity.Name,
		LogoURL:         dbCommunity.ImageURL,
		GroupChatURL:    dbCommunity.GroupChatURL,
		Admins:          admins,
		Notifications:   dbCommunity.Notifications,
		CensusType:      dbCommunity.Census.Type,
		CensusAddresses: cAddresses,
		CensusChannel:   cChannel,
		Channels:        dbCommunity.Channels,
	})
	if err != nil {
		return ctx.Send([]byte("error encoding community"), http.StatusInternalServerError)
	}
	return ctx.Send(res, http.StatusOK)
}

// disableCommunityHanler allows to an admin of a community to disable it.
func (v *vocdoniHandler) disableCommunityHanler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// extract userFID from auth token
	userFID, err := v.db.UserFromAuthToken(msg.AuthToken)
	if err != nil {
		return fmt.Errorf("cannot get user from auth token: %w", err)
	}
	// get community id from the URL and parse to int
	communityID := ctx.URLParam("communityID")
	if communityID == "" {
		return ctx.Send([]byte("no community ID provided"), http.StatusBadRequest)
	}
	id, err := strconv.Atoi(communityID)
	if err != nil {
		return ctx.Send([]byte("invalid community ID"), http.StatusBadRequest)
	}
	// get the community from the database by its id
	community, err := v.db.Community(uint64(id))
	if err != nil {
		return ctx.Send([]byte("error getting community"), http.StatusInternalServerError)
	}
	if community == nil {
		return ctx.Send([]byte("community not found"), http.StatusNotFound)
	}
	// check if the current user is an admin of the community
	var authorized bool
	for _, admin := range community.Admins {
		if admin == userFID {
			authorized = true
			break
		}
	}
	if !authorized {
		return ctx.Send([]byte("you are not an admin of this community"), http.StatusUnauthorized)
	}
	// disable the community in the community hub contract and delete it from
	// the database
	if err := v.comhub.DisableCommunity(community.ID); err != nil {
		return ctx.Send([]byte("error disabling community"), http.StatusInternalServerError)
	}
	if err := v.db.DelCommunity(community.ID); err != nil {
		return ctx.Send([]byte("error deleting community"), http.StatusInternalServerError)
	}
	return ctx.Send([]byte("ok"), http.StatusOK)
}

package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
)

func (v *vocdoniHandler) listCommunitiesHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	urlQuery := ctx.Request.URL.Query()
	var err error
	var dbCommunities []mongo.Community
	if byAdminFID := urlQuery.Get("byAdminFID"); byAdminFID != "" {
		var adminFID int
		adminFID, err = strconv.Atoi(byAdminFID)
		if err != nil {
			return ctx.Send([]byte("Invalid admin FID"), http.StatusBadRequest)
		}
		if dbCommunities, err = v.db.ListCommunitiesByAdminFID(uint64(adminFID)); err != nil {
			return ctx.Send([]byte("Error listing communities"), http.StatusInternalServerError)
		}
	} else if byAdminUsername := urlQuery.Get("byAdminUsername"); byAdminUsername != "" {
		if dbCommunities, err = v.db.ListCommunitiesByAdminUsername(byAdminUsername); err != nil {
			return ctx.Send([]byte("Error listing communities"), http.StatusInternalServerError)
		}
	} else {
		if dbCommunities, err = v.db.ListCommunities(); err != nil {
			return ctx.Send([]byte("Error listing communities"), http.StatusInternalServerError)
		}
	}
	if len(dbCommunities) == 0 {
		return ctx.Send([]byte("No communities found"), http.StatusNotFound)
	}
	communities := CommunityList{
		Communities: []*Community{},
	}
	for _, c := range dbCommunities {
		// get admin profiles
		admins := []*FarcasterProfile{}
		for _, admin := range c.Admins {
			user, err := v.fcapi.UserDataByFID(ctx.Request.Context(), admin)
			if err != nil {
				if err == farcasterapi.ErrNoDataFound {
					continue
				}
				return ctx.Send([]byte(err.Error()), apirest.HTTPstatusInternalErr)
			}
			admins = append(admins, &FarcasterProfile{
				FID:           user.FID,
				Username:      user.Username,
				DisplayName:   user.Displayname,
				Avatar:        user.Avatar,
				Bio:           user.Bio,
				Custody:       user.CustodyAddress,
				Verifications: user.VerificationsAddresses,
			})
		}
		// get census channel or addresses based on the type
		var censusChannel *Channel
		var censusAddresses []*CensusAddress
		if c.Census.Type == mongo.TypeCommunityCensusChannel {
			channel, err := v.fcapi.Channel(ctx.Request.Context(), c.Census.Channel)
			if err != nil {
				if err == farcasterapi.ErrChannelNotFound {
					return ctx.Send([]byte("Census channel not found"), http.StatusNotFound)
				}
				return ctx.Send([]byte(err.Error()), apirest.HTTPstatusInternalErr)
			}
			censusChannel = &Channel{
				ID:          channel.ID,
				Name:        channel.Name,
				Description: channel.Description,
				Followers:   channel.Followers,
				ImageURL:    channel.Image,
				URL:         channel.URL,
			}
		} else { // ERC20 or NFT census type
			censusAddresses = []*CensusAddress{}
			if len(c.Census.Addresses) > 0 {
				for _, addr := range c.Census.Addresses {
					censusAddresses = append(censusAddresses, &CensusAddress{
						Address:    addr.Address,
						Blockchain: addr.Blockchain,
					})
				}
			}
		}
		// get channels details
		channels := []*Channel{}
		for _, ch := range c.Channels {
			channel, err := v.fcapi.Channel(ctx.Request.Context(), ch)
			if err != nil {
				if err == farcasterapi.ErrChannelNotFound {
					continue
				}
				return ctx.Send([]byte(err.Error()), apirest.HTTPstatusInternalErr)
			}
			channels = append(channels, &Channel{
				ID:          channel.ID,
				Name:        channel.Name,
				Description: channel.Description,
				Followers:   channel.Followers,
				ImageURL:    channel.Image,
				URL:         channel.URL,
			})
		}
		// add community to the list
		communities.Communities = append(communities.Communities, &Community{
			ID:              c.ID,
			Name:            c.Name,
			LogoURL:         c.ImageURL,
			Admins:          admins,
			Notifications:   c.Notifications,
			CensusName:      c.Census.Name,
			CensusType:      c.Census.Type,
			CensusAddresses: censusAddresses,
			CensusChannel:   censusChannel,
			Channels:        channels,
		})
	}
	res, err := json.Marshal(communities)
	if err != nil {
		return ctx.Send([]byte("Error encoding communities"), http.StatusInternalServerError)
	}
	return ctx.Send(res, http.StatusOK)
}

func (v *vocdoniHandler) getCommunityHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	communityID := ctx.URLParam("communityID")
	if communityID == "" {
		return ctx.Send([]byte("No community ID provided"), http.StatusBadRequest)
	}
	id, err := strconv.Atoi(communityID)
	if err != nil {
		return ctx.Send([]byte("Invalid community ID"), http.StatusBadRequest)
	}
	dbCommunity, err := v.db.Community(uint64(id))
	if err != nil {
		return ctx.Send([]byte("Error getting community"), http.StatusInternalServerError)
	}
	if dbCommunity == nil {
		return ctx.Send([]byte("Community not found"), http.StatusNotFound)
	}
	// get admin profiles
	admins := []*FarcasterProfile{}
	for _, admin := range dbCommunity.Admins {
		user, err := v.fcapi.UserDataByFID(ctx.Request.Context(), admin)
		if err != nil {
			if err == farcasterapi.ErrNoDataFound {
				continue
			}
			return ctx.Send([]byte(err.Error()), apirest.HTTPstatusInternalErr)
		}
		admins = append(admins, &FarcasterProfile{
			FID:           user.FID,
			Username:      user.Username,
			DisplayName:   user.Displayname,
			Avatar:        user.Avatar,
			Bio:           user.Bio,
			Custody:       user.CustodyAddress,
			Verifications: user.VerificationsAddresses,
		})
	}
	// get census channel or addresses based on the type
	var censusChannel *Channel
	var censusAddresses []*CensusAddress
	if dbCommunity.Census.Type == "channel" {
		channel, err := v.fcapi.Channel(ctx.Request.Context(), dbCommunity.Census.Channel)
		if err != nil {
			if err == farcasterapi.ErrChannelNotFound {
				return ctx.Send([]byte("Census channel not found"), http.StatusNotFound)
			}
			return ctx.Send([]byte(err.Error()), apirest.HTTPstatusInternalErr)
		}
		censusChannel = &Channel{
			ID:          channel.ID,
			Name:        channel.Name,
			Description: channel.Description,
			Followers:   channel.Followers,
			ImageURL:    channel.Image,
			URL:         channel.URL,
		}
	} else if dbCommunity.Census.Type == "erc20" || dbCommunity.Census.Type == "nft" {
		censusAddresses = []*CensusAddress{}
		if len(dbCommunity.Census.Addresses) > 0 {
			for _, addr := range dbCommunity.Census.Addresses {
				censusAddresses = append(censusAddresses, &CensusAddress{
					Address:    addr.Address,
					Blockchain: addr.Blockchain,
				})
			}
		}
	}
	// get channels details
	channels := []*Channel{}
	for _, ch := range dbCommunity.Channels {
		channel, err := v.fcapi.Channel(ctx.Request.Context(), ch)
		if err != nil {
			if err == farcasterapi.ErrChannelNotFound {
				continue
			}
			return ctx.Send([]byte(err.Error()), apirest.HTTPstatusInternalErr)
		}
		channels = append(channels, &Channel{
			ID:          channel.ID,
			Name:        channel.Name,
			Description: channel.Description,
			Followers:   channel.Followers,
			ImageURL:    channel.Image,
			URL:         channel.URL,
		})
	}
	// encode the community
	res, err := json.Marshal(Community{
		ID:              dbCommunity.ID,
		Name:            dbCommunity.Name,
		LogoURL:         dbCommunity.ImageURL,
		Admins:          admins,
		Notifications:   dbCommunity.Notifications,
		CensusName:      dbCommunity.Census.Name,
		CensusType:      dbCommunity.Census.Type,
		CensusAddresses: censusAddresses,
		CensusChannel:   censusChannel,
		Channels:        channels,
	})
	if err != nil {
		return ctx.Send([]byte("Error encoding community"), http.StatusInternalServerError)
	}
	return ctx.Send(res, http.StatusOK)
}

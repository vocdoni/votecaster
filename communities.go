package main

import (
	"encoding/json"
	"net/http"

	"github.com/vocdoni/vote-frame/farcasterapi"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
)

func (v *vocdoniHandler) listCommunitiesHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	dbCommunities, err := v.db.ListCommunities()
	if err != nil {
		return ctx.Send([]byte("Error listing communities"), http.StatusInternalServerError)
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
		// get census addresses details
		censusAddresses := []CensusAddress{}
		for _, addr := range c.Census.Addresses {
			censusAddresses = append(censusAddresses, CensusAddress{
				Address:    addr.Address,
				Blockchain: addr.Blockchain,
			})
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
	return nil
}

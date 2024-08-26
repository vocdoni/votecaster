package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vocdoni/vote-frame/communityhub"
	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
)

func (v *vocdoniHandler) parseCommunityIDFromURL(ctx *httprouter.HTTPContext) (string, string, uint64, error) {
	// get community id from the URL
	strID := ctx.URLParam("communityID")
	if strID == "" {
		return "", "", 0, fmt.Errorf("no community ID provided")
	}
	// check if the community ID is prefixed and decode it
	if _, prefixedID, ok := communityhub.DecodePrefix(strID); ok {
		strID = prefixedID
	}
	id, err := strconv.ParseUint(strID, 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid community ID: %w", err)
	}
	// get community chain from the URL
	chainAlias := ctx.URLParam("chainAlias")
	if chainAlias == "" {
		return "", "", 0, fmt.Errorf("no community chain short name provided")
	}
	communityID, ok := v.comhub.CommunityIDByChainAlias(id, chainAlias)
	if !ok {
		return "", "", 0, fmt.Errorf("invalid community ID")
	}
	return communityID, chainAlias, id, nil
}

// CommunityStatus method checks the status of a community based on the census
// type. If the census type is not based on ERC20 or NFT, it returns true and
// 100% progress, because different types of census do not require syncing, they
// depend on external sources. If the census type is based on ERC20 or NFT, it
// iterates over the addresses of the community getting the status of the token
// in census3. It returns false if any token is not synced, and the average
// progress of all tokens.
func (v *vocdoniHandler) CommunityStatus(community *mongo.Community) (bool, int, error) {
	// return error if the community is nil
	if community == nil {
		return false, 0, fmt.Errorf("community not found")
	}
	// return true if the community is not based on ERC20 or NFT
	if community.Census.Type != mongo.TypeCommunityCensusERC20 &&
		community.Census.Type != mongo.TypeCommunityCensusNFT {
		return true, 100, nil
	}
	synced := true
	progress := 0
	nTokens := len(community.Census.Addresses)
	// iterate over the addresses of the community getting status of token in census3
	for _, contract := range community.Census.Addresses {
		chainID, ok := v.comhub.Census3ChainID(contract.Blockchain)
		if !ok {
			return false, 0, fmt.Errorf("invalid blockchain alias")
		}
		tokenInfo, err := v.census3.Token(contract.Address, chainID, "")
		if err != nil {
			return false, 0, fmt.Errorf("error getting token info: %w", err)
		}
		if tokenInfo == nil {
			return false, 0, fmt.Errorf("token not found")
		}
		synced = synced && tokenInfo.Status.Synced
		progress += tokenInfo.Status.Progress / nTokens
	}
	return synced, progress, nil
}

// censusChannelOrAddresses gets the census channel or addresses based on the
// type of the census provided from the database. If the census provided is
// based on a channel, it gets the channel information from the farcaster API,
// and returns a nil for the addresses and the channel information. If the
// census provided is based on addresses, it converts the address from the
// database to the API format and returns them, with a nil for the channel
// information.
func (v *vocdoniHandler) censusChannelOrAddresses(ctx context.Context,
	dbCensus mongo.CommunityCensus,
) ([]*CensusAddress, *Channel, *User, error) {
	var censusChannel *Channel
	var censusAddresses []*CensusAddress
	var user *User
	switch dbCensus.Type {
	case mongo.TypeCommunityCensusFollowers:
		fid, err := communityhub.DecodeUserChannelFID(dbCensus.Channel)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("invalid user reference: %w", err)
		}
		dbUser, err := v.db.User(fid)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error getting user: %w", err)
		}
		if dbUser == nil {
			return nil, nil, nil, fmt.Errorf("user not found")
		}
		user = &User{
			FID:         dbUser.UserID,
			Username:    dbUser.Username,
			DisplayName: dbUser.Displayname,
			Avatar:      dbUser.Avatar,
		}
	case mongo.TypeCommunityCensusChannel:
		channel, err := v.fcapi.Channel(ctx, dbCensus.Channel)
		if err != nil {
			return nil, nil, nil, err
		}
		if channel == nil {
			return nil, nil, nil, farcasterapi.ErrChannelNotFound
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
		return nil, nil, nil, fmt.Errorf("invalid census type")
	}
	return censusAddresses, censusChannel, user, nil
}

func (v *vocdoniHandler) listCommunitiesHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	var err error
	var dbCommunities []mongo.Community
	// check if the query has the byAdminFID or byAdminUsername parameters and
	// list communities by admin FID or username respectively, otherwise list
	// all communities
	byAdminFID := ctx.Request.URL.Query().Get("byAdminFID")
	byAdminUsername := ctx.Request.URL.Query().Get("byAdminUsername")
	featured := ctx.Request.URL.Query().Get("featured")
	// get optional parameters to paginate the results
	limit := maxPaginatedItems
	strLimit := ctx.Request.URL.Query().Get("limit")
	if strLimit != "" {
		// if the query has the limit parameter, list the first n communities
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
		// if the query has the offset parameter, list communities starting from
		// the n th community
		n, err := strconv.Atoi(strOffset)
		if err != nil {
			return ctx.Send([]byte("invalid offset"), http.StatusBadRequest)
		}
		offset = int64(n)
	}
	var totalCommunities int64
	switch {
	case byAdminFID != "":
		// if the query has the byAdminFID parameter, list communities by admin FID
		var adminFID int
		adminFID, err = strconv.Atoi(byAdminFID)
		if err != nil {
			return ctx.Send([]byte("invalid admin FID"), http.StatusBadRequest)
		}
		if dbCommunities, totalCommunities, err = v.db.ListCommunitiesByAdminFID(uint64(adminFID), limit, offset); err != nil {
			return ctx.Send([]byte("error listing communities"), http.StatusInternalServerError)
		}
	case byAdminUsername != "":
		// if the query has the byAdminUsername parameter, list communities by admin username
		if dbCommunities, totalCommunities, err = v.db.ListCommunitiesByAdminUsername(byAdminUsername, limit, offset); err != nil {
			return ctx.Send([]byte("error listing communities"), http.StatusInternalServerError)
		}
	case featured == "true":
		// if the query has the featured parameter, list featured communities
		if dbCommunities, totalCommunities, err = v.db.ListFeaturedCommunities(limit, offset); err != nil {
			return ctx.Send([]byte("error listing communities"), http.StatusInternalServerError)
		}
	default:
		// otherwise, list all communities
		if dbCommunities, totalCommunities, err = v.db.ListCommunities(limit, offset); err != nil {
			return ctx.Send([]byte("error listing communities"), http.StatusInternalServerError)
		}
	}
	if len(dbCommunities) == 0 {
		return ctx.Send([]byte("no communities found"), http.StatusNotFound)
	}
	communities := CommunityList{
		Communities: []*Community{},
		Pagination: &Pagination{
			Limit:  limit,
			Offset: offset,
			Total:  totalCommunities,
		},
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
		// get census channel, addresses or user reference based on the type
		cAddresses, cChannel, userRef, err := v.censusChannelOrAddresses(ctx.Request.Context(), c.Census)
		if err != nil && err != farcasterapi.ErrChannelNotFound {
			return ctx.Send([]byte(err.Error()), http.StatusInternalServerError)
		}
		// check if the community is ready (soft check, if it fails, continue)
		ready, _, err := v.CommunityStatus(&c)
		if err != nil {
			log.Warnw("error getting community status", "err", err, "community", c.ID)
		}
		// add community to the list
		communities.Communities = append(communities.Communities, &Community{
			ID:                   c.ID,
			Name:                 c.Name,
			LogoURL:              c.ImageURL,
			GroupChatURL:         c.GroupChatURL,
			Admins:               admins,
			Notifications:        c.Notifications,
			CensusType:           c.Census.Type,
			CensusAddresses:      cAddresses,
			CensusChannel:        cChannel,
			UserRef:              userRef,
			Channels:             c.Channels,
			Disabled:             c.Disabled,
			Ready:                ready,
			CanSendAnnouncements: c.LastAnnouncement.Add(DefaultAnnouncementTimeSpan).Before(time.Now()),
		})
	}
	res, err := json.Marshal(communities)
	if err != nil {
		return ctx.Send([]byte("error encoding communities"), http.StatusInternalServerError)
	}
	return ctx.Send(res, http.StatusOK)
}

func (v *vocdoniHandler) communityHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get community id from the URL
	communityID, _, _, err := v.parseCommunityIDFromURL(ctx)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusBadRequest)
	}
	// get the community from the database by its id
	dbCommunity, err := v.db.Community(communityID)
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
	cAddresses, cChannel, userRef, err := v.censusChannelOrAddresses(ctx.Request.Context(), dbCommunity.Census)
	if err != nil && err != farcasterapi.ErrChannelNotFound {
		return ctx.Send([]byte(err.Error()), http.StatusInternalServerError)
	}
	// check if the community is ready (hard check, if it fails return an error)
	ready, _, err := v.CommunityStatus(dbCommunity)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusInternalServerError)
	}
	// encode the community
	res, err := json.Marshal(Community{
		ID:                   dbCommunity.ID,
		Name:                 dbCommunity.Name,
		LogoURL:              dbCommunity.ImageURL,
		GroupChatURL:         dbCommunity.GroupChatURL,
		Admins:               admins,
		Notifications:        dbCommunity.Notifications,
		CensusType:           dbCommunity.Census.Type,
		CensusAddresses:      cAddresses,
		CensusChannel:        cChannel,
		UserRef:              userRef,
		Channels:             dbCommunity.Channels,
		Disabled:             dbCommunity.Disabled,
		Ready:                ready,
		CanSendAnnouncements: dbCommunity.LastAnnouncement.Add(DefaultAnnouncementTimeSpan).Before(time.Now()),
	})
	if err != nil {
		return ctx.Send([]byte("error encoding community"), http.StatusInternalServerError)
	}
	return ctx.Send(res, http.StatusOK)
}

func (v *vocdoniHandler) communityStatusHandler(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get community id from the URL
	communityID, _, _, err := v.parseCommunityIDFromURL(ctx)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusBadRequest)
	}
	// get the community from the database by its id
	dbCommunity, err := v.db.Community(communityID)
	if err != nil {
		return ctx.Send([]byte("error getting community"), http.StatusInternalServerError)
	}
	if dbCommunity == nil {
		return ctx.Send([]byte("community not found"), http.StatusNotFound)
	}
	// get the status of the community
	ready, progress, err := v.CommunityStatus(dbCommunity)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusInternalServerError)
	}
	// encode the status
	res, err := json.Marshal(Community{
		Ready:    ready,
		Progress: progress,
	})
	if err != nil {
		return ctx.Send([]byte("error encoding community status"), http.StatusInternalServerError)
	}
	return ctx.Send(res, http.StatusOK)
}

// communitySettingsHandler allows to an admin of a community to update the
// community information.
func (v *vocdoniHandler) communitySettingsHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// extract userFID from auth token
	userFID, err := v.db.UserFromAuthToken(msg.AuthToken)
	if err != nil {
		return fmt.Errorf("cannot get user from auth token: %w", err)
	}
	// get community id from the URL
	communityID, chainAlias, contractID, err := v.parseCommunityIDFromURL(ctx)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusBadRequest)
	}
	chainID, ok := v.comhub.ChainIDFromAlias(chainAlias)
	if !ok {
		return ctx.Send([]byte("invalid community chain alias provided"), http.StatusBadRequest)
	}
	// get the community from the database by its id
	dbCommunity, err := v.db.Community(communityID)
	if err != nil {
		return ctx.Send([]byte("error getting community"), http.StatusInternalServerError)
	}
	if dbCommunity == nil {
		return ctx.Send([]byte("community not found"), http.StatusNotFound)
	}
	// check if the current user is an admin of the community
	var authorized bool
	for _, admin := range dbCommunity.Admins {
		if admin == userFID {
			authorized = true
			break
		}
	}
	if !authorized {
		return ctx.Send([]byte("you are not an admin of this community"), http.StatusUnauthorized)
	}
	var typedCommunity Community
	if err := json.Unmarshal(msg.Data, &typedCommunity); err != nil {
		return ctx.Send([]byte("error decoding community data"), http.StatusBadRequest)
	}
	// check optional booleans fields from a map to avoid setting them to false
	// if they are not provided
	var mapCommunity map[string]interface{}
	if err := json.Unmarshal(msg.Data, &mapCommunity); err != nil {
		return ctx.Send([]byte("error decoding community data"), http.StatusBadRequest)
	}
	notification := &dbCommunity.Notifications
	if _, ok := mapCommunity["notifications"]; ok {
		*notification = typedCommunity.Notifications
	}
	disabled := &dbCommunity.Disabled
	if _, ok := mapCommunity["disabled"]; ok {
		*disabled = typedCommunity.Disabled
	}
	// parse the admins and census addresses
	admins := []uint64{}
	for _, user := range typedCommunity.Admins {
		admins = append(admins, user.FID)
	}
	// parse the census channel or addresses based on the census type
	var censusChannel string
	censusAddresses := []*communityhub.ContractAddress{}
	switch communityhub.CensusType(typedCommunity.CensusType) {
	case communityhub.CensusTypeERC20, communityhub.CensusTypeNFT:
		for _, addr := range typedCommunity.CensusAddresses {
			censusAddresses = append(censusAddresses, &communityhub.ContractAddress{
				Blockchain: addr.Blockchain,
				Address:    common.HexToAddress(addr.Address),
			})
		}
	case communityhub.CensusTypeFollowers:
		censusChannel = communityhub.EncodeUserChannelFID(userFID)
	case communityhub.CensusTypeChannel:
		if typedCommunity.CensusChannel != nil {
			censusChannel = typedCommunity.CensusChannel.ID
		}
	default:
		return ctx.Send([]byte("invalid census type"), http.StatusBadRequest)
	}
	// update the community image
	if typedCommunity.LogoURL != "" && typedCommunity.LogoURL != dbCommunity.ImageURL {
		// check if the current avatar is an internal image
		avatarID, isInternalAvatar := avatarIDfromURL(dbCommunity.ImageURL)
		// if is internal delete the current avatar from the database after
		// uploading the new one
		if isInternalAvatar {
			if err := v.db.RemoveAvatar(avatarID); err != nil {
				log.Warnw("error deleting avatar", "err", err, "avatarID", avatarID)
			}
		}
		// upload the new avatar if it is base64 encoded
		if isBase64Image(typedCommunity.LogoURL) {
			// empty the avatarID to generate a new one based on the data
			avatarURL, err := v.uploadAvatar("", userFID, communityID, typedCommunity.LogoURL)
			if err != nil {
				return fmt.Errorf("cannot upload avatar: %w", err)
			}
			// set the new avatar URL
			typedCommunity.LogoURL = avatarURL
		}
	}
	// update the community in the community hub
	if err := v.comhub.UpdateCommunity(&communityhub.HubCommunity{
		CommunityID:    communityID,
		ContractID:     contractID,
		ChainID:        chainID,
		Name:           typedCommunity.Name,
		ImageURL:       typedCommunity.LogoURL,
		GroupChatURL:   typedCommunity.GroupChatURL,
		CensusType:     communityhub.CensusType(typedCommunity.CensusType),
		CensusAddesses: censusAddresses,
		CensusChannel:  censusChannel,
		Channels:       typedCommunity.Channels,
		Admins:         admins,
		Notifications:  notification,
		Disabled:       disabled,
	}); err != nil {
		return fmt.Errorf("error updating community: %w", err)
	}
	return ctx.Send([]byte("ok"), http.StatusOK)
}

func (v *vocdoniHandler) communityDelegationsHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get community id from the URL
	communityID, _, _, err := v.parseCommunityIDFromURL(ctx)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusBadRequest)
	}
	delegations, err := v.db.DelegationsByCommunity(communityID)
	if err != nil {
		return ctx.Send([]byte("error getting delegations"), http.StatusInternalServerError)
	}
	if len(delegations) == 0 {
		return ctx.Send(nil, http.StatusNoContent)
	}
	res, err := json.Marshal(delegations)
	if err != nil {
		return ctx.Send([]byte("error encoding delegations"), http.StatusInternalServerError)
	}
	return ctx.Send(res, http.StatusOK)
}

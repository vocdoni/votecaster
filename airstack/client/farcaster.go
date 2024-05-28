package client

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	gql "github.com/vocdoni/vote-frame/airstack/graphql"
)

// FarcasterUser wraps useful information of a Farcaster user.
type FarcasterUser struct {
	FID          string
	EVMAddresses []common.Address
	ProfileName  string
}

// farcasterUsersWithAssociatedAddresses is a wrapper around the generated function
// for GraphQL query farcasterUsersWithAssociatedAddresses.
func (c *Client) farcasterUsersWithAssociatedAddresses(
	limit int,
	cursor string,
) (*gql.GetFarcasterUsersWithAssociatedAddressesResponse, error) {
	cctx, cancel := context.WithTimeout(c.ctx, apiTimeout)
	defer cancel()
	r := 0
	var err error
	var resp *gql.GetFarcasterUsersWithAssociatedAddressesResponse
	for r < maxAPIRetries {
		resp, err = gql.GetFarcasterUsersWithAssociatedAddresses(cctx, c.Client, limit, cursor)
		if err != nil {
			r += 1
			time.Sleep(time.Second * 3)
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("max GraphQL retries reached, error: %w", err)
}

// FarcasterUsersWithAssociatedAddresses gets all the Farcaster users ids with their
// associated EVM addresses calling the Airstack API. This function also takes care of Airstack API pagination.
func (c *Client) FarcasterUsersWithAssociatedAddresses() ([]*FarcasterUser, error) {
	hasNextPage := true
	cursor := ""
	fu := make([]*FarcasterUser, 0)
	for hasNextPage {
		resp, err := c.farcasterUsersWithAssociatedAddresses(airstackAPIlimit, cursor)
		if err != nil {
			return nil, fmt.Errorf("cannot get users from Airstack: %w", err)
		}
		for _, u := range resp.Socials.Social {
			fu = append(fu, &FarcasterUser{
				FID:          u.UserId,
				EVMAddresses: u.UserAssociatedAddresses,
			})
		}
		cursor = resp.Socials.PageInfo.NextCursor
		hasNextPage = cursor != ""
	}
	return fu, nil
}

// farcasterUsersByChannel is a wrapper around the generated function
// for GraphQL query FarcasterUsersByChannel.
func (c *Client) farcasterUsersByChannel(
	channelName string, limit int, cursor string,
) (*gql.GetFarcasterUsersByChannelResponse, error) {
	cctx, cancel := context.WithTimeout(c.ctx, apiTimeout)
	defer cancel()
	r := 0
	var err error
	var resp *gql.GetFarcasterUsersByChannelResponse
	for r < maxAPIRetries {
		resp, err = gql.GetFarcasterUsersByChannel(cctx, c.Client, channelName, limit, cursor)
		if err != nil {
			r += 1
			time.Sleep(time.Second * 3)
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("max GraphQL retries reached, error: %w", err)
}

// FarcasterUsersByChannel gets all the Farcaster user ids of a given channel
// calling the Airstack API. This function also takes care of Airstack API pagination.
func (c *Client) FarcasterUsersByChannel(channelId string) ([]*FarcasterUser, error) {
	hasNextPage := true
	cursor := ""
	fuser := make([]*FarcasterUser, 0)
	for hasNextPage {
		resp, err := c.farcasterUsersByChannel(channelId, airstackAPIlimit, cursor)
		if err != nil {
			return nil, fmt.Errorf("cannot get channel users id from Airstack: %w", err)
		}
		for _, fcc := range resp.FarcasterChannels.FarcasterChannel {
			for _, participant := range fcc.Participants {
				p := participant.GetParticipant()
				fuser = append(fuser, &FarcasterUser{
					FID:          p.Fid,
					ProfileName:  p.ProfileName,
					EVMAddresses: p.UserAssociatedAddresses,
				})
			}
		}
		cursor = resp.FarcasterChannels.PageInfo.NextCursor
		hasNextPage = cursor != ""
	}
	return fuser, nil
}

// FarcasterUserFollowerCount gets the number of followers of a given Farcaster user.
func (c *Client) FarcasterUserFollowerCount(farcasterUserId string) (int, error) {
	followersCount, err := c.farcasterUserFollowersCount(farcasterUserId)
	if err != nil {
		return 0, fmt.Errorf("cannot get followers count from Airstack: %w", err)
	}
	return followersCount, nil
}

// farcasterUserFollowersCount is a wrapper around the GQL generated function GetFarcasterUserFollowers
func (c *Client) farcasterUserFollowersCount(userId string) (int, error) {
	cctx, cancel := context.WithTimeout(c.ctx, apiTimeout)
	defer cancel()
	r := 0
	var err error
	var resp *gql.GetFarcasterUserFollowersResponse
	for r < maxAPIRetries {
		resp, err = gql.GetFarcasterUserFollowers(cctx, c.Client, userId)
		if err != nil {
			r += 1
			time.Sleep(time.Second * 3)
			continue
		}
		return resp.Socials.Social[0].FollowerCount, nil
	}
	return 0, fmt.Errorf("max GraphQL retries reached, error: %w", err)
}

// FarcasterCheckIfUserIsFollowing checks if a user is following another user.
func (c *Client) FarcasterCheckIfUserIsFollowing(followerUserId, followedUserId string) (bool, error) {
	isFollowing, err := c.farcasterUserIsFollowing(followerUserId, followedUserId)
	if err != nil {
		return false, fmt.Errorf("cannot check if user is following another user: %w", err)
	}
	return isFollowing, nil
}

// farcasterUserIsFollowing is a wrapper around the GQL generated function CheckFarcasterFollowing
func (c *Client) farcasterUserIsFollowing(followerUserId, followedUserId string) (bool, error) {
	cctx, cancel := context.WithTimeout(c.ctx, apiTimeout)
	defer cancel()
	r := 0
	var err error
	var resp *gql.CheckFarcasterFollowingResponse
	// convert followerUserId and followedUserId to expected internal string format
	// "fc_fid:followerUserId" and "fc_fid:followedUserId"
	followerUserId = fmt.Sprintf("fc_fid:%s", followerUserId)
	followedUserId = fmt.Sprintf("fc_fid:%s", followedUserId)
	for r < maxAPIRetries {
		resp, err = gql.CheckFarcasterFollowing(cctx, c.Client, followerUserId, followedUserId)
		if err != nil {
			r += 1
			time.Sleep(time.Second * 3)
			continue
		}
		return resp.SocialFollowings.Following[0].FollowingAddress.Socials[0].ProfileName == followedUserId, nil
	}
	return false, fmt.Errorf("max GraphQL retries reached, error: %w", err)
}

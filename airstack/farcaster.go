package airstack

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	gql "github.com/vocdoni/vote-frame/airstack/graphql"
)

// FarcasterUser wraps useful information of a Farcaster user.
type FarcasterUser struct {
	FID          string
	EVMAddresses []common.Address
	ProfileName  string
}

// getFarcasterUsersWithAssociatedAddresses is a wrapper around the generated function
// for GraphQL query GetFarcasterUsersWithAssociatedAddresses.
func (c *Client) getFarcasterUsersWithAssociatedAddresses(
	limit int,
	cursor string,
) (*gql.GetFarcasterUsersWithAssociatedAddressesResponse, error) {
	return gql.GetFarcasterUsersWithAssociatedAddresses(c.ctx, c.Client, limit, cursor)
}

// GetFarcasterUsersWithAssociatedAddresses gets all the Farcaster users ids with their
// associated EVM addresses calling the Airstack API. This function also takes care of Airstack API pagination.
func (c *Client) GetFarcasterUsersWithAssociatedAddresses() ([]*FarcasterUser, error) {
	hasNextPage := true
	cursor := ""
	fu := make([]*FarcasterUser, 0)
	for hasNextPage {
		resp, err := c.getFarcasterUsersWithAssociatedAddresses(airstackAPIlimit, cursor)
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
		if resp.Socials.PageInfo.NextCursor == "" {
			hasNextPage = false
		}
	}
	return fu, nil
}

// getFarcasterUsersByChannel is a wrapper around the generated function
// for GraphQL query GetFarcasterUsersByChannel.
func (c *Client) getFarcastersUsersByChannel(
	channelName string, limit int, cursor string,
) (*gql.GetFarcasterUsersByChannelResponse, error) {
	cctx, cancel := context.WithTimeout(c.ctx, apiTimeout)
	defer cancel()
	return gql.GetFarcasterUsersByChannel(cctx, c.Client, channelName, limit, cursor)
}

// GetFarcasterUsersByChannel gets all the Farcaster user ids of a given channel
// calling the Airstack API. This function also takes care of Airstack API pagination.
func (c *Client) GetFarcasterUsersByChannel(channelId string) ([]*FarcasterUser, error) {
	hasNextPage := true
	cursor := ""
	fuser := make([]*FarcasterUser, 0)
	for hasNextPage {
		resp, err := c.getFarcastersUsersByChannel(channelId, airstackAPIlimit, cursor)
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
		if resp.FarcasterChannels.PageInfo.NextCursor == "" {
			hasNextPage = false
		}
	}
	return fuser, nil
}

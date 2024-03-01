package airstack

import (
	"context"
	"fmt"

	gql "github.com/vocdoni/vote-frame/airstack/graphql"
)

// getFarcasterUsersWithAssociatedAddresses is a wrapper around the generated function
// for GraphQL query GetFarcasterUsersWithAssociatedAddresses.
func (c *Client) getFarcasterUsersWithAssociatedAddresses(
	limit int,
	cursor string,
) (*gql.GetFarcasterUsersWithAssociatedAddressesResponse, error) {
	return gql.GetFarcasterUsersWithAssociatedAddresses(c.ctx, c.Client, limit, cursor)
}

// TODO GetFarcasterUsersWithAssociatedAddresses implement pagination

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
func (c *Client) GetFarcasterUsersByChannel(channelId string) ([]string, error) {
	hasNextPage := true
	cursor := ""
	fids := make([]string, 0)
	for hasNextPage {
		resp, err := c.getFarcastersUsersByChannel(channelId, airstackAPIlimit, cursor)
		if err != nil {
			return nil, fmt.Errorf("cannot channel users id from Airstack: %w", err)
		}
		for _, fcc := range resp.FarcasterChannels.FarcasterChannel {
			for _, fid := range fcc.Participants {
				fids = append(fids, fid.ParticipantId)
			}
		}
		cursor = resp.FarcasterChannels.PageInfo.NextCursor
		if resp.FarcasterChannels.PageInfo.NextCursor == "" {
			hasNextPage = false
		}
	}
	return fids, nil
}

package alfafrens

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

// ChannelFids fetches the fids given a channel address using pagination.
func ChannelFids(channelAddress types.HexBytes) ([]uint64, error) {
	channel := channelAddress.String()
	if !strings.HasPrefix(channel, "0x") {
		channel = "0x" + channel
	}

	var fids []uint64
	skip := 0

	for {
		url := fmt.Sprintf(channelSubscribersURL+"&first=200&skip=%d", channel, skip)
		log.Debugw("fetching alfafrens channel subscribers", "url", url)
		// Perform the HTTP GET request
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// Parse the JSON response
		var channelResponse ChannelResponse
		err = json.Unmarshal(body, &channelResponse)
		if err != nil {
			return nil, err
		}

		// Extract the fids and convert to []uint64
		for _, member := range channelResponse.Members {
			// Only add the fid if the member is subscribed
			if member.IsSubscribed {
				fids = append(fids, uint64(member.FID))
			}
		}

		// Check if there are more pages
		if !channelResponse.HasMore {
			break
		}
		skip += len(channelResponse.Members)
	}
	return fids, nil
}

// ChannelByFid fetches the user details given a fid.
func ChannelByFid(fid uint64) (types.HexBytes, error) {
	url := fmt.Sprintf(userByFidInfo, fid)

	// Perform the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return user.ChannelAddress, nil
}

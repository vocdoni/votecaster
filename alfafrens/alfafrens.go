package alfafrens

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.vocdoni.io/dvote/types"
)

// ChannelFids fetches the fids given a channel address.
func ChannelFids(channelAddress types.HexBytes) ([]uint64, error) {
	channel := channelAddress.String()
	if !strings.HasPrefix(channel, "0x") {
		channel = "0x" + channel
	}
	url := fmt.Sprintf(channelSubscribersURL, channel)

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
	var fids []uint64
	for _, member := range channelResponse.Members {
		// Only add the fid if the member is subscribed and has a fid
		if member.IsSubscribed && member.FID != 0 {
			fids = append(fids, uint64(member.FID))
		}
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

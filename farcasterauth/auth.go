package farcasterauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.vocdoni.io/dvote/log"
)

const (
	baseURL = "https://relay.farcaster.xyz/v1"
)

// ErrAuthenticationPending is returned when the authentication process is still pending.
var ErrAuthenticationPending = fmt.Errorf("authentication pending")

// Client holds the HTTP client used for the requests and the ChannelToken needed for authentication.
// Implements FIP-11: https://github.com/farcasterxyz/protocol/discussions/110
type Client struct {
	client       *http.Client
	channelToken string
	channelNonce string
}

// New creates a farcasterAuth client.
func New() *Client {
	return &Client{
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// ChannelResponse holds the response data from creating a new channel.
type ChannelResponse struct {
	ChannelToken string `json:"channelToken"`
	URL          string `json:"url"`
	ConnectUri   string `json:"connectUri"`
	Nonce        string `json:"nonce"`
}

// StatusResponse holds the response data for checking the channel's status.
type StatusResponse struct {
	State         string   `json:"state,omitempty"`
	Nonce         string   `json:"nonce,omitempty"`
	Message       string   `json:"message,omitempty"`
	Signature     string   `json:"signature,omitempty"`
	Fid           uint64   `json:"fid,omitempty"`
	Username      string   `json:"username,omitempty"`
	DisplayName   string   `json:"displayName,omitempty"`
	Bio           string   `json:"bio,omitempty"`
	PfpUrl        string   `json:"pfpUrl,omitempty"`
	Custody       string   `json:"custody,omitempty"`
	Verifications []string `json:"verifications,omitempty"`
}

// CreateChannel initiates the authentication process by creating a new channel.
// The siweUri is the URI of the SIWE server and the domain is the domain of the application that is requesting authentication.
// The returned response includes the URL that must be relayed to the user for authentication.
func (c *Client) CreateChannel(siweUri string) (*ChannelResponse, error) {
	domain := strings.Split(strings.Split(siweUri, "/")[2], ":")[0]
	payload := map[string]string{
		"siweUri": siweUri,
		"domain":  domain,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Post(fmt.Sprintf("%s/channel", baseURL), "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		r, _ := io.ReadAll(resp.Body)
		log.Warnw("auth request failed", "resp", string(r), "code", resp.StatusCode)
		return nil, fmt.Errorf("failed to create channel")
	}

	var channelResp ChannelResponse
	if err := json.NewDecoder(resp.Body).Decode(&channelResp); err != nil {
		return nil, err
	}

	if channelResp.ChannelToken == "" || channelResp.Nonce == "" {
		return nil, fmt.Errorf("channel token and nonce not found in response")
	}
	c.channelToken = channelResp.ChannelToken
	c.channelNonce = channelResp.Nonce
	return &channelResp, nil
}

// CheckStatus checks the status of the authentication process. A channel must be created first.
// If the error message is ErrAuthenticationPending, the client should retry after a few seconds.
// If the status is "completed", the authentication process is successful and the returned data can be used.
// Expect only valid responses when the status is "completed".
func (c *Client) CheckStatus() (*StatusResponse, error) {
	if c.channelToken == "" || c.channelNonce == "" {
		return nil, fmt.Errorf("channel must be created first")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/channel/status", baseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.channelToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to check status")
	}

	var statusResp StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, err
	}

	if statusResp.Nonce != c.channelNonce {
		return nil, fmt.Errorf("nonce mismatch")
	}

	if statusResp.State == "pending" {
		return &statusResp, ErrAuthenticationPending
	}

	if statusResp.State != "completed" {
		return nil, fmt.Errorf("authentication failed: %s", statusResp.Message)
	}
	// Clear the response data
	statusResp.Message = ""
	statusResp.Nonce = ""
	statusResp.Signature = ""
	statusResp.State = ""
	return &statusResp, nil
}

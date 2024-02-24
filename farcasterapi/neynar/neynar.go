package neynar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vocdoni/farcaster-poc/farcasterapi"
)

const (
	// endpoints
	neynarGetUsernameEndpoint = "v1/farcaster/user?fid=%d"
	neynarGetCastsEndpoint    = "v1/farcaster/mentions-and-replies?fid=%d&limit=150&cursor=%s"
	neynarReplyEndpoint       = "v2/farcaster/cast"
	neynarUserByEthAddresses  = "v2/farcaster/user/bulk-by-address?addresses=%s"
	// timeouts
	getBotUsernameTimeout   = 10 * time.Second
	getCastByMentionTimeout = 60 * time.Second
	postCastTimeout         = 10 * time.Second
	// other
	neynarMentionType = "cast-mention"
	timeLayout        = "2006-01-02T15:04:05.000Z"
)

type NeynarAPI struct {
	fid        uint64
	username   string
	signerUUID string
	apiKey     string
	endpoint   string
}

// Init initializes the API with the given arguments. apiKeys must have
// a single element, which is the api key for the Neynar API.
func (n *NeynarAPI) Init(apiEndpoint string, apiKeys []string) error {
	if len(apiKeys) == 0 {
		return fmt.Errorf("no api keys provided")
	}
	n.endpoint = apiEndpoint
	n.apiKey = apiKeys[0]
	return nil
}

// SetFarcasterUser method sets the farcaster user with the given fid and signer.
// The signer is the UUID of the user that signs the messages.
func (n *NeynarAPI) SetFarcasterUser(fid uint64, signer string) error {
	n.fid = fid
	n.signerUUID = signer
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), getBotUsernameTimeout)
	defer cancel()
	userdata, err := n.UserDataByFID(ctx, n.fid)
	if err != nil {
		return fmt.Errorf("error getting bot username: %w", err)
	}
	n.username = userdata.Username
	return nil
}

func (n *NeynarAPI) Stop() error {
	return nil
}

func (n *NeynarAPI) LastMentions(ctx context.Context, timestamp uint64) ([]*farcasterapi.APIMessage, uint64, error) {
	if n.fid == 0 {
		return nil, 0, fmt.Errorf("farcaster user not set")
	}
	baseURL := fmt.Sprintf("%s/%s", n.endpoint, neynarGetCastsEndpoint)

	internalCtx, cancel := context.WithTimeout(ctx, getCastByMentionTimeout)
	defer cancel()

	messages := []*farcasterapi.APIMessage{}
	lastTimestamp := timestamp
	cursor := ""
	for {
		// create request with the given cursor and set the api key header
		url := fmt.Sprintf(baseURL, n.fid, cursor)
		req, err := http.NewRequestWithContext(internalCtx, http.MethodGet, url, nil)
		if err != nil {
			return nil, 0, fmt.Errorf("error creating request: %w", err)
		}
		req.Header.Set("api_key", n.apiKey)
		// send request and check response status
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, 0, fmt.Errorf("error downloading json: %w", err)
		}
		if res.StatusCode != http.StatusOK {
			return nil, 0, fmt.Errorf("error downloading json: %s", res.Status)
		}
		// read response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, 0, fmt.Errorf("error reading response body: %w", err)
		}
		defer res.Body.Close()
		// decode mentions
		notificationsResponse := &NotificationsResponse{}
		if err := json.Unmarshal(body, notificationsResponse); err != nil {
			return nil, 0, fmt.Errorf("error unmarshalling response body: %w", err)
		}
		// parse mentions
		for _, notification := range notificationsResponse.Result.Notifications {
			// skip non-mentions
			if notification.Type != neynarMentionType {
				continue
			}
			// parse timestamp
			parsedTimestamp, err := time.Parse(timeLayout, notification.Timestamp)
			if err != nil {
				return nil, 0, fmt.Errorf("error parsing timestamp: %w", err)
			}
			// skip old mentions
			notificationTimestamp := uint64(parsedTimestamp.Unix())
			if notificationTimestamp <= timestamp {
				continue
			}
			// parse the text to remove the bot username and add mention to the
			// list
			mention := fmt.Sprintf("@%s", n.username)
			text := strings.TrimSpace(strings.TrimPrefix(notification.Text, mention))
			messages = append(messages, &farcasterapi.APIMessage{
				IsMention: true,
				Author:    notification.Author.FID,
				Content:   text,
				Hash:      notification.Hash,
			})
			// update last timestamp
			if notificationTimestamp > lastTimestamp {
				lastTimestamp = notificationTimestamp
			}
		}
		// stop if there are no new mentions
		if notificationsResponse.Result.NextCursor.Cursor == "" {
			break
		}
		cursor = notificationsResponse.Result.NextCursor.Cursor
	}
	return messages, lastTimestamp, nil
}

func (n *NeynarAPI) Reply(ctx context.Context, fid uint64, parentHash, content string) error {
	if n.fid == 0 {
		return fmt.Errorf("farcaster user not set")
	}
	// create request body
	castReq := &CastPostRequest{
		Signer: n.signerUUID,
		Text:   content,
		Parent: parentHash,
	}
	body, err := json.Marshal(castReq)
	if err != nil {
		return fmt.Errorf("error marshalling request body: %w", err)
	}
	url := fmt.Sprintf("%s/%s", n.endpoint, neynarReplyEndpoint)
	internalCtx, cancel := context.WithTimeout(ctx, postCastTimeout)
	defer cancel()
	// create request with the bot fid and set the api key header
	req, err := http.NewRequestWithContext(internalCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("api_key", n.apiKey)
	req.Header.Set("Content-Type", "application/json")
	// send request and check response status
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending cast: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error sending cast: %s", res.Status)
	}
	return nil
}

// UserData method returns the username, the custody address and the
// verification addresses of the user with the given fid. If something goes
// wrong, it returns an error.
func (n *NeynarAPI) UserDataByFID(ctx context.Context, fid uint64) (*farcasterapi.Userdata, error) {
	internalCtx, cancel := context.WithTimeout(ctx, getBotUsernameTimeout)
	defer cancel()

	// create request with the bot fid
	baseURL := fmt.Sprintf("%s/%s", n.endpoint, neynarGetUsernameEndpoint)
	url := fmt.Sprintf(baseURL, fid)
	req, err := http.NewRequestWithContext(internalCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("api_key", n.apiKey)
	// send request and check response status
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error downloading json: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading json: %s", res.Status)
	}
	// read response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	defer res.Body.Close()
	// decode username
	usernameResponse := &UserdataResponse{}
	if err := json.Unmarshal(body, usernameResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}
	return &farcasterapi.Userdata{
		FID:                    fid,
		Username:               usernameResponse.Result.User.Username,
		CustodyAddress:         usernameResponse.Result.User.CustodyAddress,
		VerificationsAddresses: usernameResponse.Result.User.VerificationsAddresses,
	}, nil
}

func (n *NeynarAPI) UserDataByVerificationAddress(ctx context.Context, address string) (*farcasterapi.Userdata, error) {
	internalCtx, cancel := context.WithTimeout(ctx, getBotUsernameTimeout)
	defer cancel()

	baseURL := fmt.Sprintf("%s/%s", n.endpoint, neynarUserByEthAddresses)
	url := fmt.Sprintf(baseURL, address)
	req, err := http.NewRequestWithContext(internalCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("api_key", n.apiKey)
	// send request and check response status
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error downloading json: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading json: %s", res.Status)
	}
	// read response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	defer res.Body.Close()
	// decode username
	results := map[string][]*UserdataV2{}
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}
	dataItems, ok := results[address]
	if !ok || len(dataItems) == 0 {
		return nil, fmt.Errorf("no data found for the given address")
	}
	var data *UserdataV2
	for _, item := range dataItems {
		if item.Username != "" {
			data = item
			break
		}
	}
	if data == nil {
		return nil, fmt.Errorf("no valid data found for the given address")
	}
	return &farcasterapi.Userdata{
		FID:                    data.FID,
		Username:               data.Username,
		CustodyAddress:         data.CustodyAddress,
		VerificationsAddresses: data.VerificationsAddresses,
	}, nil
}

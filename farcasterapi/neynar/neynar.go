package neynar

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/farcasterapi/web3"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/util"
)

const (
	NeynarHubEndpoint      = "https://hub-api.neynar.com/v1"
	NeynarAPIEndpoint      = "https://api.neynar.com"
	WarpcastClientEndpoint = "https://client.warpcast.com/v2"

	// endpoints
	neynarGetUsernameEndpoint = NeynarAPIEndpoint + "/v1/farcaster/user?fid=%d"
	neynarGetCastsEndpoint    = NeynarAPIEndpoint + "/v1/farcaster/mentions-and-replies?fid=%d&limit=150&cursor=%s"
	neynarReplyEndpoint       = NeynarAPIEndpoint + "/v2/farcaster/cast"
	neynarUserByEthAddresses  = NeynarAPIEndpoint + "/v2/farcaster/user/bulk-by-address?addresses=%s"
	neynarUserFollowers       = NeynarAPIEndpoint + "/v1/farcaster/followers?fid=%d&limit=150&cursor=%s"
	neynarChannelDataByID     = NeynarAPIEndpoint + "/v2/farcaster/channel?id=%s"
	neynarUsersByChannelID    = NeynarAPIEndpoint + "/v2/farcaster/channel/followers?id=%s&limit=1000&cursor=%s"
	neynarVerificationsByFID  = NeynarHubEndpoint + "/verificationsByFid?fid=%d"
	warpcastChannelInfo       = WarpcastClientEndpoint + "/channel?key=%s"

	MaxAddressesPerRequest = 300

	// timeouts
	getBotUsernameTimeout   = 10 * time.Second
	getCastByMentionTimeout = 60 * time.Second
	postCastTimeout         = 10 * time.Second
	defaultRequestTimeout   = 10 * time.Second

	// Requests backoff parameters
	maxConcurrentRequests = 2
	maxRetries            = 12              // Maximum number of retries
	baseDelay             = 1 * time.Second // Initial delay, increases exponentially

	// other
	neynarMentionType     = "cast-mention"
	neynarCastCreatedType = "cast.created"
	neynarCastType        = "cast"
	timeLayout            = "2006-01-02T15:04:05.000Z"
)

type NeynarAPI struct {
	fid          uint64
	username     string
	signerUUID   string
	apiKey       string
	reqSemaphore chan struct{} // Semaphore to limit concurrent requests
	newCasts     map[uint64]*farcasterapi.APIMessage
	newCastsMtx  sync.Mutex
	web3provider *web3.FarcasterProvider
}

func NewNeynarAPI(apiKey string, web3endpoints []string) (*NeynarAPI, error) {
	web3provider := web3.NewFarcasterProvider()
	for _, web3endpoint := range web3endpoints {
		if err := web3provider.AddEndpoint(web3endpoint); err != nil {
			return nil, err
		}
		// Run a quick test to check if the web3 endpoint is working
		signers, err := web3provider.GetAppKeysByFid(big.NewInt(3))
		if err != nil {
			return nil, err
		}
		if len(signers) == 0 {
			log.Warnw("web3 endpoint not working", "endpoint", web3endpoint)
			web3provider.DelEndpoint(web3endpoint)
		}
	}
	return &NeynarAPI{
		apiKey:       apiKey,
		reqSemaphore: make(chan struct{}, maxConcurrentRequests),
		newCasts:     make(map[uint64]*farcasterapi.APIMessage),
		web3provider: web3provider,
	}, nil
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
	// get new mentions from the queue and calculate the last timestamp
	messages := []*farcasterapi.APIMessage{}
	lastTimestamp := timestamp
	n.newCastsMtx.Lock()
	for ts, msg := range n.newCasts {
		if ts > timestamp {
			messages = append(messages, msg)
			lastTimestamp = ts
		}
	}
	// clear the new mentions queue
	n.newCasts = make(map[uint64]*farcasterapi.APIMessage)
	n.newCastsMtx.Unlock()
	return messages, lastTimestamp, nil
}

func (n *NeynarAPI) Reply(ctx context.Context, fid uint64, parentHash, content string, _ ...string) error {
	if n.fid == 0 {
		return fmt.Errorf("farcaster user not set")
	}
	// create request body
	castReq := &castPostRequest{
		Signer: n.signerUUID,
		Text:   content,
		Parent: parentHash,
	}
	body, err := json.Marshal(castReq)
	if err != nil {
		return fmt.Errorf("error marshalling request body: %w", err)
	}
	// create request with the bot fid and set the api key header
	_, err = n.request(neynarReplyEndpoint, http.MethodPost, body, 0)
	return err
}

// UserData method returns the username, the custody address and the
// verification addresses of the user with the given fid.
func (n *NeynarAPI) UserDataByFID(ctx context.Context, fid uint64) (*farcasterapi.Userdata, error) {
	// create request with the bot fid
	url := fmt.Sprintf(neynarGetUsernameEndpoint, fid)
	body, err := n.request(url, http.MethodGet, nil, defaultRequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	// decode username
	usernameResponse := &userdataV1Response{}
	if err := json.Unmarshal(body, usernameResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	// get signers
	signers, err := n.SignersFromFID(fid)
	if err != nil {
		return nil, fmt.Errorf("error getting signers: %w", err)
	}

	return &farcasterapi.Userdata{
		FID:                    fid,
		Username:               usernameResponse.Result.User.Username,
		CustodyAddress:         usernameResponse.Result.User.CustodyAddress,
		VerificationsAddresses: usernameResponse.Result.User.VerificationsAddresses,
		Signers:                signers,
	}, nil
}

// SignersFromFid method returns the signers (appkeys) of the user with the given fid.
func (n *NeynarAPI) SignersFromFID(fid uint64) ([]string, error) {
	signersBytes, err := n.web3provider.GetAppKeysByFid(big.NewInt(int64(fid)))
	if err != nil {
		return nil, fmt.Errorf("error getting signers: %w", err)
	}
	signers := []string{}
	for _, signer := range signersBytes {
		signers = append(signers, hex.EncodeToString(signer))
	}
	return signers, nil
}

// UserDataByVerificationAddress method returns the userdata of the user with the given Ethereum address.
func (n *NeynarAPI) UserDataByVerificationAddress(ctx context.Context, addresses []string) ([]*farcasterapi.Userdata, error) {
	if len(addresses) > MaxAddressesPerRequest {
		return nil, fmt.Errorf("address slice exceeds the maximum limit of 350 addresses")
	}
	// Concatenate addresses separated by commas
	addressesStr := strings.Join(addresses, ",")
	// Construct the URL with multiple addresses
	url := fmt.Sprintf(neynarUserByEthAddresses, addressesStr)
	// Make the request
	body, err := n.request(url, http.MethodGet, nil, defaultRequestTimeout)
	if err != nil {
		return nil, err
	}
	// Decode the response
	var results map[string][]userdataV2
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}
	// Process results into []*farcasterapi.Userdata
	userDataSlice := make([]*farcasterapi.Userdata, 0)
	for address, dataItems := range results {
		for _, item := range dataItems {
			if item.Username != "" {
				if len(item.VerifiedAddresses.EthAddresses) == 0 {
					log.Warnw("no verified addresses found", "user", item.Username)
					continue
				}
				// Normalize addresses to the Ethereum hex standard format
				var normalizedAddresses []string
				for _, addr := range item.VerifiedAddresses.EthAddresses {
					normalizedAddresses = append(normalizedAddresses, common.HexToAddress(addr).Hex())
				}
				// Get signers
				signers, err := n.SignersFromFID(item.Fid)
				if err != nil {
					return nil, fmt.Errorf("error getting signers for address %s: %w", address, err)
				}
				if len(signers) == 0 {
					log.Warnw("no signers found", "user", item.Username, "address", address)
					continue
				}
				userData := &farcasterapi.Userdata{
					FID:                    item.Fid,
					Username:               item.Username,
					CustodyAddress:         common.HexToAddress(item.CustodyAddress).Hex(),
					VerificationsAddresses: normalizedAddresses,
					Signers:                signers,
				}
				userDataSlice = append(userDataSlice, userData)
				break // we only need the first valid user data per address
			}
		}
	}
	if len(userDataSlice) == 0 {
		return nil, farcasterapi.ErrNoDataFound
	}
	return userDataSlice, nil
}

// UserFollowers method returns the FIDs of the followers of the user with the
// given id. If something goes wrong, it returns an error.
func (n *NeynarAPI) UserFollowers(ctx context.Context, fid uint64) ([]uint64, error) {
	cursor := ""
	userFIDs := []uint64{}
	for {
		// create request with the channel id provided
		url := fmt.Sprintf(neynarUserFollowers, fid, cursor)
		body, err := n.request(url, http.MethodGet, nil, defaultRequestTimeout)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}
		usersResponse := &UsersdataV1Response{}
		if err := json.Unmarshal(body, &usersResponse); err != nil {
			return nil, fmt.Errorf("error unmarshalling response body: %w", err)
		}
		if usersResponse.Result.Users == nil {
			return nil, farcasterapi.ErrNoDataFound
		}
		for _, user := range usersResponse.Result.Users {
			userFIDs = append(userFIDs, user.Fid)
		}
		if usersResponse.Result.NextCursor == nil || usersResponse.Result.NextCursor.Cursor == "" {
			break
		}
		cursor = usersResponse.Result.NextCursor.Cursor
	}
	return userFIDs, nil
}

// ChannelFIDs method returns the FIDs of the users that follow the channel with
// the given id. If something goes wrong, it returns an error. It return an
// specific error if the channel does not exist to be handled by the caller.
func (n *NeynarAPI) ChannelFIDs(ctx context.Context, channelID string, progress chan int) ([]uint64, error) {
	// check if the channel exists
	exists, err := n.ChannelExists(channelID)
	if err != nil {
		return nil, fmt.Errorf("error checking channel existence: %w", err)
	}
	if !exists {
		return nil, farcasterapi.ErrChannelNotFound
	}
	// get the followers of the channel to update the progress
	totalFollowers := 0
	channelURL := fmt.Sprintf(warpcastChannelInfo, channelID)
	body, err := n.request(channelURL, http.MethodGet, nil, defaultRequestTimeout)
	if err == nil {
		channelResponse := &warpcastChannelResponse{}
		if err := json.Unmarshal(body, &channelResponse); err == nil {
			totalFollowers = channelResponse.Result.Channel.Followers
		}
	}
	cursor := ""
	userFIDs := []uint64{}
	for {
		// create request with the channel id provided
		url := fmt.Sprintf(neynarUsersByChannelID, channelID, cursor)
		body, err := n.request(url, http.MethodGet, nil, defaultRequestTimeout)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}
		usersResult := &userdataV2Result{}
		if err := json.Unmarshal(body, &usersResult); err != nil {
			return nil, fmt.Errorf("error unmarshalling response body: %w", err)
		}
		for _, user := range usersResult.Users {
			userFIDs = append(userFIDs, user.Fid)
		}
		// update the progress calculating the percentage of the followers
		// already processed
		if progress != nil && totalFollowers > 0 {
			processedFollowers := len(userFIDs)
			progress <- int(float64(processedFollowers) / float64(totalFollowers) * 100)
		}
		if usersResult.NextCursor == nil || usersResult.NextCursor.Cursor == "" {
			break
		}
		cursor = usersResult.NextCursor.Cursor
	}
	if progress != nil {
		progress <- 100
	}
	return userFIDs, nil
}

// ChannelExists method returns a boolean indicating if the channel with the
// given id exists. If something goes wrong checking the channel existence,
// it returns an error.
func (n *NeynarAPI) ChannelExists(channelID string) (bool, error) {
	// create request with the channel id provided
	url := fmt.Sprintf(neynarChannelDataByID, channelID)
	if _, err := n.request(url, http.MethodGet, nil, defaultRequestTimeout); err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("error creating request: %w", err)
	}
	return true, nil
}

func (n *NeynarAPI) WebhookHandler(body []byte) error {
	// decode the request body
	castWebhookReq := &castWebhookRequest{}
	if err := json.Unmarshal(body, castWebhookReq); err != nil {
		return fmt.Errorf("error unmarshalling request body: %s", err.Error())
	}
	// check if the req type is a not created cast or data type is not cast and
	// skip if so
	if castWebhookReq.Type != neynarCastCreatedType || castWebhookReq.Data.Object != neynarCastType {
		return nil
	}
	// parse timestamp
	parsedTimestamp, err := time.Parse(timeLayout, castWebhookReq.Data.Timestamp)
	if err != nil {
		return fmt.Errorf("error parsing timestamp: %s", err.Error())
	}
	notificationTimestamp := uint64(parsedTimestamp.Unix())
	// check if the cast is a mention and skip if not
	mention := fmt.Sprintf("@%s", n.username)
	if !strings.HasPrefix(castWebhookReq.Data.Text, mention) {
		return nil
	}
	// parse the text to remove the bot username and add mention to the
	// list
	text := strings.TrimSpace(strings.TrimPrefix(castWebhookReq.Data.Text, mention))
	message := &farcasterapi.APIMessage{
		IsMention: true,
		Author:    castWebhookReq.Data.Author.Fid,
		Content:   text,
		Hash:      castWebhookReq.Data.Hash,
	}
	// add the new mention to the list
	n.newCastsMtx.Lock()
	n.newCasts[notificationTimestamp] = message
	n.newCastsMtx.Unlock()
	return nil
}

func (n *NeynarAPI) request(url, method string, body []byte, timeout time.Duration) ([]byte, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}
		req.Header.Set("api_key", n.apiKey)
		req.Header.Set("Content-Type", "application/json")

		// We need to avoid too much concurrent requests and penalization from the API
		n.reqSemaphore <- struct{}{}
		res, err := http.DefaultClient.Do(req)
		<-n.reqSemaphore
		if err != nil {
			return nil, fmt.Errorf("error downloading json: %w", err)
		}
		defer res.Body.Close()
		if res.StatusCode == http.StatusTooManyRequests {
			time.Sleep(time.Duration(attempt+1)*baseDelay + time.Duration(util.RandomInt(0, 2000))*time.Millisecond)
		} else if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error downloading json: %s", res.Status)
		} else {
			respBody, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response body: %w", err)
			}
			return respBody, nil // Success
		}
		log.Debugw("retrying request", "attempt", attempt+1, "url", url, "method", method)
	}

	return nil, fmt.Errorf("error downloading json: exceeded retry limit")
}

// VerifyRequest method verifies the request signature and returns a boolean
// indicating if the signature is valid and an error if something goes wrong.
func VerifyRequest(secret, signature string, body []byte) (bool, error) {
	// Create HMAC with SHA512 and update it with the body
	hmac := hmac.New(sha512.New, []byte(secret))
	hmac.Write(body)
	// Calculate the HMAC signature and encode it in hexadecimal
	bodySig := hex.EncodeToString(hmac.Sum(nil))
	// verify the signature
	return signature == bodySig, nil
}

package neynar

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	c3web3 "github.com/vocdoni/census3/helpers/web3"
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
	neynarGetUsernameEndpoint = NeynarAPIEndpoint + "/v2/farcaster/user/bulk?fids=%d"
	neynarGetCastsEndpoint    = NeynarAPIEndpoint + "/v1/farcaster/mentions-and-replies?fid=%d&limit=150&cursor=%s"
	neynarGetCastEndpoint     = NeynarAPIEndpoint + "/v2/farcaster/cast?identifier=%s&type=hash"
	neynarReplyEndpoint       = NeynarAPIEndpoint + "/v2/farcaster/cast"
	neynarUserByEthAddresses  = NeynarAPIEndpoint + "/v2/farcaster/user/bulk-by-address?addresses=%s"
	neynarUserFollowers       = NeynarAPIEndpoint + "/v1/farcaster/followers?fid=%d&limit=150&cursor=%s"
	neynarChannelDataByID     = NeynarAPIEndpoint + "/v2/farcaster/channel?id=%s"
	neynarSuggestChannels     = NeynarAPIEndpoint + "/v2/farcaster/channel/search?q=%s"
	neynarUsersByChannelID    = NeynarAPIEndpoint + "/v2/farcaster/channel/followers?id=%s&limit=1000&cursor=%s"
	neynarVerificationsByFID  = NeynarHubEndpoint + "/verificationsByFid?fid=%d"
	warpcastChannelInfo       = WarpcastClientEndpoint + "/channel?key=%s"

	MaxAddressesPerRequest = 200

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

func NewNeynarAPI(apiKey string, web3Pool *c3web3.Web3Pool) (*NeynarAPI, error) {
	log.Infow("starting neynar api with web3 support", "web3endpoints", len(web3Pool.String()))
	web3provider, err := web3.NewFarcasterProvider(web3Pool)
	if err != nil {
		return nil, fmt.Errorf("error creating web3 provider: %w", err)
	}
	// Run a quick test to check if the web3 endpoint is working
	signers, err := web3provider.GetAppKeysByFid(big.NewInt(3))
	if err != nil {
		return nil, err
	}
	if len(signers) == 0 {
		return nil, fmt.Errorf("no valid web3 endpoints found")
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

// FID method returns the fid of the farcaster user set in the API.
func (n *NeynarAPI) FID() uint64 {
	return n.fid
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

func (n *NeynarAPI) GetCast(ctx context.Context, _ uint64, hash string) (*farcasterapi.APIMessage, error) {
	msgResponse := &castResponseV2{}
	url := fmt.Sprintf(neynarGetCastEndpoint, hash)
	body, err := n.neynarReq(ctx, url, http.MethodGet, nil, defaultRequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get the cast: %w", err)
	}
	if err := json.Unmarshal(body, msgResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}
	if msgResponse.Data == nil {
		return nil, farcasterapi.ErrNoDataFound
	}
	message, err := n.parseCastData(msgResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("error parsing cast data: %w", err)
	}
	return message, nil
}

func (n *NeynarAPI) Publish(ctx context.Context, content string, _ []uint64, embeds ...string) error {
	if n.fid == 0 {
		return fmt.Errorf("farcaster user not set")
	}
	// check if the content is too long
	if len([]byte(content)) > farcasterapi.MaxCastBytes {
		return fmt.Errorf("content is too long")
	}
	castEmbeds := []*castEmbed{}
	if len(embeds) > 0 {
		for _, embed := range embeds {
			castEmbeds = append(castEmbeds, &castEmbed{embed})
		}
	}
	// create request body
	castReq := &castPostRequest{
		Signer: n.signerUUID,
		Text:   content,
		Embeds: castEmbeds,
	}
	body, err := json.Marshal(castReq)
	if err != nil {
		return fmt.Errorf("error marshalling request body: %w", err)
	}
	// create request with the bot fid and set the api key header
	_, err = n.neynarReq(ctx, neynarReplyEndpoint, http.MethodPost, body, defaultRequestTimeout)
	return err
}

func (n *NeynarAPI) Reply(ctx context.Context, targetMsg *farcasterapi.APIMessage,
	content string, _ []uint64, embeds ...string,
) error {
	if n.fid == 0 {
		return fmt.Errorf("farcaster user not set")
	}
	// check if the content is too long
	if len([]byte(content)) > farcasterapi.MaxCastBytes {
		return fmt.Errorf("content is too long")
	}
	castEmbeds := []*castEmbed{}
	if len(embeds) > 0 {
		for _, embed := range embeds {
			castEmbeds = append(castEmbeds, &castEmbed{embed})
		}
	}
	// create request body
	castReq := &castPostRequest{
		Signer: n.signerUUID,
		Text:   content,
		Parent: targetMsg.Hash,
		Embeds: castEmbeds,
	}
	body, err := json.Marshal(castReq)
	if err != nil {
		return fmt.Errorf("error marshalling request body: %w", err)
	}
	// create request with the bot fid and set the api key header
	_, err = n.neynarReq(ctx, neynarReplyEndpoint, http.MethodPost, body, 0)
	return err
}

// UserData method returns the username, the custody address and the
// verification addresses of the user with the given fid.
func (n *NeynarAPI) UserDataByFID(ctx context.Context, fid uint64) (*farcasterapi.Userdata, error) {
	// create request with the bot fid
	url := fmt.Sprintf(neynarGetUsernameEndpoint, fid)
	body, err := n.neynarReq(ctx, url, http.MethodGet, nil, defaultRequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	// decode username
	usernameResponse := &userdataV2Result{}
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
		Username:               usernameResponse.Users[0].Username,
		CustodyAddress:         usernameResponse.Users[0].CustodyAddress,
		VerificationsAddresses: usernameResponse.Users[0].VerifiedAddresses.EthAddresses,
		Avatar:                 usernameResponse.Users[0].PfpUrl,
		Bio:                    usernameResponse.Users[0].Profile.Bio.Text,
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
	body, err := n.neynarReq(ctx, url, http.MethodGet, nil, defaultRequestTimeout)
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
		body, err := n.neynarReq(ctx, url, http.MethodGet, nil, defaultRequestTimeout)
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

// Channel method returns the details of a channel given its channelID. If
// something goes wrong it returns an error.
func (n *NeynarAPI) Channel(ctx context.Context, channelID string) (*farcasterapi.Channel, error) {
	if channelID == "" {
		return nil, nil
	}
	// create request with the channel id provided
	url := fmt.Sprintf(neynarChannelDataByID, channelID)
	res, err := n.neynarReq(ctx, url, http.MethodGet, nil, defaultRequestTimeout)
	if err != nil {
		log.Warnw("error getting channel", "channel", channelID, "error", err)
		if strings.Contains(err.Error(), "404") {
			return nil, farcasterapi.ErrChannelNotFound
		}
		return nil, fmt.Errorf("error getting channel from API: %w", err)
	}
	channelData := &warpcastChannelResult{}
	if err := json.Unmarshal(res, channelData); err != nil {
		return nil, fmt.Errorf("error decoding channel information: %w", err)
	}
	if channelData.Channel == nil {
		return nil, fmt.Errorf("no data for this channel")
	}
	return &farcasterapi.Channel{
		ID:          channelData.Channel.ID,
		Name:        channelData.Channel.Name,
		Description: channelData.Channel.Description,
		Followers:   channelData.Channel.Followers,
		Image:       channelData.Channel.ImageURL,
		URL:         channelData.Channel.URL,
	}, nil
}

// ChannelFIDs method returns the FIDs of the users that follow the channel with
// the given id. If something goes wrong, it returns an error. It return an
// specific error if the channel does not exist to be handled by the caller.
func (n *NeynarAPI) ChannelFIDs(ctx context.Context, channelID string, progress chan int) ([]uint64, error) {
	// check if the channel exists
	channel, err := n.Channel(ctx, channelID)
	if err != nil {
		if errors.Is(err, farcasterapi.ErrChannelNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("error checking channel existence: %w", err)
	}
	if channel.Followers == 0 {
		return nil, fmt.Errorf("channel %s has no followers", channelID)
	}
	cursor := ""
	userFIDs := []uint64{}
	failedAttempts := 5
	for {
		// create request with the channel id provided
		url := fmt.Sprintf(neynarUsersByChannelID, channelID, cursor)
		body, err := n.neynarReq(ctx, url, http.MethodGet, nil, defaultRequestTimeout)
		if err != nil {
			failedAttempts--
			if failedAttempts == 0 {
				return nil, fmt.Errorf("error creating request: %w", err)
			}
			log.Warnw("error getting channel followers, retrying", "channel", channelID, "error", err)
			continue
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
		if progress != nil && channel.Followers > 0 {
			processedFollowers := len(userFIDs)
			progress <- int(float64(processedFollowers) / float64(channel.Followers) * 100)
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
func (n *NeynarAPI) ChannelExists(ctx context.Context, channelID string) (bool, error) {
	_, err := n.Channel(ctx, channelID)
	if err != nil {
		if errors.Is(err, farcasterapi.ErrChannelNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("error checking channel existence: %w", err)
	}
	return true, nil
}

// FindChannel method returns the list of channels that match the query provided.
// If something goes wrong, it returns an error.
func (n *NeynarAPI) FindChannel(ctx context.Context, query string, adminFid uint64) ([]*farcasterapi.Channel, error) {
	channels := []*farcasterapi.Channel{}
	// create request with the channel id provided
	url := fmt.Sprintf(neynarSuggestChannels, query)
	body, err := n.neynarReq(ctx, url, http.MethodGet, nil, defaultRequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	channelsResponse := &warpcastChannelsResult{}
	if err := json.Unmarshal(body, &channelsResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}
	if len(channelsResponse.Channels) == 0 {
		return nil, farcasterapi.ErrNoDataFound
	}
	for _, ch := range channelsResponse.Channels {
		// only include the channel if no adminFid is provided or the user is
		// the admin of the channel
		if adminFid == 0 || ch.Lead.Fid == adminFid {
			channels = append(channels, &farcasterapi.Channel{
				ID:          ch.ID,
				Name:        ch.Name,
				Description: ch.Description,
				Followers:   ch.Followers,
				Image:       ch.ImageURL,
				URL:         ch.URL,
			})
		}
	}
	return channels, nil
}

// DirectMessage method sends a direct message to the user with the given fid.
// If something goes wrong, it returns an error.
func (n *NeynarAPI) DirectMessage(ctx context.Context, content string, to uint64) error {
	return nil
}

func (n *NeynarAPI) WebhookHandler(body []byte) error {
	// decode the request body
	castWebhookReq := &castsWebhookRequest{}
	if err := json.Unmarshal(body, castWebhookReq); err != nil {
		return fmt.Errorf("error unmarshalling request body: %s", err.Error())
	}
	// check if the req type is a not created cast or data type is not cast and
	// skip if so
	if castWebhookReq.Type != neynarCastCreatedType {
		return nil
	}
	message, err := n.parseCastData(castWebhookReq.Data)
	if err != nil {
		return fmt.Errorf("error parsing cast data: %w", err)
	}
	// parse timestamp
	parsedTimestamp, err := time.Parse(timeLayout, castWebhookReq.Data.Timestamp)
	if err != nil {
		return fmt.Errorf("error parsing timestamp: %s", err.Error())
	}
	notificationTimestamp := uint64(parsedTimestamp.Unix())
	// add the new mention to the list
	n.newCastsMtx.Lock()
	n.newCasts[notificationTimestamp] = message
	n.newCastsMtx.Unlock()
	return nil
}

func (n *NeynarAPI) parseCastData(data *castWebhookData) (*farcasterapi.APIMessage, error) {
	// check if the req type is a not created cast or data type is not cast and
	// skip if so
	if data.Object != neynarCastType {
		return nil, fmt.Errorf("invalid object type: %s (%s expected)", data.Object, neynarCastType)
	}
	// check if the cast is a mention and skip if not
	mentionNeedle := fmt.Sprintf("@%s", n.username)
	isMention := !strings.HasPrefix(data.Text, mentionNeedle)
	// remove the username of the bot if it is a mention
	if isMention {
		data.Text = strings.TrimSpace(strings.TrimPrefix(data.Text, mentionNeedle))
	}
	// compose the message with the basic data
	message := &farcasterapi.APIMessage{
		IsMention: isMention,
		Author:    data.Author.Fid,
		Content:   data.Text,
		Hash:      data.Hash,
	}
	// include the parent parent cast info if it exists
	if data.ParentAuthor != nil {
		message.Parent = &farcasterapi.ParentAPIMessage{
			FID:  data.ParentAuthor.FID,
			Hash: data.ParentHash,
		}
	}
	// parse the embeds and include them in the message
	if len(data.Embeds) > 0 {
		message.Embeds = []string{}
		for _, embed := range data.Embeds {
			message.Embeds = append(message.Embeds, embed.Url)
		}
	}
	return message, nil
}

// neynarReq method sends a request to the Neynar API with the given URL, method,
// body and timeout. It returns the response body and an error if something goes
// wrong. It retries the request if it fails using sendRequest method.
func (n *NeynarAPI) neynarReq(ctx context.Context, url, method string, body []byte, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout*maxRetries)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("api_key", n.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return n.sendRequest(req)
}

// sendRequest method sends the given request and returns the response body and
// an error if something goes wrong. It retries the request if it fails.
func (n *NeynarAPI) sendRequest(req *http.Request) ([]byte, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
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
		log.Debugw("retrying request", "attempt", attempt+1, "url", req.URL.String(), "method", req.Method)
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

package hub

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vocdoni/vote-frame/farcasterapi"
	hubproto "github.com/vocdoni/vote-frame/farcasterapi/hub/proto"
	"github.com/zeebo/blake3"
	"go.vocdoni.io/dvote/log"
	"google.golang.org/protobuf/proto"
)

const (
	// endpoints
	ENDPOINT_CAST_BY_MENTION       = "castsByMention?fid=%d"
	ENDPOINT_SUBMIT_MESSAGE        = "submitMessage"
	ENDPOINT_USERNAME_PROOFS       = "userNameProofsByFid?fid=%d"
	ENDPOINT_USER_FOLLOWERs        = "linksByTargetFid?target_fid=%d"
	ENDPOINT_VERIFICATIONS         = "verificationsByFid?fid=%d"
	ENDPOINT_IDREGISTRY_BY_ADDRESS = "onChainIdRegistryEventByAddress?address=%s"
	// timeouts
	getCastByMentionTimeout = 15 * time.Second
	submitMessageTimeout    = 5 * time.Minute
	userdataTimeout         = 15 * time.Second
	userFollowersTimeout    = 15 * time.Second
	// message types
	MESSAGE_TYPE_CAST_ADD     = "MESSAGE_TYPE_CAST_ADD"
	MESSAGE_TYPE_USERPROOF    = "USERNAME_TYPE_FNAME"
	MESSAGE_TYPE_VERIFICATION = "MESSAGE_TYPE_VERIFICATION_ADD_ETH_ADDRESS"
	MESSAGE_TYPE_LINK         = "MESSAGE_TYPE_LINK_ADD"
	// other constants
	farcasterEpoch uint64 = 1609459200 // January 1, 2021 UTC
)

// Hub struct implements the farcasterapi.API interface and represents the
// API of a Farcaster Hub.
type Hub struct {
	fid      uint64
	privKey  []byte
	endpoint string
	auth     map[string]string
}

// Init initializes the API Hub with the given arguments.
// apiKeys must be a slice of strings with an even number of elements, where
// each pair of elements is a header and a key.
func NewHubAPI(apiEndpoint string, apiKeys []string) (*Hub, error) {
	h := &Hub{endpoint: apiEndpoint}
	// take the apikeys by group of two and set them as header/key
	if len(apiKeys)%2 != 0 {
		return nil, fmt.Errorf("invalid number of api keys")
	}
	h.auth = map[string]string{}
	for i := 0; i < len(apiKeys); i += 2 {
		h.auth[apiKeys[i]] = apiKeys[i+1]
	}
	return h, nil
}

// SetFarcasterUser sets the farcaster user with the given fid and signer
// privateKey in hex.
func (h *Hub) SetFarcasterUser(fid uint64, signerPrivKey string) error {
	var err error
	h.privKey, err = hex.DecodeString(strings.TrimPrefix(signerPrivKey, "0x"))
	if err != nil {
		return fmt.Errorf("error decoding signer: %w", err)
	}
	h.fid = fid
	return nil
}

// Stop stops the API Hub. It does nothing.
func (h *Hub) Stop() error {
	return nil
}

// LastMentions method returns the last mentions of the bot since the given
// timestamp. It returns the messages, the last timestamp and an error.
func (h *Hub) LastMentions(ctx context.Context, timestamp uint64) ([]*farcasterapi.APIMessage, uint64, error) {
	if h.fid == 0 {
		return nil, 0, fmt.Errorf("no farcaster user set")
	}
	if timestamp > farcasterEpoch {
		timestamp -= farcasterEpoch
	}
	internalCtx, cancel := context.WithTimeout(ctx, getCastByMentionTimeout)
	defer cancel()
	// download de json from API endpoint
	uri := fmt.Sprintf(ENDPOINT_CAST_BY_MENTION, h.fid)
	req, err := h.newRequest(internalCtx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating request: %w", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error downloading json: %w", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error("error closing response body")
		}
	}()
	if res.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("error downloading json: %s", res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading response body: %w", err)
	}
	// unmarshal the json
	mentions := &HubMessageResponse{}
	if err := json.Unmarshal(body, mentions); err != nil {
		return nil, 0, fmt.Errorf("error unmarshalling json: %w", err)
	}
	// filter messages and calculate the last timestamp
	lastTimestamp := uint64(0)
	messages := []*farcasterapi.APIMessage{}
	for _, m := range mentions.Messages {
		isMention := m.Data.Type == MESSAGE_TYPE_CAST_ADD && m.Data.CastAddBody != nil && m.Data.CastAddBody.Text != ""
		if !isMention {
			continue
		}
		if m.Data.Timestamp > timestamp {
			content, err := h.composeCastContent(m.Data.CastAddBody)
			if err != nil {
				log.Error(err)
			}
			messages = append(messages, &farcasterapi.APIMessage{
				IsMention: true,
				Content:   content,
				Author:    m.Data.From,
				Hash:      m.HexHash,
			})
			if m.Data.Timestamp > lastTimestamp {
				lastTimestamp = m.Data.Timestamp
			}
		}
	}
	// if there are no new casts, return an error
	if len(messages) == 0 {
		return nil, timestamp, farcasterapi.ErrNoNewCasts
	}
	// return the filtered messages and the last timestamp
	return messages, lastTimestamp + farcasterEpoch, nil
}

// Reply method sends a reply to the given targetFid and targetHash with the
// given content.
func (h *Hub) Reply(ctx context.Context, targetFid uint64, targetHash string, content string, embeds ...string) error {
	log.Infow("replying to cast", "msg", content)
	if h.fid == 0 {
		return fmt.Errorf("no farcaster user set")
	}
	// create the cast as a reply to the message with the parentFID provided
	// and the desired text
	bTargetHash, err := hex.DecodeString(strings.TrimPrefix(targetHash, "0x"))
	if err != nil {
		return fmt.Errorf("error decoding target hash: %s", err)
	}
	castEmbeds := []*hubproto.Embed{}
	if len(embeds) > 0 {
		for _, embed := range embeds {
			castEmbeds = append(castEmbeds, &hubproto.Embed{
				Embed: &hubproto.Embed_Url{Url: embed},
			})
		}
	}
	castAdd := &hubproto.CastAddBody{
		Text: content,
		// Mentions:          []uint64{targetFid},
		// MentionsPositions: []uint32{0},
		Parent: &hubproto.CastAddBody_ParentCastId{
			ParentCastId: &hubproto.CastId{
				Fid:  targetFid,
				Hash: bTargetHash,
			},
		},
		Embeds: castEmbeds,
	}
	// compose the message data with the message type, the bot FID, the current
	// timestamp, the network, and the cast add body
	msgData := &hubproto.MessageData{
		Type:      hubproto.MessageType_MESSAGE_TYPE_CAST_ADD,
		Fid:       h.fid,
		Timestamp: uint32(uint64(time.Now().Unix()) - farcasterEpoch),
		Network:   hubproto.FarcasterNetwork_FARCASTER_NETWORK_MAINNET,
		Body:      &hubproto.MessageData_CastAddBody{CastAddBody: castAdd},
	}
	// marshal the message data
	msgDataBytes, err := proto.Marshal(msgData)
	if err != nil {
		return fmt.Errorf("error marshalling message data: %s", err)
	}
	// calculate the hash of the message data
	hasher := blake3.New()
	if _, err := hasher.Write(msgDataBytes); err != nil {
		return fmt.Errorf("error hashing message data: %s", err)
	}
	hash := hasher.Sum(nil)[:20]
	// create the message with the hash scheme, the hash and the signature
	// scheme
	msg := &hubproto.Message{
		HashScheme:      hubproto.HashScheme_HASH_SCHEME_BLAKE3,
		Hash:            hash,
		SignatureScheme: hubproto.SignatureScheme_SIGNATURE_SCHEME_ED25519,
		Data:            msgData,
		DataBytes:       msgDataBytes,
	}
	// sign the message with the private key
	privateKey := ed25519.NewKeyFromSeed(h.privKey)
	signature := ed25519.Sign(privateKey, hash)
	signer := privateKey.Public().(ed25519.PublicKey)
	// set the signature and the signer to the message
	msg.Signature = signature
	msg.Signer = signer
	// marshal the message
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshalling message: %s", err)
	}
	// create a new context with a timeout
	internalCtx, cancel := context.WithTimeout(ctx, submitMessageTimeout)
	defer cancel()
	// submit the message to the API endpoint
	req, err := h.newRequest(internalCtx, http.MethodPost, ENDPOINT_SUBMIT_MESSAGE, bytes.NewBuffer(msgBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error submitting the message: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		// read the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %s", err)
		}
		return fmt.Errorf("error submitting the message: %s", string(body))
	}
	return nil
}

// UserDataByFID method returns the user data for the given FID. It includes the
// username, the custody address, the verification addresses and the signers.
func (h *Hub) UserDataByFID(ctx context.Context, fid uint64) (*farcasterapi.Userdata, error) {
	// create a intenal context with a timeout
	internalCtx, cancel := context.WithTimeout(ctx, userdataTimeout)
	defer cancel()
	// prepare the request to get username and custody address from the API
	usernameReq, err := h.newRequest(internalCtx, http.MethodGet, fmt.Sprintf(ENDPOINT_USERNAME_PROOFS, fid), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating user data request: %w", err)
	}
	// download the user data from the API and check for errors
	usernameRes, err := http.DefaultClient.Do(usernameReq)
	if err != nil {
		return nil, fmt.Errorf("error downloading user data: %w", err)
	}
	if usernameRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading user data: %s", usernameRes.Status)
	}
	// read the response body
	usernameBody, err := io.ReadAll(usernameRes.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading user data response body: %w", err)
	}
	// unmarshal the json
	userdata := &UserdataResponse{}
	if err := json.Unmarshal(usernameBody, userdata); err != nil {
		return nil, fmt.Errorf("error unmarshalling user data: %w", err)
	}
	// get the latest proof
	currentUserdata := &UsernameProofs{}
	lastUserdataTimestamp := uint64(0)
	for _, proof := range userdata.Proofs {
		// discard proofs that are not of the type we are looking for and
		// that are not from the user we are looking for
		if proof.Type != MESSAGE_TYPE_USERPROOF || proof.FID != fid {
			continue
		}
		// update the latest proof
		if proof.Timestamp > lastUserdataTimestamp {
			currentUserdata = proof
			lastUserdataTimestamp = proof.Timestamp
		}
	}
	// prepare the request to get verifications from the API
	verificationsReq, err := h.newRequest(internalCtx, http.MethodGet, fmt.Sprintf(ENDPOINT_VERIFICATIONS, fid), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating verifications request: %w", err)
	}
	// download the verifications from the API and check for errors
	verificationsRes, err := http.DefaultClient.Do(verificationsReq)
	if err != nil {
		return nil, fmt.Errorf("error downloading verifications: %w", err)
	}
	if verificationsRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading verifications: %s", verificationsRes.Status)
	}
	// read the response body
	verificationsBody, err := io.ReadAll(verificationsRes.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading verifications response body: %w", err)
	}
	// decode verifications json
	verificationsData := &VerificationsResponse{}
	if err := json.Unmarshal(verificationsBody, verificationsData); err != nil {
		return nil, fmt.Errorf("error unmarshalling verifications: %w", err)
	}
	// filter verifications addresses
	verifications := []string{}
	signersMap := make(map[string]struct{})
	for _, msg := range verificationsData.Messages {
		// if no data or verification data is found, skip. If the message data
		// type is not the one we are looking for, skip
		if msg.Data == nil || msg.Data.Type != MESSAGE_TYPE_VERIFICATION || msg.Data.Verification == nil || msg.Data.Signer == "" {
			log.Warnw("invalid verification message", "msg", msg)
			continue
		}
		verifications = append(verifications, msg.Data.Verification.Address)
		signersMap[msg.Data.Signer] = struct{}{}
	}
	signers := []string{}
	for signer := range signersMap {
		signers = append(signers, signer)
	}
	return &farcasterapi.Userdata{
		FID:                    fid,
		Username:               currentUserdata.Username,
		CustodyAddress:         currentUserdata.CustodyAddress,
		VerificationsAddresses: verifications,
		Signers:                signers,
	}, nil
}

// UserDataByVerificationAddress method returns the user data for the given
// verification addresses. It returns a slice of user data and an error. Hub
// does not implement this method.
func (h *Hub) UserDataByVerificationAddress(ctx context.Context, address []string) ([]*farcasterapi.Userdata, error) {
	return nil, fmt.Errorf("not implemented")
}

// UserFollowers method returns the FIDs of the followers of the user with the
// given id. If something goes wrong, it returns an error.
func (h *Hub) UserFollowers(ctx context.Context, fid uint64) ([]uint64, error) {
	// create a intenal context with a timeout
	internalCtx, cancel := context.WithTimeout(ctx, userFollowersTimeout)
	defer cancel()
	// prepare the request to get the followers from the API
	uri := fmt.Sprintf(ENDPOINT_USER_FOLLOWERs, fid)
	req, err := h.newRequest(internalCtx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating user followers request: %w", err)
	}
	// download the followers from the API and check for errors
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error downloading user followers: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading user followers: %s", res.Status)
	}
	// read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading user followers response body: %w", err)
	}
	// unmarshal the json
	followersResponse := &HubMessageResponse{}
	if err := json.Unmarshal(body, followersResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling user followers: %w", err)
	}
	// filter the followers FIDs and return them
	followersFids := []uint64{}
	for _, msg := range followersResponse.Messages {
		if msg.Data.Type != MESSAGE_TYPE_LINK {
			continue
		}
		followersFids = append(followersFids, msg.Data.From)
	}
	return followersFids, nil
}

// ChannelFIDs method returns the FIDs of the users that follow the channel with
// the given id. If something goes wrong, it returns an error. It return an
// specific error if the channel does not exist to be handled by the caller.
func (n *Hub) ChannelFIDs(ctx context.Context, channelID string, _ chan int) ([]uint64, error) {
	return nil, fmt.Errorf("hub api does not support channels yet")
}

// ChannelExists method returns a boolean indicating if the channel with the
// given id exists. If something goes wrong checking the channel existence,
// it returns an error.
func (n *Hub) ChannelExists(channelID string) (bool, error) {
	return false, fmt.Errorf("hub api does not support channels yet")
}

// WebhookHandler method handles the incoming webhooks. Hub does not implement
// this method.
func (h *Hub) WebhookHandler(_ []byte) error {
	return fmt.Errorf("not implemented")
}

// SignersFromFID method returns the signers for the given FID. It returns a
// slice of signers and an error. Hub does not implement this method.
func (h *Hub) SignersFromFID(fid uint64) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// newRequest method creates a new http request with the given method, uri and
// body. It returns the request and an error.
func (h *Hub) newRequest(ctx context.Context, method string, uri string, body io.Reader) (*http.Request, error) {
	endpoint := fmt.Sprintf("%s/%s", h.endpoint, uri)
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	if h.auth != nil {
		for k, v := range h.auth {
			if k == "" || v == "" {
				continue
			}
			req.Header.Set(k, v)
		}
	}
	return req, nil
}

// composeCastContent method composes the cast content with the given body. It
// returns the content and an error. If the body is nil, it returns an empty
// string and no error. If the body is not nil, it replaces the mentions with
// the usernames and returns the content.
func (h *Hub) composeCastContent(body *HubCastAddBody) (string, error) {
	if body == nil {
		return "", nil
	}
	content := body.Text
	for i := len(body.Mentions) - 1; i >= 0; i-- {
		fid := body.Mentions[i]
		pos := body.MentionsPositions[i]
		if pos == 0 && fid == h.fid {
			continue
		}
		user, err := h.UserDataByFID(context.Background(), fid)
		if err != nil {
			return "", err
		}
		content = content[:pos] + "@" + user.Username + content[pos:]
	}
	return content, nil
}

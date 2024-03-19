package hub

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	hubproto "github.com/vocdoni/vote-frame/farcasterapi/hub/proto"
	"github.com/zeebo/blake3"
	"google.golang.org/protobuf/proto"
)

var mentionRgx = regexp.MustCompile(`(@\S+)`)

// newAddCastBody method creates a new cast add body with the given content, the
// mentions fids and the embeds. It returns the cast add body and an error. It
// returns an error if the farcaster user is not set or there is an error
// decomposing the content. It replaces the mentions with the usernames and
// creates the cast add body with the mentions positions and the embeds.
func (h *Hub) newAddCastBody(content string, mentionFIDs []uint64, embeds ...string) (*hubproto.CastAddBody, error) {
	if h.fid == 0 {
		return nil, fmt.Errorf("no farcaster user set")
	}
	// decompose the content and the mentions
	castBody, err := h.decomposeContent(content, mentionFIDs)
	if err != nil {
		return nil, fmt.Errorf("error decomposing content: %s", err)
	}
	// convert the mentions positions to uint32
	mentionsPositions := make([]uint32, len(castBody.MentionsPositions))
	for i, pos := range castBody.MentionsPositions {
		mentionsPositions[i] = uint32(pos)
	}
	// create the cast add body
	castAdd := &hubproto.CastAddBody{
		Text:              castBody.Text,
		Embeds:            []*hubproto.Embed{},
		Mentions:          castBody.Mentions,
		MentionsPositions: mentionsPositions,
	}
	// if there are embeds urls, add them to the cast add body
	if len(embeds) > 0 {
		for _, embed := range embeds {
			castAdd.Embeds = append(castAdd.Embeds, &hubproto.Embed{
				Embed: &hubproto.Embed_Url{Url: embed},
			})
		}
	}
	return castAdd, nil
}

// buildAndSignAddCastBody method builds and signs the given cast add body. It
// returns the message bytes and an error. It creates the message data with the
// message type, the bot FID, the current timestamp, the network, and the cast
// add body. It marshals the message data and calculates the hash of the message
// data. It creates the message with the hash scheme, the hash and the signature
// scheme. It signs the message with the private key and sets the signature and
// the signer to the message. It marshals the message and returns the message
// bytes.
func (h *Hub) buildAndSignAddCastBody(castAddBody *hubproto.CastAddBody) ([]byte, error) {
	// compose the message data with the message type, the bot FID, the current
	// timestamp, the network, and the cast add body
	msgData := &hubproto.MessageData{
		Type:      hubproto.MessageType_MESSAGE_TYPE_CAST_ADD,
		Fid:       h.fid,
		Timestamp: uint32(uint64(time.Now().Unix()) - farcasterEpoch),
		Network:   hubproto.FarcasterNetwork_FARCASTER_NETWORK_MAINNET,
		Body:      &hubproto.MessageData_CastAddBody{CastAddBody: castAddBody},
	}
	// marshal the message data
	msgDataBytes, err := proto.Marshal(msgData)
	if err != nil {
		return nil, fmt.Errorf("error marshalling message data: %s", err)
	}
	// calculate the hash of the message data
	hasher := blake3.New()
	hasher.Write(msgDataBytes)
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
		return nil, fmt.Errorf("error marshalling message: %s", err)
	}
	return msgBytes, nil
}

// composeCastContent method composes the cast content with the given body. It
// returns the content and an error. If the body is nil, it returns an empty
// string and no error. If the body is not nil, it replaces the mentions with
// the usernames and returns the content.
func (h *Hub) composeCastContent(body *hubCastAddBody) (string, error) {
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

// decomposeContent method decomposes the content with the given mentions. It
// returns the body and an error. If the mentions are not the same length as
// the fids, it returns an error. If the mentions are the same length as the
// fids, it calculates the position of the mentions, removes them from the
// content, and returns the body.
func (h *Hub) decomposeContent(content string, mentionFids []uint64) (*hubCastAddBody, error) {
	// get the mentions from the content
	usernames := mentionRgx.FindAllString(content, -1)
	if len(usernames) == 0 {
		return &hubCastAddBody{
			Text: content,
		}, nil
	}
	// check if the mentions have the same length as the fids provided
	if len(mentionFids) != len(usernames) {
		return nil, fmt.Errorf("invalid mentions")
	}
	// get the positions of the mentions in the content
	usernamePositions := mentionRgx.FindAllStringIndex(content, -1)
	mentionsPositions := make([]uint64, len(mentionFids))
	for i := range usernames {
		mentionsPositions[i] = uint64(usernamePositions[i][0])
	}
	// remove mentions from the content
	content = mentionRgx.ReplaceAllString(content, "")
	// build the body with the cleaned content and the mentions
	return &hubCastAddBody{
		Text:              content,
		Mentions:          mentionFids,
		MentionsPositions: mentionsPositions,
	}, nil
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

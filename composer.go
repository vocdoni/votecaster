package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/vochain/transaction/proofs/farcasterproof"
)

const (
	composerActionWebappPath     = "/app"
	composerActionWebappFragment = "/composer"
	composerActionTokenQuery     = "token"
	composerActionQuestionQuery  = "question"
)

// safeURL function returns a safe URL string from the provided URL. It returns
// an empty string if the URL is nil. The resulting string will have the format:
// scheme://host/path#fragment?query. If the URL has no path, query or fragment,
// they will be omitted. The query parameters will be encoded.
func safeURL(url *url.URL) string {
	if url == nil {
		return ""
	}
	strURL := fmt.Sprintf("%s://%s", url.Scheme, url.Host)
	if url.Path != "" {
		strURL += url.Path
	}
	if url.Fragment != "" {
		strURL += fmt.Sprintf("#%s", url.Fragment)
	}
	if encoded := url.Query().Encode(); encoded != "" {
		queryParams := []string{}
		for key, values := range url.Query() {
			for _, value := range values {
				queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
			}
		}
		strURL += fmt.Sprintf("?%s", strings.Join(queryParams, "&"))
	}
	return strURL
}

// composerActionCast is the structure of the cast field in the action message
// state, used to extract the text of the cast that launched the composer action
type composerActionCast struct {
	Cast struct {
		Parent string   `json:"parent"`
		Text   string   `json:"text"`
		Embeds []string `json:"embeds"`
	} `json:"cast"`
}

func (v *vocdoniHandler) composerMetadataHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	res, err := json.Marshal(ComposerActionMetadata{
		Type:        "composer",
		Name:        "Votecaster",
		Icon:        "project-roadmap",
		Description: "Run verifiable polls",
		ImageURL:    serverURL + "/app/logo-farcastervote-action.png",
		Action: struct {
			Type string `json:"type"`
		}{Type: "post"},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal composer metadata: %w", err)
	}
	return ctx.Send(res, http.StatusOK)
}

func (v *vocdoniHandler) composerActionHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// decode the packet from the message
	packet := &FrameSignaturePacket{}
	if err := json.Unmarshal(msg.Data, packet); err != nil {
		return fmt.Errorf("failed to unmarshal frame signature packet: %w", err)
	}
	// decode the message bytes
	messageBytes, err := hex.DecodeString(packet.TrustedData.MessageBytes)
	if err != nil {
		return fmt.Errorf("failed to decode message bytes: %w", err)
	}
	// verify the frame signature and get the action message and the fid of the
	// user
	actionMessage, _, userFID, err := farcasterproof.VerifyFrameSignature(messageBytes)
	if err != nil {
		return fmt.Errorf("failed to verify frame signature: %w", err)
	}
	// Ensure the user profile exists
	if exists := v.db.UserExists(userFID); !exists {
		return fmt.Errorf("user not found: %d", userFID)
	}
	// get the token of the user from the database, or generate a new one
	var token string
	if authTokens, err := v.db.UserAuthorizations(userFID); err != nil {
		// generate new token for the user
		uuidToken, err := uuid.NewRandom()
		if err != nil {
			return fmt.Errorf("could not generate token: %v", err)
		}
		token = uuidToken.String()
	} else {
		// use the first token found
		token = authTokens[0]
	}
	// compose the action URL with the token of the user
	actionURL, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("could not parse server URL: %v", err)
	}
	actionURL.Path = composerActionWebappPath
	actionURL.Fragment = composerActionWebappFragment
	query := actionURL.Query()
	query.Set(composerActionTokenQuery, token)
	// URL-decode the cast from the action message state, and extract the text
	// to be used as a question in the composer action form, if any error occurs
	// ignore it and continue
	if decodedCast, err := url.QueryUnescape(string(actionMessage.GetState())); err == nil {
		cast := &composerActionCast{}
		if err := json.Unmarshal([]byte(decodedCast), cast); err == nil {
			// add the text of the cast that launched the composer action to the URL
			// as a question
			if cast.Cast.Text != "" {
				query.Set(composerActionQuestionQuery, url.QueryEscape(cast.Cast.Text))
			}
		}
	}
	actionURL.RawQuery = query.Encode()
	url := safeURL(actionURL)
	// encode the response with the resulting action URL
	var response []byte
	if response, err = json.Marshal(ComposerActionResponse{
		Type:  "form",
		Title: "Create a verifiable poll",
		URL:   url,
	}); err != nil {
		return err
	}
	// store the token in the database and send the response
	v.addAuthTokenFunc(userFID, token)
	log.Infow("new composer action user access", "fid", userFID, "url", url)
	return ctx.Send(response, http.StatusOK)
}

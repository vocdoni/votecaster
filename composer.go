package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/vochain/transaction/proofs/farcasterproof"
)

// composerActionCast is the structure of the cast field in the action message
// state, used to extract the text of the cast that launched the composer action
type composerActionCast struct {
	Cast struct {
		Parent string   `json:"parent"`
		Text   string   `json:"text"`
		Embeds []string `json:"embeds"`
	} `json:"cast"`
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
	// generate new token for the user
	token, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("could not generate token: %v", err)
	}
	// compose the action URL with the token of the user
	actionURL := fmt.Sprintf("%s/app#/composer?token=%s", serverURL, token.String())
	// URL-decode the cast from the action message state, and extract the text
	// to be used as a question in the composer action form, if any error occurs
	// ignore it and continue
	if decodedCast, err := url.QueryUnescape(string(actionMessage.GetState())); err == nil {
		cast := &composerActionCast{}
		if err := json.Unmarshal([]byte(decodedCast), cast); err == nil {
			// add the text of the cast that launched the composer action to the URL
			// as a question
			if cast.Cast.Text != "" {
				actionURL += "&question=" + url.QueryEscape(cast.Cast.Text)
			}
		}
	}
	// encode the response with the resulting action URL
	var response []byte
	if response, err = json.Marshal(ComposerActionResponse{
		Type:  "form",
		Title: "Create a blockchain Poll",
		URL:   actionURL,
	}); err != nil {
		return err
	}
	// store the token in the database and send the response
	v.addAuthTokenFunc(userFID, token.String())
	return ctx.Send(response, http.StatusOK)
}

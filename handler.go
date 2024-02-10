package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/util"
)

type vocodniHandler struct {
	cli *apiclient.HTTPclient
}

func NewVocdoniHandler(apiEndpoint, accountPrivKey string) (*vocodniHandler, error) {
	// Get the vocdoni account
	if accountPrivKey == "" {
		accountPrivKey = util.RandomHex(32)
		log.Infow("generated new vocdoni account", "privkey", accountPrivKey)
	}

	// Create the Vocdoni API client
	hostURL, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse apiEndpoint: %w", err)
	}
	log.Debugf("connecting to %s", hostURL.String())
	token := uuid.New()
	cli, err := apiclient.NewHTTPclient(hostURL, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to create apiclient: %w", err)
	}
	log.Infow("using bearer token", "token", token.String())

	if err := cli.SetAccount(accountPrivKey); err != nil {
		return nil, fmt.Errorf("failed to set account: %w", err)
	}

	// Create the account if it doesn't exist and return the handler
	return &vocodniHandler{
		cli: cli,
	}, ensureAccountExist(cli)
}

func (v *vocodniHandler) showElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	log.Infow("received show election request", "electionID", ctx.URLParam("electionID"))

	// create a PNG image with the election description
	election, err := v.cli.Election(electionID)
	if err != nil {
		return fmt.Errorf("failed to fetch election: %w", err)
	}
	text := strings.Builder{}
	text.WriteString(fmt.Sprintf(election.Metadata.Title["default"]))
	text.WriteString("\n----------------------------------------\n\n")
	text.WriteString(fmt.Sprintf("> Started %s ago\n", time.Since(election.StartDate).Round(time.Minute).String()))
	text.WriteString(fmt.Sprintf("> Remaining time %s\n", time.Until(election.EndDate).Round(time.Minute).String()))
	// text.WriteString(fmt.Sprintf("> Census size %d\n", election.Census.MaxCensusSize))
	text.WriteString(fmt.Sprintf("> Executed on network %s\n", v.cli.ChainID()))
	png, err := textToImage(text.String(), "#33ff33", "#000000", Pixeloid, 42)
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}

	// send the response
	response := strings.ReplaceAll(frame(frameVote), "{image}", base64.StdEncoding.EncodeToString(png))
	response = strings.ReplaceAll(response, "{processID}", ctx.URLParam("electionID"))
	for i, r := range election.Metadata.Questions[0].Choices {
		response = strings.ReplaceAll(response, fmt.Sprintf("{option%d}", i), r.Title["default"])
	}

	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func (v *vocodniHandler) vote(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the electionID from the URL and the frame signature packet from the body of the request
	electionID := ctx.URLParam("electionID")
	electionIDbytes, err := hex.DecodeString(electionID)
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}

	election, err := v.cli.Election(electionIDbytes)
	if err != nil {
		log.Warnw("failed to fetch election", "error", err)
		png, err := textToImage(fmt.Sprintf("Error: %s", err.Error()), "#ff0000", "#000000", Pixeloid, 36)
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameError), "{image}", base64.StdEncoding.EncodeToString(png))
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	packet := &FrameSignaturePacket{}
	if err := json.Unmarshal(msg.Data, packet); err != nil {
		return fmt.Errorf("failed to unmarshal frame signature packet: %w", err)
	}
	// cast the vote
	nullifier, voterID, err := vote(packet, electionIDbytes, election.Census.CensusRoot, v.cli)

	// handle the vote result
	if errors.Is(err, ErrNotInCensus) {
		log.Infow("participant not in the census", "voterID", fmt.Sprintf("%x", voterID))
		png, err := textToImage("Sorry, you are not elegible! 🔍", "#ff0000", "#000000", Pixeloid, 52)
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameNotElegible), "{image}", base64.StdEncoding.EncodeToString(png))
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	if errors.Is(err, ErrAlreadyVoted) {
		log.Infow("participant already voted", "voterID", fmt.Sprintf("%x", voterID))
		png, err := textToImage("You already voted!", "#ff0000", "#000000", Pixeloid, 52)
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameAlreadyVoted), "{image}", base64.StdEncoding.EncodeToString(png))
		response = strings.ReplaceAll(response, "{nullifier}", fmt.Sprintf("%x", nullifier))
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	if err != nil {
		log.Warnw("failed to vote", "error", err)
		png, err := textToImage(fmt.Sprintf("Error: %s", err.Error()), "#ff0000", "#000000", Pixeloid, 36)
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameError), "{image}", base64.StdEncoding.EncodeToString(png))
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	response := strings.ReplaceAll(frame(frameAfterVote), "{nullifier}", fmt.Sprintf("%x", nullifier))
	response = strings.ReplaceAll(response, "{processID}", electionID)
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func (v *vocodniHandler) results(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID := ctx.URLParam("electionID")
	log.Infow("received results request", "electionID", electionID)
	electionIDbytes, err := hex.DecodeString(electionID)
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}

	// get the election from the vochain and create a PNG image with the results
	election, err := v.cli.Election(electionIDbytes)
	if err != nil {
		return fmt.Errorf("failed to fetch election: %w", err)
	}
	if election.Results == nil {
		return fmt.Errorf("election results not ready")
	}

	text := strings.Builder{}
	for _, r := range election.Metadata.Questions[0].Choices {
		_, err := text.WriteString(fmt.Sprintf("%s: %s\n",
			r.Title["default"],
			election.Results[0][r.Value].String(),
		))
		if err != nil {
			return fmt.Errorf("failed to write results: %w", err)
		}
	}

	png, err := textToImage(text.String(), "#33ff33", "#000000", Pixeloid, 42)
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}

	response := strings.ReplaceAll(frame(frameResults), "{image}", base64.StdEncoding.EncodeToString(png))
	response = strings.ReplaceAll(response, "{processID}", electionID)
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

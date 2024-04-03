package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vocdoni/vote-frame/bot"
	"github.com/vocdoni/vote-frame/bot/poll"
	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/farcasterapi/neynar"
	"github.com/vocdoni/vote-frame/shortener"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
)

// initBot helper function initializes the bot and starts listening for new polls
// to create elections
func initBot(ctx context.Context, handler *vocdoniHandler, api farcasterapi.API,
	defaultCensus *CensusInfo,
) (*bot.Bot, error) {
	voteBot, err := bot.New(bot.BotConfig{
		API: api,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}
	voteBot.Start(ctx)
	// handle new messages in background
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-voteBot.Messages:
				// check if the message is a poll and create an election
				user, poll, isPool, err := voteBot.PollMessageHandler(ctx, msg, maxElectionDuration)
				if err == nil && isPool {
					log.Infow("new poll received, creating election...",
						"poll", poll,
						"userdata", user,
						"msg-hash", msg.Hash)
					if err := pollToCast(ctx, handler, poll, user, msg, voteBot, defaultCensus); err != nil {
						log.Errorf("error creating election: %s", err)
					}
					continue
				}
				// check if the message is a mute request and mute the user
				user, userToMute, isMuteRequest, err := voteBot.MuteRequestHandler(ctx, msg)
				if err == nil && isMuteRequest {
					log.Infow("mute request received, muting user...",
						"userdata", user,
						"msg-hash", msg.Hash,
						"muted-userdata", userToMute)
					// handle mute request
					continue
				}
			}
		}
	}()
	return voteBot, nil
}

// pollToCast helper function creates an election from a poll and sends the poll
// URL to the user replying to the message with the poll frame. If something
// goes wrong it returns an error.
func pollToCast(ctx context.Context, handler *vocdoniHandler, poll *poll.Poll,
	user *farcasterapi.Userdata, msg *farcasterapi.APIMessage, voteBot *bot.Bot,
	defaultCensus *CensusInfo,
) error {
	description := &ElectionDescription{
		Question:  poll.Question,
		Options:   poll.Options,
		Duration:  poll.Duration,
		Overwrite: false,
	}
	profile := &FarcasterProfile{
		FID:           user.FID,
		Username:      user.Username,
		Custody:       user.CustodyAddress,
		Verifications: user.VerificationsAddresses,
	}
	electionID, err := handler.createAndSaveElectionAndProfile(description,
		defaultCensus, profile, true, false, ElectionSourceBot)
	if err != nil {
		return fmt.Errorf("error creating election: %w", err)
	}
	log.Infow("election created",
		"electionID", electionID,
		"poll", poll)
	frameUrl := fmt.Sprintf("%s/%s", serverURL, electionID.String())
	shortenedUrl, err := shortener.ShortURL(ctx, frameUrl)
	if err != nil {
		// if shortening fails, use the original url
		shortenedUrl = frameUrl
	}
	if err := voteBot.ReplyWithPollURL(ctx, msg, shortenedUrl); err != nil {
		return fmt.Errorf("error replying to poll: %s", err)
	}
	log.Infow("poll reply sent",
		"frame-url", frameUrl,
		"author", msg.Author,
		"msg-hash", msg.Hash)
	return nil
}

// neynarWebhook helper function returns a function that handles neynar webhooks.
// It verifies the request and handles the webhook using the neynar client.
func neynarWebhook(neynarcli *neynar.NeynarAPI, webhookSecret string) func(*apirest.APIdata, *httprouter.HTTPContext) error {
	return func(msg *apirest.APIdata, h *httprouter.HTTPContext) error {
		neynarSig := h.Request.Header.Get("X-Neynar-Signature")
		verified, err := neynar.VerifyRequest(webhookSecret, neynarSig, msg.Data)
		if err != nil {
			log.Errorf("error verifying request: %s", err)
			return h.Send([]byte("error verifying request"), http.StatusBadRequest)
		}
		if !verified {
			log.Error("request not verified")
			return h.Send([]byte("request not verified"), http.StatusUnauthorized)
		}
		if err := neynarcli.WebhookHandler(msg.Data); err != nil {
			log.Errorf("error handling webhook: %s", err)
			return fmt.Errorf("error handling webhook: %s", err)
		}
		return h.Send([]byte("ok"), http.StatusOK)
	}
}

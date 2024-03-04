package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vocdoni/vote-frame/bot"
	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/farcasterapi/neynar"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
)

// initBot helper function initializes the bot and starts listening for new polls
// to create elections
func initBot(ctx context.Context, handler *vocdoniHandler, api farcasterapi.API,
	census *CensusInfo,
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
				user, poll, err := voteBot.PollMessageHandler(ctx, msg, maxElectionDuration)
				if err != nil {
					log.Errorf("error handling poll message: %s", err)
					continue
				}
				log.Infow("new poll received, creating election...",
					"poll", poll,
					"userdata", user,
					"msg-hash", msg.Hash)
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
				electionID, err := handler.createAndSaveElectionAndProfile(description, census, profile, true, ElectionSourceBot)
				if err != nil {
					log.Errorf("error creating election: %s", err)
					continue
				}
				log.Infow("election created",
					"electionID", electionID,
					"poll", poll)
				frameUrl := fmt.Sprintf("%s/%s", serverURL, electionID.String())
				shortenedUrl, err := ShortElectionURL(ctx, frameUrl)
				if err != nil {
					log.Errorf("error shortening election url: %s", err)
					continue
				}

				if err := voteBot.ReplyWithPollURL(ctx, msg, shortenedUrl); err != nil {
					log.Errorf("error replying to poll: %s", err)
					continue
				}
				log.Infow("poll reply sent",
					"frame-url", frameUrl,
					"author", msg.Author,
					"msg-hash", msg.Hash)
			}
		}
	}()
	return voteBot, nil
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

package bot

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/vocdoni/vote-frame/farcasterapi"
	"go.vocdoni.io/dvote/log"
)

// defaultCoolDown is the default time to wait between casts
const defaultCoolDown = time.Second * 10

// BotConfig is the configuration definition for the bot, it includes the API
// instance and the cool down time between casts (default is 10 seconds)
type BotConfig struct {
	API      farcasterapi.API
	CoolDown time.Duration
}

// Bot struct represents a bot that listens for new casts and sends them to a
// channel, it also has a cool down time to avoid spamming the API and a last
// cast timestamp to retrieve new casts from that point, ensuring no cast is
// missed or duplicated
type Bot struct {
	api      farcasterapi.API
	ctx      context.Context
	cancel   context.CancelFunc
	coolDown time.Duration
	lastCast uint64
	Messages chan *farcasterapi.APIMessage
}

// New function creates a new bot with the given configuration, it returns an
// error if the API is not set in the configuration.
func New(config BotConfig) (*Bot, error) {
	if config.API == nil {
		return nil, ErrAPINotSet
	}
	if config.CoolDown == 0 {
		config.CoolDown = defaultCoolDown
	}
	return &Bot{
		api:      config.API,
		coolDown: config.CoolDown,
		lastCast: uint64(time.Now().Unix()),
		Messages: make(chan *farcasterapi.APIMessage),
	}, nil
}

// Start function starts the bot, it listens for new casts and sends them to the
// Messages channel. It does this in a goroutine to avoid blocking the main and
// every cool down time.
func (b *Bot) Start(ctx context.Context) {
	b.ctx, b.cancel = context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(b.coolDown)
		for {
			select {
			case <-b.ctx.Done():
				return
			default:
				// retrieve new messages from the last cast
				messages, lastCast, err := b.api.LastMentions(b.ctx, b.lastCast)
				if err != nil && !errors.Is(err, farcasterapi.ErrNoNewCasts) {
					log.Errorw(err, "error retrieving new casts")
					continue
				}
				b.lastCast = lastCast
				if len(messages) > 0 {
					for _, msg := range messages {
						log.Infow("new bot cast", "from", msg.Author, "message", msg.Content)
						b.Messages <- msg
					}
				}
				// wait for the cool down time
				<-ticker.C
				ticker.Reset(b.coolDown)
			}
		}
	}()
}

// Stop function stops the bot and its goroutine, and closes the Messages channel.
func (b *Bot) Stop() {
	if err := b.api.Stop(); err != nil {
		log.Errorf("error stopping bot: %s", err)
	}
	b.cancel()
	close(b.Messages)
}

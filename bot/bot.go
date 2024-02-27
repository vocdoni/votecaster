package bot

import (
	"context"
	_ "embed"
	"time"

	"github.com/vocdoni/vote-frame/farcasterapi"
	"go.vocdoni.io/dvote/log"
)

// defaultCoolDown is the default time to wait between casts
const defaultCoolDown = time.Second * 10

type BotConfig struct {
	API      farcasterapi.API
	CoolDown time.Duration
}

type Bot struct {
	api      farcasterapi.API
	ctx      context.Context
	cancel   context.CancelFunc
	coolDown time.Duration
	lastCast uint64
	Messages chan *farcasterapi.APIMessage
}

func New(config BotConfig) (*Bot, error) {
	log.Infow("initializing bot", "config", config)
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

func (b *Bot) Start(ctx context.Context) {
	b.ctx, b.cancel = context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(b.coolDown)
		for {
			select {
			case <-b.ctx.Done():
				return
			default:
				log.Debugw("checking for new casts", "last-cast", b.lastCast)
				// retrieve new messages from the last cast
				messages, lastCast, err := b.api.LastMentions(b.ctx, b.lastCast)
				if err != nil && err != ErrNoNewCasts {
					log.Errorf("error retrieving new casts: %s", err)
				}
				b.lastCast = lastCast
				if len(messages) > 0 {
					for _, msg := range messages {
						b.Messages <- msg
					}
				} else {
					log.Debugw("no new casts", "last-cast", b.lastCast)
				}
				<-ticker.C
			}
		}
	}()
}

func (b *Bot) Stop() {
	if err := b.api.Stop(); err != nil {
		log.Errorf("error stopping bot: %s", err)
	}
	b.cancel()
	close(b.Messages)
}

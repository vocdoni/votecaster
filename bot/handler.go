package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vocdoni/vote-frame/bot/poll"
	"github.com/vocdoni/vote-frame/farcasterapi"
)

var (
	// PollReplyTemplate is the template for the reply to a cast with a poll. It
	// must be formatted with the poll URL.
	PollReplyTemplate = `🗳️ Your verifiable poll has been created using farcaster.vote

Now copy & paste the URL of the frame into a Cast to share your poll with others!

👉 %s`
)

func (b *Bot) PollMessageHandler(ctx context.Context, msg *farcasterapi.APIMessage, maxDuration time.Duration) (*farcasterapi.Userdata, *poll.Poll, error) {
	// when a new cast is received, check if it is a mention and if
	// it is not, continue to the next cast
	if !msg.IsMention {
		return nil, nil, nil
	}
	// try to parse the message as a poll, if it fails continue to
	// the next cast
	pollConf := poll.DefaultConfig
	pollConf.DefaultDuration = maxDuration
	poll, err := poll.ParseString(msg.Content, pollConf)
	if err != nil {
		return nil, nil, errors.Join(ErrParsingPoll, err)
	}
	// get the user data such as username, custody address and
	// verification addresses to create the election frame
	userdata, err := b.api.UserDataByFID(ctx, msg.Author)
	if err != nil {
		return nil, nil, errors.Join(ErrGettingUserData, err)
	}
	return userdata, poll, nil
}

func (b *Bot) ReplyWithPollURL(ctx context.Context, msg *farcasterapi.APIMessage, pollURL string) error {
	if err := b.api.Reply(ctx, msg.Author, msg.Hash, fmt.Sprintf(PollReplyTemplate, pollURL)); err != nil {
		return errors.Join(ErrReplyingToCast, err)
	}
	return nil
}

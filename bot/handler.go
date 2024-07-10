package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vocdoni/vote-frame/bot/poll"
	"github.com/vocdoni/vote-frame/farcasterapi"
)

// PollReplyTemplate is the template for the reply to a cast with a poll. It
// must be formatted with the poll URL.
var PollReplyTemplate = `🗳️ Your poll is ready! And just so you know, we used the Vocdoni blockchain to make it verifiable and censorship-resistant! 😎

%s

Now copy the URL, paste the Frame into a cast and share it with others!

👇`

// PollMessageHandler is a function that handles a new cast and checks if it is a
// poll, if it is, it parses the poll and returns the user data and the poll. If
// it is not a poll, it returns false. It returns an error if something goes
// wrong, including when message parsing fails.
func (b *Bot) PollMessageHandler(ctx context.Context, msg *farcasterapi.APIMessage, maxDuration time.Duration) (*farcasterapi.Userdata, *poll.Poll, bool, error) {
	// when a new cast is received, check if it is a mention and if
	// it is not, continue to the next cast
	if msg == nil || !msg.IsMention {
		return nil, nil, false, nil
	}
	// try to parse the message as a poll, if it fails continue to
	// the next cast
	pollConf := poll.DefaultConfig
	pollConf.DefaultDuration = maxDuration
	poll, err := poll.ParseString(msg.Content, pollConf)
	if err != nil {
		return nil, nil, false, errors.Join(ErrParsingPoll, err)
	}
	// get the user data such as username, custody address and
	// verification addresses to create the election frame
	userdata, err := b.api.UserDataByFID(ctx, msg.Author)
	if err != nil {
		return nil, nil, true, errors.Join(ErrGettingUserData, err)
	}
	return userdata, poll, true, nil
}

func (b *Bot) MuteRequestHandler(ctx context.Context, msg *farcasterapi.APIMessage) (*farcasterapi.Userdata, *farcasterapi.APIMessage, bool, error) {
	// when a new cast is received, check if it is a mention and if
	// it is not, continue to the next cast
	if msg == nil || !msg.IsMention {
		return nil, nil, false, nil
	}
	if msg.Parent == nil {
		return nil, nil, false, nil
	}
	// check if the content of the cast is a mute request, if it is
	if strings.TrimSpace(msg.Content) != muteRequestContent {
		return nil, nil, false, nil
	}
	// get the user data such as username, custody address and verification
	// addresses of the user that wants to mute the creator of the poll
	userdata, err := b.api.UserDataByFID(ctx, msg.Author)
	if err != nil {
		return nil, nil, true, errors.Join(ErrGettingUserData, err)
	}
	// get the parent message to recover the poll
	parentMsg, err := b.api.GetCast(ctx, msg.Parent.FID, msg.Parent.Hash)
	if err != nil {
		return nil, nil, true, errors.Join(ErrGettingParentCast, err)
	}
	return userdata, parentMsg, true, nil
}

func (b *Bot) ReplyWithPollURL(ctx context.Context, msg *farcasterapi.APIMessage, pollURL string) error {
	if err := b.api.Reply(ctx, msg, fmt.Sprintf(PollReplyTemplate, pollURL), nil, pollURL); err != nil {
		return errors.Join(ErrReplyingToCast, err)
	}
	return nil
}

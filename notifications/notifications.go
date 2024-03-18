package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/log"
)

const (
	DefaultListenCoolDown = 30 * time.Second
	DefaultSendCoolDown   = 500 * time.Millisecond
	NotificationMessage   = `Hey @%s!

There is a new poll in which you're eligible to vote!

🗳️ Cast your vote on the frame below`
)

type NotificationManager struct {
	ctx            context.Context
	cancel         context.CancelFunc
	db             *mongo.MongoStorage
	api            farcasterapi.API
	listenCoolDown time.Duration
}

func New(ctx context.Context, db *mongo.MongoStorage, api farcasterapi.API, listenCoolDown time.Duration) *NotificationManager {
	ctx, cancel := context.WithCancel(ctx)
	return &NotificationManager{
		ctx:            ctx,
		cancel:         cancel,
		db:             db,
		api:            api,
		listenCoolDown: listenCoolDown,
	}
}

func (nm *NotificationManager) Start() {
	go func() {
		for {
			select {
			case <-nm.ctx.Done():
				return
			case <-time.After(nm.listenCoolDown):
				notifications, err := nm.db.LastNotifications()
				if err != nil {
					log.Errorf("error getting notifications: %s", err)
					continue
				}
				if err := nm.sendNotifications(notifications); err != nil {
					log.Errorf("error sending notifications: %s", err)
				}
			}
		}
	}()
}

func (nm *NotificationManager) Stop() {
	nm.cancel()
}

func (nm *NotificationManager) sendNotifications(notification []mongo.Notification) error {
	for _, n := range notification {
		log.Debugw("sending notification", "username", n.Username, "userID", n.UserID, "electionID", n.ElectionID, "frameURL", n.FrameUrl)
		msg := fmt.Sprintf(NotificationMessage, n.Username)
		if err := nm.api.Publish(nm.ctx, msg, []uint64{n.UserID}, n.FrameUrl); err != nil {
			return fmt.Errorf("error sending notification: %s", err)
		}
		if err := nm.db.RemoveNotification(n.ID); err != nil {
			return fmt.Errorf("error deleting notification: %s", err)
		}
		time.Sleep(DefaultSendCoolDown)
	}

	return nil
}

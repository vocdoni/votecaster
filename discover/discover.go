package discover

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vocdoni/farcaster-poc/mongo"
	"go.vocdoni.io/dvote/log"
)

const (
	FarcasterV2API = "https://client.warpcast.com/v2"
	Throttle       = 5 * time.Second
)

// UserProfile represents the user profile from the Farcaster API v2.
type UserProfile struct {
	Result struct {
		User struct {
			Fid         uint64 `json:"fid"`
			Username    string `json:"username"`
			DisplayName string `json:"displayName"`
			Pfp         struct {
				Url      string `json:"url"`
				Verified bool   `json:"verified"`
			} `json:"pfp"`
			Profile struct {
				Bio struct {
					Text            string   `json:"text"`
					Mentions        []string `json:"mentions"`
					ChannelMentions []string `json:"channelMentions"`
				} `json:"bio"`
				Location struct {
					PlaceId     string `json:"placeId"`
					Description string `json:"description"`
				} `json:"location"`
			} `json:"profile"`
			FollowerCount     int  `json:"followerCount"`
			FollowingCount    int  `json:"followingCount"`
			ActiveOnFcNetwork bool `json:"activeOnFcNetwork"`
			ViewerContext     struct {
				Following            bool `json:"following"`
				FollowedBy           bool `json:"followedBy"`
				CanSendDirectCasts   bool `json:"canSendDirectCasts"`
				HasUploadedInboxKeys bool `json:"hasUploadedInboxKeys"`
			} `json:"viewerContext"`
		} `json:"user"`
		InviterIsReferrer bool          `json:"inviterIsReferrer"`
		CollectionsOwned  []interface{} `json:"collectionsOwned"`
		Extras            struct {
			Fid            uint64 `json:"fid"`
			CustodyAddress string `json:"custodyAddress"`
		} `json:"extras"`
	} `json:"result"`
}

// FarcasterDiscover is a service to discover user profiles from the Farcaster API v2.
type FarcasterDiscover struct {
	db  *mongo.MongoStorage
	cli *http.Client
}

// NewFarcasterDiscover returns a new FarcasterDiscover instance.
// The instance is used to discover user profiles from the Farcaster API v2.
// And update the pending user profiles in the database.
func NewFarcasterDiscover(db *mongo.MongoStorage) *FarcasterDiscover {
	return &FarcasterDiscover{
		db:  db,
		cli: &http.Client{Timeout: 10 * time.Second},
	}
}

// UserProfile returns the user profile from the Farcaster API v2.
func (d *FarcasterDiscover) UserProfile(fid uint64) (*UserProfile, error) {
	var profile *UserProfile
	resp, err := d.cli.Get(fmt.Sprintf("%s/user?fid=%d", FarcasterV2API, fid))
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user profile: %w", err)
	}
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user profile: %w", err)
	}
	return profile, nil
}

// Run starts the discovery process to update user profiles that are pending in the database.
// This is a blocking function that will run indefinitely.
func (d *FarcasterDiscover) Run(ctx context.Context) {
	log.Infow("starting user profile discovery", "throttle", Throttle, "api", FarcasterV2API)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			users, err := d.db.UsersWithPendingProfile()
			if err != nil {
				log.Warnw("failed to get users with pending profile", "error", err)
				continue
			}
			if len(users) == 0 {
				// no pending users, wait a bit more
				time.Sleep(Throttle * 6)
				continue
			}
			log.Infow("discovering pending user profile", "count", len(users))
			for _, fid := range users {
				time.Sleep(Throttle)
				user, err := d.db.User(fid)
				if err != nil {
					log.Warnw("failed to get user", "fid", fid, "error", err)
					continue
				}
				if user.Username != "" {
					// already updated
					log.Warnw("user profile already updated", "fid", user, "username", user.Username)
					continue
				}
				profile, err := d.UserProfile(user.UserID)
				if err != nil {
					log.Warnw("failed to get user profile", "error", err)
					continue
				}
				if profile.Result.User.Fid != user.UserID || profile.Result.User.Username == "" {
					log.Warnw("user profile seems invalid, skipping", "fid", user.UserID)
					continue
				}
				if err := d.db.UpdateUser(&mongo.User{
					UserID:        user.UserID,
					Username:      profile.Result.User.Username,
					CastedVotes:   user.CastedVotes,
					ElectionCount: user.ElectionCount,
					Addresses:     user.Addresses,
				}); err != nil {
					log.Warnw("failed to update user profile", "error", err)
				}
				log.Debugw("updated user profile", "fid", user.UserID, "username", profile.Result.User.Username)
			}
		}
	}
}

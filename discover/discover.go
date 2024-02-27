package discover

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/log"
)

const (
	farcasterV2APIuser          = "https://client.warpcast.com/v2/user?fid=%d"
	farcasterV2APIverifications = "https://client.warpcast.com/v2/verifications?fid=%d&limit=100"
	// https://api.warpcast.com/v2/recent-users?filter=off&limit=100
	Throttle                = 15 * time.Second
	updatedUsersByIteration = 10
	protocolEthereum        = "ethereum"
	userAgent               = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"
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

// VerificationResponse is the response from the Farcaster API v2 for the verifications endpoint.
type VerificationResponse struct {
	Result struct {
		Verifications []struct {
			FID       int    `json:"fid"`
			Address   string `json:"address"`
			Timestamp int64  `json:"timestamp"`
			Version   string `json:"version"`
			Protocol  string `json:"protocol"`
		} `json:"verifications"`
	} `json:"result"`
}

// FarcasterDiscover is a service to discover user profiles from the Farcaster API v2.
type FarcasterDiscover struct {
	db           *mongo.MongoStorage
	cli          *http.Client
	invalid      sync.Map
	farcasterAPI farcasterapi.API
}

// NewFarcasterDiscover returns a new FarcasterDiscover instance.
// The instance is used to discover user profiles from the Farcaster API v2.
// And update the pending user profiles in the database.
func NewFarcasterDiscover(db *mongo.MongoStorage, farcasterAPI farcasterapi.API) *FarcasterDiscover {
	return &FarcasterDiscover{
		db:           db,
		cli:          &http.Client{Timeout: 10 * time.Second},
		farcasterAPI: farcasterAPI,
	}
}

// UserProfile returns the user profile from the Farcaster API v2.
func (d *FarcasterDiscover) UserProfile(fid uint64) (*UserProfile, error) {
	var profile *UserProfile
	// Create a new HTTP request
	req, err := http.NewRequest("GET", fmt.Sprintf(farcasterV2APIuser, fid), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// Set a custom user-agent
	req.Header.Set("User-Agent", userAgent)
	resp, err := d.cli.Do(req)
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

// updateUser updates the user profile in the database. It adds the user if it doesn't exist.
func (d *FarcasterDiscover) updateUser(fid uint64) error {
	if _, ok := d.invalid.Load(fid); ok {
		// already invalid
		return nil
	}
	profile, err := d.UserProfile(fid)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}
	if profile.Result.User.Fid != fid || profile.Result.User.Username == "" || profile.Result.Extras.CustodyAddress == "" {
		log.Warnw("user profile seems invalid, skipping", "fid", fid, "username", profile.Result.User.Username)
		d.invalid.Store(fid, true)
		return nil
	}
	addresses, err := d.Addresses(fid)
	if err != nil {
		return fmt.Errorf("failed to get user addresses: %w", err)
	}
	var castedVotes, electionCount uint64
	user, err := d.db.User(fid)
	if err != nil {
		if !errors.Is(err, mongo.ErrUserUnknown) {
			return fmt.Errorf("failed to get user from database: %w", err)
		}
	} else {
		castedVotes = user.CastedVotes
		electionCount = user.ElectionCount
	}

	signers, err := d.farcasterAPI.SignersFromFID(fid)
	if err != nil {
		return fmt.Errorf("failed to get user signers: %w", err)
	}

	if err := d.db.UpdateUser(&mongo.User{
		UserID:         profile.Result.User.Fid,
		Username:       profile.Result.User.Username,
		CastedVotes:    castedVotes,
		ElectionCount:  electionCount,
		Addresses:      addresses,
		CustodyAddress: profile.Result.Extras.CustodyAddress,
		Signers:        signers,
	}); err != nil {
		log.Warnw("failed to update user profile", "error", err)
	}
	return nil
}

func (d *FarcasterDiscover) Addresses(fid uint64) ([]string, error) {
	var verifications *VerificationResponse
	// Create a new HTTP request
	req, err := http.NewRequest("GET", fmt.Sprintf(farcasterV2APIverifications, fid), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// Set a custom user-agent
	req.Header.Set("User-Agent", userAgent)
	resp, err := d.cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user verifications: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user verifications: %w", err)
	}
	if err := json.Unmarshal(data, &verifications); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user verifications: %w", err)
	}
	var addresses []string
	for _, v := range verifications.Result.Verifications {
		if v.Protocol == protocolEthereum {
			addresses = append(addresses, common.HexToAddress(v.Address).Hex())
		}
	}
	return addresses, nil
}

// Run starts the discovery process to update user profiles that are pending in the database.
// This is a non blocking function that runs in the background.
func (d *FarcasterDiscover) Run(ctx context.Context) {
	log.Infow("starting user profile discovery", "throttle", Throttle)
	go d.runPendingProfiles(ctx)
	go d.runGeneralProfileUpdate(ctx)
}

// runGeneralProfileUpdate starts the discovery process to update user profiles in the database.
func (d *FarcasterDiscover) runGeneralProfileUpdate(ctx context.Context) {
	startID := uint64(0)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			users, err := d.db.UserIDs(startID, updatedUsersByIteration)
			if err != nil {
				log.Warnw("failed to get users with pending profile", "error", err)
				continue
			}
			if len(users) == 0 {
				// no pending users, wait a bit more and reset the startID
				time.Sleep(Throttle * 6)
				startID = 0
				continue
			}
			log.Infow("updating user profiles", "count", len(users), "from", startID, "to", users[len(users)-1])
			for _, fid := range users {
				time.Sleep(Throttle)
				if err := d.updateUser(fid); err != nil {
					log.Warnw("failed to update user profile", "error", err)
				}
				startID = fid
			}
			startID++
		}
	}
}

func (d *FarcasterDiscover) runPendingProfiles(ctx context.Context) {
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
				if err := d.updateUser(fid); err != nil {
					log.Warnw("failed to update user profile", "error", err)
				}
			}
		}
	}
}

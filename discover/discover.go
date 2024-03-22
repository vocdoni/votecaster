package discover

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
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
	farcasterV2APIrecentUsers   = "https://api.warpcast.com/v2/recent-users?filter=off&limit=%d"
	Throttle                    = 1 * time.Second
	updatedUsersByIteration     = 100
	protocolEthereum            = "ethereum"
	userAgent                   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"
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
	if profile.Result.User.Fid != fid || profile.Result.User.Username == "" ||
		profile.Result.Extras.CustodyAddress == "" || profile.Result.User.FollowerCount < 0 {
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
		Followers:      uint64(profile.Result.User.FollowerCount),
		LastUpdated:    time.Now(),
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

// lastRegisteredFID returns the last registered FID from the Farcaster API v2.
func (d *FarcasterDiscover) lastRegisteredFID() (uint64, error) {
	var recentUsers struct {
		Result struct {
			Users []struct {
				Fid uint64 `json:"fid"`
			} `json:"users"`
		} `json:"result"`
	}
	// Create a new HTTP request
	req, err := http.NewRequest("GET", fmt.Sprintf(farcasterV2APIrecentUsers, 1), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	// Set a custom user-agent
	req.Header.Set("User-Agent", userAgent)
	resp, err := d.cli.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get recent users: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read recent users: %w", err)
	}
	if err := json.Unmarshal(data, &recentUsers); err != nil {
		return 0, fmt.Errorf("failed to unmarshal recent users: %w", err)
	}
	if len(recentUsers.Result.Users) == 0 {
		return 0, errors.New("no recent users")
	}
	return recentUsers.Result.Users[0].Fid, nil
}

// Run starts the discovery process to update user profiles that are pending in the database.
// This is a non blocking function that runs in the background.
func (d *FarcasterDiscover) Run(ctx context.Context, indexNewUsers bool) {
	go d.runPendingProfiles(ctx)
	go d.runExistingProfilesUpdate(ctx)
	if indexNewUsers {
		go d.runDiscoverProfilesFromRandomStart(ctx, 10)
	}
}

// runGeneralProfileUpdate starts the discovery process to update user profiles in the database.
func (d *FarcasterDiscover) runExistingProfilesUpdate(ctx context.Context) {
	startID, _ := d.randomFID()
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
				time.Sleep(Throttle * 10)
				startID = 0
				continue
			}
			log.Debugw("updating user profiles", "count", len(users),
				"fromFID", startID, "toFID", users[len(users)-1], "totalKnown", d.db.CountUsers())
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
				time.Sleep(Throttle * 10)
				continue
			}
			for _, fid := range users {
				time.Sleep(Throttle)
				if err := d.updateUser(fid); err != nil {
					log.Warnw("failed to update user profile", "error", err)
				}
			}
		}
	}
}

// randomFID returns a random FID and the last registered FID from the Farcaster API v2.
func (d *FarcasterDiscover) randomFID() (uint64, uint64) {
	lastFID, err := d.lastRegisteredFID()
	if err != nil {
		log.Errorw(err, "failed to get last registered FID, fallback to 370000")
		lastFID = 370000
	}
	return uint64(rand.NewSource(time.Now().UnixNano()).Int63())%lastFID + 1, lastFID
}

// runDiscoverProfilesFromRandomStart initializes the update process from a random FID.
// It allows for running up to N parallel workers to update profiles consecutively.
func (d *FarcasterDiscover) runDiscoverProfilesFromRandomStart(ctx context.Context, workerCount int) {
	startFID, lastFID := d.randomFID()
	fidChan := make(chan uint64, workerCount*2) // Buffer to hold twice the number of workers to keep them busy
	log.Infow("starting user profile discovery", "startFID", startFID, "lastFID", lastFID, "totalUsers", d.db.CountUsers())

	var wg sync.WaitGroup
	wg.Add(workerCount)
	var updateCounter int64
	timer := time.NewTimer(30 * time.Second)

	// Start N worker goroutines
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case fid := <-fidChan:
					_ = d.updateUser(fid)
					time.Sleep(time.Second)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Generate FIDs to update, starting from the random FID and wrapping around if necessary
	go func() {
		startTime := time.Now()
		partialTime := time.Now()
		partialUpdateCounter := 0
		for fid := startFID; ; fid++ {
			if fid > lastFID { // Reset FID to 1 after reaching lastFID
				fid = 1
			}

			select {
			case fidChan <- fid:
				updateCounter++
				partialUpdateCounter++
			case <-timer.C:
				totalCount := float64(updateCounter) / time.Since(startTime).Seconds()
				partialCount := float64(partialUpdateCounter) / time.Since(partialTime).Seconds()
				log.Monitor("discovery indexer (users/second)",
					map[string]any{"partial u/s": partialCount, "total u/s": totalCount, "fid": fid, "totalUsers": d.db.CountUsers()})
				partialTime = time.Now()
				partialUpdateCounter = 0
				timer.Reset(30 * time.Second)
			case <-ctx.Done():
				close(fidChan)
				return
			}
		}
	}()

	wg.Wait() // Wait for all workers to finish
}

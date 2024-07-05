package reputation

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vocdoni/census3/apiclient"
	"github.com/vocdoni/vote-frame/airstack"
	"github.com/vocdoni/vote-frame/alfafrens"
	"github.com/vocdoni/vote-frame/communityhub"
	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/log"
)

// Updater is a struct to update user reputation data in the database
// periodically. It calculates the reputation of each user based on their
// activity and boosters. It gets the activity data from the database and the
// boosters data from the Airstack and the Census3 API.
type Updater struct {
	ctx    context.Context
	cancel context.CancelFunc
	waiter sync.WaitGroup

	db            *mongo.MongoStorage
	fapi          farcasterapi.API
	airstack      *airstack.Airstack
	census3       *apiclient.HTTPclient
	lastUpdate    time.Time
	maxConcurrent int

	alfafrensFollowers  map[uint64]bool
	vocdoniFollowers    map[uint64]bool
	votecasterFollowers map[uint64]bool
	recasters           map[uint64]bool
	followersMtx        sync.Mutex
}

// NewUpdater creates a new Updater instance with the given parameters,
// including the parent context, the database, the Airstack client, the Census3
// client, and the maximum number of concurrent updates.
func NewUpdater(ctx context.Context, db *mongo.MongoStorage, fapi farcasterapi.API,
	as *airstack.Airstack, c3 *apiclient.HTTPclient, maxConcurrent int,
) (*Updater, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}
	if fapi == nil {
		return nil, errors.New("farcaster api is required")
	}
	if as == nil {
		return nil, errors.New("airstack client is required")
	}
	if c3 == nil {
		return nil, errors.New("census3 client is required")
	}
	internalCtx, cancel := context.WithCancel(ctx)
	return &Updater{
		ctx:                 internalCtx,
		cancel:              cancel,
		db:                  db,
		fapi:                fapi,
		airstack:            as,
		census3:             c3,
		lastUpdate:          time.Time{},
		maxConcurrent:       maxConcurrent,
		alfafrensFollowers:  make(map[uint64]bool),
		vocdoniFollowers:    make(map[uint64]bool),
		votecasterFollowers: make(map[uint64]bool),
		recasters:           make(map[uint64]bool),
	}, nil
}

// Start method starts the updater with the given cooldown time between updates.
// It will run until the context is canceled, calling the updateUsers method
// periodically and updating the last update time accordingly.
func (u *Updater) Start(coolDown time.Duration) error {
	u.waiter.Add(1)
	go func() {
		defer u.waiter.Done()

		for {
			select {
			case <-u.ctx.Done():
				return
			default:
				// check if is time to update users
				if time.Since(u.lastUpdate) < coolDown {
					time.Sleep(time.Second * 30)
					continue
				}
				// update internal followers
				if err := u.updateFollowersAndRecasters(); err != nil {
					log.Warnw("error updating internal followers", "error", err)
				}
				// launch update communities
				if err := u.updateCommunities(); err != nil {
					log.Warnw("error updating communities", "error", err)
				}
				// launch update
				if err := u.updateUsers(); err != nil {
					log.Warnw("error updating users", "error", err)
				}
				// update last update time
				u.lastUpdate = time.Now()
			}
		}
	}()
	return nil
}

// Stop method stops the updater by canceling the context and waiting for the
// updater to finish.
func (u *Updater) Stop() {
	u.cancel()
	u.waiter.Wait()
}

// updateFollowersAndRecasters method updates the internal followers of the
// Vocdoni and Votecaster profiles in Farcaster and warpcast users that have
// recasted the Votecaster Launch cast announcement. It fetches the followers
// and recasters data from the Farcaster API and updates the internal followers
// maps accordingly. It returns an error if the followers data cannot be fetched.
func (u *Updater) updateFollowersAndRecasters() error {
	internalCtx, cancel := context.WithTimeout(u.ctx, time.Second*30)
	defer cancel()
	u.followersMtx.Lock()
	defer u.followersMtx.Unlock()
	// update alfafrens followers
	alfafrensFollowers, err := alfafrens.ChannelFids(VotecasterAlphafrensChannelAddress.Bytes())
	if err == nil {
		for _, fid := range alfafrensFollowers {
			u.alfafrensFollowers[fid] = true
		}
	}
	// update vocdoni followers
	vocdoniFollowers, err1 := u.fapi.UserFollowers(internalCtx, VocdoniFarcasterFID)
	if err1 == nil {
		for _, fid := range vocdoniFollowers {
			u.vocdoniFollowers[fid] = true
		}
	}
	// update votecaster followers
	votecasterFollowers, err2 := u.fapi.UserFollowers(internalCtx, VotecasterFarcasterFID)
	if err2 == nil {
		for _, fid := range votecasterFollowers {
			u.votecasterFollowers[fid] = true
		}
	}
	// update recasters
	recasters, err3 := u.fapi.RecastsFIDs(internalCtx, &farcasterapi.APIMessage{
		Author: VocdoniFarcasterFID,
		Hash:   VotecasterAnnouncementCastHash,
	})
	if err3 == nil {
		for _, fid := range recasters {
			u.recasters[fid] = true
		}
	}
	if err != nil || err1 != nil || err2 != nil || err3 != nil {
		return fmt.Errorf("error updating internal followers: %w, %w, %w, %w", err, err1, err2, err3)
	}
	return nil
}

// isFollowerAndRecaster method checks if a given user is a Vocdoni Alfafrens
// Folllower, is a follower of the Vocdoni and Votecaster profiles in
// Farcaster, and if the user has recasted the Votecaster Launch cast
// announcement. It returns four boolean values in if the user is a Folllower
// of Vocdoni Alfafrens, a follower of Vocdoni, a follower of Votecaster, and a
// recaster of the Votecaster Launch cast announcement.
func (u *Updater) isFollowerAndRecaster(userID uint64) (bool, bool, bool, bool) {
	u.followersMtx.Lock()
	defer u.followersMtx.Unlock()
	return u.alfafrensFollowers[userID],
		u.vocdoniFollowers[userID],
		u.votecasterFollowers[userID],
		u.recasters[userID]
}

// updateUsers method iterates over all users in the database and updates their
// reputation data. It uses a concurrent approach to update multiple users at
// the same time, limiting the number of concurrent updates to the maximum
// number of concurrent updates set in the Updater instance. It fetches the
// activity data from the database and the boosters data from the Airstack and
// the Census3 API.
func (u *Updater) updateUsers() error {
	log.Info("updating users reputation")
	ctx, cancel := context.WithCancel(u.ctx)
	defer cancel()
	// limit the number of concurrent updates and create the channel to receive
	// the users, creates also the inner waiter to wait for all updates to
	// finish
	concurrentUpdates := make(chan struct{}, u.maxConcurrent)
	usersChan := make(chan *mongo.User)
	innerWaiter := sync.WaitGroup{}
	// counters for total and updated users
	updates := atomic.Int64{}
	total := atomic.Int64{}
	// listen for users and update them concurrently
	innerWaiter.Add(1)
	go func() {
		defer innerWaiter.Done()
		for user := range usersChan {
			total.Add(1)
			// get a slot in the concurrent updates channel
			concurrentUpdates <- struct{}{}
			go func(user *mongo.User) {
				// release the slot when the update is done
				defer func() {
					<-concurrentUpdates
				}()
				// update user reputation
				if err := u.updateUser(user, true, true); err != nil {
					log.Errorf("error updating user %d: %v", user.UserID, err)
				} else {
					updates.Add(1)
				}
			}(user)
		}
	}()
	// iterate over users and send them to the channel
	if err := u.db.UsersIterator(ctx, usersChan); err != nil {
		return fmt.Errorf("error iterating users: %w", err)
	}
	innerWaiter.Wait()
	log.Infow("users reputation updated", "total", total.Load(), "updated", updates.Load())
	return nil
}

func (u *Updater) updateCommunities() error {
	log.Info("updating communities reputation")
	// limit the number of concurrent updates and create the channel to receive
	// the communities, creates also the inner waiter to wait for all updates to
	// finish
	concurrentUpdates := make(chan struct{}, u.maxConcurrent)
	innerWaiter := sync.WaitGroup{}
	communities, total, err := u.db.ListCommunities(-1, 0)
	if err != nil {
		return fmt.Errorf("error listing communities: %w", err)
	}
	// counters for total and updated communities
	updates := atomic.Int64{}
	// listen for communities and update them concurrently
	innerWaiter.Add(1)
	go func() {
		defer innerWaiter.Done()
		for _, community := range communities {
			// get a slot in the concurrent updates channel
			concurrentUpdates <- struct{}{}
			go func(community *mongo.Community) {
				// release the slot when the update is done
				defer func() {
					<-concurrentUpdates
				}()
				participation, censusSize, err := u.communityPoints(community)
				if err != nil {
					log.Errorf("error getting community %d points: %v", community.ID, err)
					return
				}
				if err := u.db.SetCommunityPoints(community.ID, participation, censusSize); err != nil {
					log.Errorf("error updating community %d reputation: %v", community.ID, err)
					return
				}
				updates.Add(1)
			}(&community)
		}
	}()
	innerWaiter.Wait()
	log.Infow("communities reputation updated", "total", total, "updated", updates.Load())
	return nil
}

// updateUser method updates the reputation data of a given user. It fetches the
// activity data from the database and the boosters data from the Airstack and
// the Census3 API. It then updates the reputation data in the database.
func (u *Updater) updateUser(user *mongo.User, activity, boosters bool) error {
	if u.db == nil {
		return fmt.Errorf("database not set")
	}
	rep, err := u.db.DetailedUserReputation(user.UserID)
	if err != nil {
		// if the user is not found, create a new user with blank data
		if errors.Is(err, mongo.ErrUserUnknown) {
			return u.db.SetDetailedReputationForUser(user.UserID, &mongo.UserReputation{})
		}
		// return the error if it is not a user unknown error
		return err
	}
	// get activiy data if needed
	if activity {
		activityRep, err := u.userActivityReputation(user)
		if err != nil {
			// if there is an error fetching the activity data, log the error and
			// continue updating the no failed activity data
			log.Warnw("error getting user activity reputation", "error", err, "user", user.UserID)
		}
		// update reputation
		rep.FollowersCount = activityRep.FollowersCount
		rep.ElectionsCreated = activityRep.ElectionsCreated
		rep.CastedVotes = activityRep.CastedVotes
		rep.VotesCastedOnCreatedElections = activityRep.VotesCastedOnCreatedElections
		rep.CommunitiesCount = activityRep.CommunitiesCount
	}
	// get boosters data if needed
	if boosters {
		boostersRep, err := u.userBoosters(user)
		// if there is an error fetching the boosters data, log the error and
		// continue updating the no failed boosters data
		if err != nil {
			log.Warnw("error getting some boosters", "error", err, "user", user.UserID)
		}
		// update reputation
		rep.HasVotecasterNFTPass = boostersRep.HasVotecasterNFTPass
		rep.HasVotecasterLaunchNFT = boostersRep.HasVotecasterLaunchNFT
		rep.IsVotecasterAlphafrensFollower = boostersRep.IsVotecasterAlphafrensFollower
		rep.IsVotecasterFarcasterFollower = boostersRep.IsVotecasterFarcasterFollower
		rep.IsVocdoniFarcasterFollower = boostersRep.IsVocdoniFarcasterFollower
		rep.VotecasterAnnouncementRecasted = boostersRep.VotecasterAnnouncementRecasted
		rep.HasKIWI = boostersRep.HasKIWI
		rep.HasDegenDAONFT = boostersRep.HasDegenDAONFT
		rep.Has10kDegenAtLeast = boostersRep.Has10kDegenAtLeast
		rep.HasTokyoDAONFT = boostersRep.HasTokyoDAONFT
		rep.Has5ProxyAtLeast = boostersRep.Has5ProxyAtLeast
		rep.HasNameDegen = boostersRep.HasNameDegen
	}
	// commit reputation
	return u.db.SetDetailedReputationForUser(user.UserID, rep)
}

// userActivityReputation method fetches the activity data of a given user from
// the database. It returns the activity data as an ActivityReputation struct.
// The activity data includes the number of followers, the number of elections
// created, the number of casted votes, the number of votes casted on elections
// created by the user, and the number of communities where the user is an
// admin. It returns an error if the activity data cannot be fetched.
func (u *Updater) userActivityReputation(user *mongo.User) (*ActivityReputation, error) {
	// Fetch the total votes cast on elections created by the user
	totalVotes, err := u.db.TotalVotesForUserElections(user.UserID)
	if err != nil {
		return &ActivityReputation{}, fmt.Errorf("error fetching total votes for user elections: %w", err)
	}
	// Fetch the number of communities where the user is an admin
	communitiesCount, err := u.db.CommunitiesCountForUser(user.UserID)
	if err != nil {
		return &ActivityReputation{}, fmt.Errorf("error fetching communities count for user: %w", err)
	}
	return &ActivityReputation{
		FollowersCount:                user.Followers,
		ElectionsCreated:              user.ElectionCount,
		CastedVotes:                   user.CastedVotes,
		VotesCastedOnCreatedElections: totalVotes,
		CommunitiesCount:              communitiesCount,
	}, nil
}

func (u *Updater) communityPoints(community *mongo.Community) (float64, uint64, error) {
	participation, err := u.db.CommunityParticipationMean(community.ID)
	if err != nil {
		return 0, 0, fmt.Errorf("error fetching community participation mean: %w", err)
	}
	ctx, cancel := context.WithTimeout(u.ctx, time.Second*30)
	defer cancel()
	var censusSize uint64
	switch community.Census.Type {
	case mongo.TypeCommunityCensusChannel:
		users, err := u.fapi.ChannelFIDs(ctx, community.Census.Channel, nil)
		if err != nil {
			return 0, 0, fmt.Errorf("error fetching channel users: %w", err)
		}
		censusSize = uint64(len(users))
	case mongo.TypeCommunityCensusERC20, mongo.TypeCommunityCensusNFT:
		singleUsers := map[common.Address]bool{}
		for _, token := range community.Census.Addresses {
			holders, err := u.airstack.TokenBalances(common.HexToAddress(token.Address), token.Blockchain)
			if err != nil {
				return 0, 0, fmt.Errorf("error fetching token holders: %w", err)
			}
			for _, holder := range holders {
				if _, ok := singleUsers[holder.Address]; !ok {
					singleUsers[holder.Address] = true
				}
			}
		}
		censusSize = uint64(len(singleUsers))
	case mongo.TypeCommunityCensusFollowers:
		fid, err := communityhub.UserRefToFID(community.Census.Channel)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid follower census user reference: %w", err)
		}
		users, err := u.fapi.UserFollowers(ctx, fid)
		if err != nil {
			return 0, 0, fmt.Errorf("error fetching user followers: %w", err)
		}
		censusSize = uint64(len(users))
	default:
		return 0, 0, fmt.Errorf("invalid census type")
	}
	return participation, censusSize, nil
}

// userBoosters method fetches the boosters data of a given user from the
// Airstack and the Census3 API. It returns the boosters data as a Boosters
// struct. The boosters data includes whether the user has the Votecaster NFT
// pass, the Votecaster Launch NFT, the user is subscribed to Votecaster
// Alphafrens channel, the user follows Votecaster and Vocdoni profiles on
// Farcaster, the user has recasted the Votecaster Launch cast announcement,
// the user has KIWI, the user has the DegenDAO NFT, the user has at least 10k
// Degen, the user has Haberdashery NFT, the user has the TokyoDAO NFT, the user
// has a Proxy, the user has at least 5 Proxies, the user has the ProxyStudio
// NFT, and the user has the NameDegen NFT. It returns an error if the boosters
// data cannot be fetched.
func (u *Updater) userBoosters(user *mongo.User) (*Boosters, error) {
	// create new boosters struct and slice for errors
	boosters := &Boosters{}
	var errs []error
	// check if user is votecaster alphafrens follower, is vocdoni or votecaster
	// farcaster follower, and if the user has recasted the votecaster launch
	// cast announcement
	alfafrensFolllower, vocdoniFollower, votecasterFollower, announcementRecaster := u.isFollowerAndRecaster(user.UserID)
	boosters.IsVotecasterAlphafrensFollower = alfafrensFolllower
	boosters.IsVocdoniFarcasterFollower = vocdoniFollower
	boosters.IsVotecasterFarcasterFollower = votecasterFollower
	boosters.VotecasterAnnouncementRecasted = announcementRecaster
	// for every user address check every booster only if it is not already set
	for _, strAddr := range user.Addresses {
		addr := common.HexToAddress(strAddr)
		// check if user has votecaster nft pass
		if !boosters.HasVotecasterNFTPass {
			balance, err := u.airstack.Client.CheckIfHolder(VotecasterNFTPassAddress, VotecasterNFTPassChainShortName, addr)
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting votecaster nft pass balance for %s: %w", addr, err))
			} else {
				boosters.HasVotecasterNFTPass = balance > 0
			}
		}
		// check if user has votecaster launch nft
		if !boosters.HasVotecasterLaunchNFT {
			balance, err := u.airstack.Client.CheckIfHolder(VotecasterLaunchNFTAddress, VotecasterLaunchNFTChainShortName, addr)
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting votecaster launch nft balance for %s: %w", addr, err))
			} else {
				boosters.HasVotecasterLaunchNFT = balance > 0
			}
		}
		// check if user has KIWI
		if !boosters.HasKIWI {
			balance, err := u.census3.TokenHolder(KIWIAddress.Hex(), KIWIChainID, "", strAddr)
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting KIWI balance for %s: %w", addr, err))
			} else if balance != nil {
				boosters.HasKIWI = balance.Cmp(big.NewInt(0)) > 0
			}
		}
		// check if user has DegenDAO NFT
		if !boosters.HasDegenDAONFT {
			balance, err := u.airstack.Client.CheckIfHolder(DegenDAONFTAddress, DegenDAONFTChainShortName, addr)
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting DegenDAO NFT balance for %s: %w", addr, err))
			} else {
				boosters.HasDegenDAONFT = balance > 0
			}
		}
		// check if user has Haberdashery NFT
		if !boosters.HasHaberdasheryNFT {
			balance, err := u.airstack.Client.CheckIfHolder(HaberdasheryNFTAddress, HaberdasheryNFTChainShortName, addr)
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting Haberdashery NFT balance for %s: %w", addr, err))
			} else {
				boosters.HasHaberdasheryNFT = balance > 0
			}
		}
		// check if user has 10k Degen
		if !boosters.Has10kDegenAtLeast {
			balance, err := u.airstack.Client.CheckIfHolder(DegenAddress, DegenChainShortName, addr)
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting 10k Degen balance for %s: %w", addr, err))
			} else {
				boosters.Has10kDegenAtLeast = balance >= 10000
			}
		}
		// check if user has TokyoDAO NFT
		if !boosters.HasTokyoDAONFT {
			balance, err := u.airstack.Client.CheckIfHolder(TokyoDAONFTAddress, TokyoDAONFTChainShortName, addr)
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting TokyoDAO NFT balance for %s: %w", addr, err))
			} else {
				boosters.HasTokyoDAONFT = balance > 0
			}
		}
		// check if user has Proxy and at least 5 Proxies
		if !boosters.Has5ProxyAtLeast {
			balance, err := u.airstack.Client.CheckIfHolder(ProxyAddress, ProxyChainShortName, addr)
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting Proxy balance for %s: %w", addr, err))
			} else {
				boosters.Has5ProxyAtLeast = balance >= 5
			}
		}
		// check if user has ProxyStudio NFT
		if !boosters.HasProxyStudioNFT {
			balance, err := u.airstack.Client.CheckIfHolder(ProxyStudioNFTAddress, ProxyStudioNFTShortName, addr)
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting ProxyStudio NFT balance for %s: %w", addr, err))
			} else {
				boosters.HasProxyStudioNFT = balance > 0
			}
		}
		// check if user has NameDegen
		if !boosters.HasNameDegen {
			balance, err := u.airstack.Client.CheckIfHolder(NameDegenAddress, NameDegenChainShortName, addr)
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting NameDegen balance for %s: %w", addr, err))
			} else {
				boosters.HasNameDegen = balance > 0
			}
		}
	}
	// if there are errors, return the boosters and the errors
	if len(errs) > 0 {
		return boosters, errors.Join(errs...)
	}
	return boosters, nil
}

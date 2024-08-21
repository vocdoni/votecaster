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
	dbmongo "github.com/vocdoni/vote-frame/mongo"
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

	db            *dbmongo.MongoStorage
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
	cachedFollowers     atomic.Bool

	votecasterNFTPassHolders   map[common.Address]*big.Int
	votecasterLaunchNFTHolders map[common.Address]*big.Int
	kiwiHolders                map[common.Address]*big.Int
	degenDAONFTHolders         map[common.Address]*big.Int
	haberdasheryNFTHolders     map[common.Address]*big.Int
	tokyoDAONFTHolders         map[common.Address]*big.Int
	proxyHolders               map[common.Address]*big.Int
	proxyStudioNFTHolders      map[common.Address]*big.Int
	nameDegenHolders           map[common.Address]*big.Int
	farcasterOGNFTHolders      map[common.Address]*big.Int
	moxiePassHolders           map[common.Address]*big.Int
	holdersMtx                 sync.Mutex
	cachedHolders              atomic.Bool
}

// NewUpdater creates a new Updater instance with the given parameters,
// including the parent context, the database, the Airstack client, the Census3
// client, and the maximum number of concurrent updates.
func NewUpdater(ctx context.Context, db *dbmongo.MongoStorage, fapi farcasterapi.API,
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
		ctx:                        internalCtx,
		cancel:                     cancel,
		db:                         db,
		fapi:                       fapi,
		airstack:                   as,
		census3:                    c3,
		lastUpdate:                 time.Time{},
		maxConcurrent:              maxConcurrent,
		alfafrensFollowers:         make(map[uint64]bool),
		vocdoniFollowers:           make(map[uint64]bool),
		votecasterFollowers:        make(map[uint64]bool),
		recasters:                  make(map[uint64]bool),
		votecasterNFTPassHolders:   make(map[common.Address]*big.Int),
		votecasterLaunchNFTHolders: make(map[common.Address]*big.Int),
		kiwiHolders:                make(map[common.Address]*big.Int),
		degenDAONFTHolders:         make(map[common.Address]*big.Int),
		haberdasheryNFTHolders:     make(map[common.Address]*big.Int),
		tokyoDAONFTHolders:         make(map[common.Address]*big.Int),
		proxyHolders:               make(map[common.Address]*big.Int),
		proxyStudioNFTHolders:      make(map[common.Address]*big.Int),
		nameDegenHolders:           make(map[common.Address]*big.Int),
		farcasterOGNFTHolders:      make(map[common.Address]*big.Int),
		moxiePassHolders:           make(map[common.Address]*big.Int),
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
				// fetch internal followers
				if err := u.fetchFollowersAndRecasters(); err != nil {
					log.Warnw("error fetching internal followers", "error", err)
				}
				// fetch holders
				if err := u.fetchHolders(); err != nil {
					log.Warnw("error fetching holders", "error", err)
				}
				// update communities contants (participation mean and census size)
				if err := u.updateCommunitiesContants(); err != nil {
					if mongo.IsDBClosed(err) {
						return
					}
					log.Warnw("error updating communities constants", "error", err)
				}
				// update users constants (activity reputation and boosters)
				if err := u.updateUsersConstants(); err != nil {
					if mongo.IsDBClosed(err) {
						return
					}
					log.Warnw("error updating users constants", "error", err)
				}
				// update total reputations of both, communities and users
				if err := u.updateTotalReputations(); err != nil {
					if mongo.IsDBClosed(err) {
						return
					}
					log.Warnw("error updating total reputations", "error", err)
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
	log.Info("stopping reputation updater")
	u.cancel()
	u.waiter.Wait()
}

// UserReputation method calculates the reputation and points of a user based on
// their activity and boosters. It fetches the user data from the database and
// returns the user reputation. If the commit parameter is true, it updates the
// user reputation in the database.
func (u *Updater) UserReputation(userID uint64, commit bool) (*Reputation, error) {
	user, err := u.db.User(userID)
	if err != nil {
		// return the error if it is not a user unknown error
		return nil, err
	}
	// fetch internal followers
	if !u.cachedFollowers.Load() {
		if err := u.fetchFollowersAndRecasters(); err != nil {
			log.Warnw("error fetching internal followers", "error", err)
		}
	}
	// fetch holders
	if !u.cachedHolders.Load() {
		if err := u.fetchHolders(); err != nil {
			log.Warnw("error fetching holders", "error", err)
		}
	}
	// calculate user reputation
	rep, err := u.userReputation(user)
	if err != nil {
		return nil, fmt.Errorf("error calculating user reputation: %w", err)
	}
	// calculate total points
	if rep == nil {
		return nil, fmt.Errorf("user reputation not found")
	}
	rep.TotalPoints, err = u.userPoints(user.UserID, rep.TotalReputation)
	if err != nil {
		return nil, fmt.Errorf("error calculating user points: %w", err)
	}
	// update user reputation in the database if commit is true
	if commit {
		if err := u.db.SetDetailedReputationForUser(user.UserID, rep); err != nil {
			return nil, fmt.Errorf("error updating user reputation: %w", err)
		}
	}
	return ReputationToAPIResponse(rep), nil
}

// userPoints method calculates the points of a user based on the user
// reputation and the activity as a voter and as a creator of communities. It
// returns the total points of the user. The total points of a user are the sum
// of the points of the communities where the user is the creator and the points
// of the communities where the user is a voter, ponderated by specific
// multipliers by role.
func (u *Updater) userPoints(userID, totalReputation uint64) (uint64, error) {
	points := uint64(0)
	// get community reputation of communities where the user is the creator
	// of the community
	ownerCommunities, _, err := u.db.ListCommunitiesByCreatorFID(userID, -1, 0)
	if err != nil {
		return 0, fmt.Errorf("error fetching owner communities: %w", err)
	}
	for _, community := range ownerCommunities {
		comRep, err := u.db.DetailedCommunityReputation(community.ID)
		if err != nil {
			log.Warnw("error fetching community reputation", "error", err)
			continue
		}
		points += communityTotalPoints(
			community.Census.Type,
			ownerMultiplier,
			comRep.Participation,
			comRep.CensusSize,
			totalReputation)
	}
	// get community reputation of communities where the user is a member
	// of the community
	voterCommunities, err := u.db.CommunitiesByVoter(userID)
	if err != nil {
		return 0, fmt.Errorf("error fetching voter communities: %w", err)
	}
	for _, community := range voterCommunities {
		comRep, err := u.db.DetailedCommunityReputation(community.ID)
		if err != nil {
			log.Warnw("error fetching community reputation", "error", err)
			continue
		}
		points += communityTotalPoints(
			community.Census.Type,
			voterMultiplier,
			comRep.Participation,
			comRep.CensusSize,
			totalReputation)
	}
	return points, nil
}

// communityTotalPoints method calculates the points of a community based on the
// community type, the multiplier, the participation, the census size, and the
// reputation of the owner. The total points are calculated as the yield rate
// multiplied by the participation rate and the census size.
func (u *Updater) communityPoints(communityID string) (uint64, error) {
	// get community to get the user fid of the creator
	community, err := u.db.Community(communityID)
	if err != nil {
		return 0, fmt.Errorf("error fetching community: %w", err)
	}
	if community == nil {
		return 0, fmt.Errorf("community not found")
	}
	rep, err := u.db.DetailedCommunityReputation(communityID)
	if err != nil {
		return 0, fmt.Errorf("error fetching community reputation: %w", err)
	}
	// get reputation of the creator
	userRep, err := u.db.DetailedUserReputation(community.Creator)
	if err != nil {
		return 0, fmt.Errorf("error fetching user reputation: %w", err)
	}
	// calculate the points of the community based on the creator reputation
	return communityTotalPoints(
		community.Census.Type,
		communityMultiplier,
		rep.Participation,
		rep.CensusSize,
		userRep.TotalReputation), nil
}

// fetchFollowersAndRecasters method updates the internal followers of the
// Vocdoni and Votecaster profiles in Farcaster and warpcast users that have
// recasted the Votecaster Launch cast announcement. It fetches the followers
// and recasters data from the Farcaster API and updates the internal followers
// maps accordingly. It returns an error if the followers data cannot be fetched.
func (u *Updater) fetchFollowersAndRecasters() error {
	log.Info("fetching followers and recasters")
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
	u.cachedFollowers.Store(true)
	if err != nil || err1 != nil || err2 != nil || err3 != nil {
		return fmt.Errorf("error updating internal followers: %w, %w, %w, %w", err, err1, err2, err3)
	}
	return nil
}

// fetchHolders method updates the internal holders lists to cache the holders
// of the Votecaster NFT pass, the Votecaster Launch NFT, the KIWI token, the
// DegenDAO NFT, the Haberdashery NFT, the TokyoDAO NFT, the Proxy, the
// ProxyStudio NFT, and the NameDegen NFT. It fetches the holders data from the
// Airstack API and the Census3 API. It returns an error if the holders data
// cannot be fetched.
func (u *Updater) fetchHolders() error {
	log.Info("fetching holders of reputation erc20's and nft's")
	u.holdersMtx.Lock()
	defer u.holdersMtx.Unlock()
	var err error
	var errs []error
	// update Votecaster NFT pass holders
	u.votecasterNFTPassHolders, err = u.tokenHolders(VotecasterNFTPassAddress, VotecasterNFTPassChainID)
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting votecaster nft pass holders: %w", err))
	}
	log.Debugw("votecaster nft pass holders", "holders", len(u.votecasterNFTPassHolders))
	// update Votecaster Launch NFT holders
	u.votecasterLaunchNFTHolders, err = u.tokenHolders(VotecasterLaunchNFTAddress, VotecasterLaunchNFTChainID)
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting votecaster launch nft holders: %w", err))
	}
	log.Debugw("votecaster launch nft holders", "holders", len(u.votecasterLaunchNFTHolders))
	// update KIWI holders
	u.kiwiHolders, err = u.tokenHolders(KIWIAddress, KIWIChainID)
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting kiwi holders: %w", err))
	}
	log.Debugw("kiwi holders", "holders", len(u.kiwiHolders))
	// update DegenDAO NFT holders
	u.degenDAONFTHolders, err = u.tokenHolders(DegenDAONFTAddress, DegenDAONFTChainChainID)
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting degen dao nft holders: %w", err))
	}
	log.Debugw("degen dao nft holders", "holders", len(u.degenDAONFTHolders))
	// Note: Degen holders are not cached because they are too many, instead
	// every time we need to check if a user is a Degen holder we will check
	// the balance of the user

	// update Haberdashery NFT holders
	u.haberdasheryNFTHolders, err = u.tokenHolders(HaberdasheryNFTAddress, HaberdasheryNFTChainChainID)
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting haberdashery nft holders: %w", err))
	}
	log.Debugw("haberdashery nft holders", "holders", len(u.haberdasheryNFTHolders))
	// update TokyoDAO NFT holders
	tokyoDAONFTHolders, err := u.airstack.Client.TokenBalances(
		TokyoDAONFTAddress, TokyoDAONFTChainShortName)
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting TokyoDAO NFT holders: %w", err))
	} else {
		u.tokyoDAONFTHolders = make(map[common.Address]*big.Int)
		for _, holder := range tokyoDAONFTHolders {
			if holder.Balance.Cmp(big.NewInt(0)) > 0 {
				u.tokyoDAONFTHolders[holder.Address] = holder.Balance
			}
		}
	}
	log.Debugw("tokyo dao nft holders", "holders", len(u.tokyoDAONFTHolders))
	// update Proxy
	u.proxyHolders, err = u.tokenHolders(ProxyAddress, ProxyChainID)
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting proxy holders: %w", err))
	}
	log.Debugw("proxy holders", "holders", len(u.proxyHolders))
	// update ProxyStudio NFT holders
	u.proxyStudioNFTHolders, err = u.tokenHolders(ProxyStudioNFTAddress, ProxyStudioNFTChainID)
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting proxy studio nft holders: %w", err))
	}
	log.Debugw("proxy studio nft holders", "holders", len(u.proxyStudioNFTHolders))
	// update NameDegen NFT holders
	u.nameDegenHolders, err = u.tokenHolders(NameDegenAddress, NameDegenChainID)
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting name degen nft holders: %w", err))
	}
	log.Debugw("name degen nft holders", "holders", len(u.nameDegenHolders))
	// update Farcaster OG NFT holders
	u.farcasterOGNFTHolders, err = u.tokenHolders(FarcasterOGNFTAddress, FarcasterOGNFTChainID)
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting farcaster og nft holders: %w", err))
	}
	log.Debugw("farcaster og nft holders", "holders", len(u.farcasterOGNFTHolders))
	// update Moxie Pass NFT holders
	u.moxiePassHolders, err = u.tokenHolders(MoxiePassAddress, MoxiePassChainChainID)
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting moxie pass nft holders: %w", err))
	}
	u.cachedHolders.Store(true)
	if len(errs) > 0 {
		return fmt.Errorf("error updating holders: %v", errs)
	}
	return nil
}

// updateCommunitiesContants method updates the participation mean and the
// census size of all the communities in the database. It fetches the
// communities data from the database and an external source to get the census
// size of each community. It updates the reputation of each community and
// returns an error if the communities data cannot be fetched or the reputation
// cannot be updated.
func (u *Updater) updateCommunitiesContants() error {
	log.Info("updating communities contants")
	// limit the number of concurrent updates and create the channel to receive
	// the communities, creates also the inner waiter to wait for all updates to
	// finish
	concurrentUpdates := make(chan struct{}, u.maxConcurrent)
	innerWaiter := sync.WaitGroup{}
	communities, total, err := u.db.AllCommunities(-1, 0)
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
			updates.Add(1)
			innerWaiter.Add(1)
			go func(community *dbmongo.Community) {
				// release the slot when the update is done
				defer func() {
					<-concurrentUpdates
					innerWaiter.Done()
				}()
				participation, censusSize, err := u.communityConstants(community)
				if err != nil {
					log.Errorf("error getting community %s points: %v", community.ID, err)
					return
				}
				if err := u.db.SetDetailedReputationForCommunity(community.ID, &dbmongo.Reputation{
					Participation: participation,
					CensusSize:    censusSize,
				}); err != nil {
					if !mongo.IsDBClosed(err) {
						log.Errorf("error updating community %s reputation: %v", community.ID, err)
					}
					return
				}
			}(&community)
		}
	}()
	innerWaiter.Wait()
	log.Infow("communities reputation updated", "total", total, "updated", updates.Load())
	return nil
}

// updateUsersConstants method updates the reputation of all the users in the
// database. It fetches the users data from the database and updates the
// reputation of each user based on their activity and boosters. It returns an
// error if the users data cannot be fetched or the reputation cannot be
// updated.
func (u *Updater) updateUsersConstants() error {
	log.Info("updating users contants")
	// limit the number of concurrent updates and create the channel to receive
	// the users, creates also the inner waiter to wait for all updates to
	// finish
	concurrentUpdates := make(chan struct{}, u.maxConcurrent)
	innerWaiter := sync.WaitGroup{}
	// counters for total and updated users
	updates := atomic.Int64{}
	total := atomic.Int64{}
	// listen for users and update them concurrently
	innerWaiter.Add(1)
	go func() {
		defer innerWaiter.Done()
		users, err := u.db.ReputableUsers()
		if err != nil {
			log.Errorf("error fetching reputable users: %v", err)
			return
		}
		total.Store(int64(len(users)))
		for _, user := range users {
			// get a slot in the concurrent updates channel
			concurrentUpdates <- struct{}{}
			innerWaiter.Add(1)
			go func(user *dbmongo.User) {
				// release the slot when the update is done
				defer func() {
					<-concurrentUpdates
					innerWaiter.Done()
				}()
				// update user reputation
				if err := u.updateUserContants(user); err != nil {
					if mongo.IsDBClosed(err) {
						return
					}
					log.Errorf("error updating user %d: %v", user.UserID, err)
				} else {
					updates.Add(1)
				}
			}(user)
		}
	}()
	innerWaiter.Wait()
	close(concurrentUpdates)
	log.Infow("users reputation updated", "total", total.Load(), "updated", updates.Load())
	return nil
}

// updateTotalReputations method updates the total reputation of all the users
// and communities in the database. It fetches the reputations data from the
// database and updates the total reputation of each user and community. It
// returns an error if the reputations data cannot be fetched or the total
// reputation cannot be updated.
func (u *Updater) updateTotalReputations() error {
	log.Info("updating total reputations")
	// limit the number of concurrent updates and create the channel to receive
	// the users, creates also the inner waiter to wait for all updates to
	// finish
	concurrentUpdates := make(chan struct{}, u.maxConcurrent)
	innerWaiter := sync.WaitGroup{}
	// counters for total and updated users
	updates := atomic.Int64{}
	total := atomic.Int64{}
	// listen for users and update them concurrently
	innerWaiter.Add(1)
	go func() {
		defer innerWaiter.Done()
		reputations, err := u.db.Reputations()
		if err != nil {
			log.Errorf("error fetching reputations: %v", err)
			return
		}
		total.Store(int64(len(reputations)))
		for _, reputation := range reputations {
			// get a slot in the concurrent updates channel
			concurrentUpdates <- struct{}{}
			innerWaiter.Add(1)
			go func(rep *dbmongo.Reputation) {
				// release the slot when the update is done
				defer func() {
					<-concurrentUpdates
					innerWaiter.Done()
				}()
				// update total reputation
				if err := u.updateTotalReputation(rep); err != nil {
					log.Errorf("error total reputation (userID: %d, communityID: %s): %v",
						rep.UserID, rep.CommunityID, err)
					return
				}
				updates.Add(1)
			}(reputation)
		}
	}()
	innerWaiter.Wait()
	close(concurrentUpdates)
	log.Infow("users reputation updated", "total", total.Load(), "updated", updates.Load())
	return nil
}

// communityConstants method calculates the participation mean and the census
// size of a given community. It fetches the participation mean from the
// database and the census size from an external source based on the type of
// census of the community. It returns the participation mean and the census
// size of the community and an error if the data cannot be fetched.
func (u *Updater) communityConstants(community *dbmongo.Community) (float64, uint64, error) {
	participation, err := u.db.CommunityParticipationMean(community.ID)
	if err != nil {
		return 0, 0, fmt.Errorf("error fetching community participation mean: %w", err)
	}
	ctx, cancel := context.WithTimeout(u.ctx, time.Minute*2)
	defer cancel()
	var censusSize uint64
	switch community.Census.Type {
	case dbmongo.TypeCommunityCensusChannel:
		users, err := u.fapi.ChannelFIDs(ctx, community.Census.Channel, nil)
		if err != nil {
			return 0, 0, fmt.Errorf("error fetching channel users: %w", err)
		}
		censusSize = uint64(len(users))
	case dbmongo.TypeCommunityCensusERC20, dbmongo.TypeCommunityCensusNFT:
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
	case dbmongo.TypeCommunityCensusFollowers:
		fid, err := communityhub.DecodeUserChannelFID(community.Census.Channel)
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

// updateUserContants method updates the reputation data of a given user. It
// fetches the activity data from the database and the boosters data from the
// Airstack and the Census3 API. It then updates the reputation data in the
// database.
func (u *Updater) updateUserContants(user *dbmongo.User) error {
	rep, err := u.userReputation(user)
	if err != nil {
		return fmt.Errorf("error getting user reputation: %w", err)
	}
	if rep == nil {
		return nil
	}
	// commit reputation
	return u.db.SetDetailedReputationForUser(user.UserID, rep)
}

// updateTotalReputation method updates the total reputation of a given user or
// community. It fetches the reputation data from the database and calculates
// the total reputation based on the activity and the boosters. It then updates
// the total reputation in the database.
func (u *Updater) updateTotalReputation(reputation *dbmongo.Reputation) error {
	if u.db == nil {
		return fmt.Errorf("database not set")
	}
	if reputation.UserID == 0 && reputation.CommunityID == "" {
		return fmt.Errorf("invalid reputation data")
	}
	// check if the reputation is about a user or a community
	if reputation.CommunityID != "" {
		// get the community points based on the type of census of the community
		points, err := u.communityPoints(reputation.CommunityID)
		if err != nil {
			return fmt.Errorf("error calculating community points: %w", err)
		}
		if points != 0 {
			if err := u.db.SetDetailedReputationForCommunity(reputation.CommunityID, &dbmongo.Reputation{
				TotalPoints: points,
			}); err != nil {
				return fmt.Errorf("error updating community total reputation: %w", err)
			}
		}
	} else {
		// get the user points based on the current reputation
		points, err := u.userPoints(reputation.UserID, reputation.TotalReputation)
		if err != nil {
			return fmt.Errorf("error calculating user points: %w", err)
		}
		if points != 0 {
			if err := u.db.SetDetailedReputationForUser(reputation.UserID, &dbmongo.Reputation{
				TotalPoints: points,
			}); err != nil {
				return fmt.Errorf("error updating user total reputation: %w", err)
			}
		}
	}
	return nil
}

// userReputation method calculates the reputation of a given user based on the
// user activity and the boosters. It fetches the user data from the database
// and returns the user reputation. It returns an error if the user data cannot
// be fetched or the reputation cannot be calculated.
func (u *Updater) userReputation(user *dbmongo.User) (*mongo.Reputation, error) {
	if u.db == nil {
		return nil, fmt.Errorf("database not set")
	}
	rep, err := u.db.DetailedUserReputation(user.UserID)
	if err != nil {
		// if the user is not found, create a new user with blank data
		if errors.Is(err, dbmongo.ErrUserUnknown) {
			return nil, u.db.SetDetailedReputationForUser(user.UserID, &dbmongo.Reputation{})
		}
		// return the error if it is not a user unknown error
		return nil, err
	}
	// get activiy data and update the reputation
	activityRep, err := u.userActivityReputation(user)
	if err != nil {
		if mongo.IsDBClosed(err) {
			return nil, err
		}
		// if there is an error fetching the activity data, log the error and
		// continue updating the no failed activity data
		log.Warnw("error getting user activity reputation", "error", err, "user", user.UserID)
	} else {
		rep.FollowersCount = activityRep.FollowersCount
		rep.ElectionsCreatedCount = activityRep.ElectionsCreatedCount
		rep.CastVotesCount = activityRep.CastVotesCount
		rep.ParticipationsCount = activityRep.ParticipationsCount
		rep.CommunitiesCount = activityRep.CommunitiesCount
	}
	// get boosters data and update reputation
	boostersRep := u.userBoosters(user)
	rep.HasVotecasterNFTPass = boostersRep.HasVotecasterNFTPass
	rep.HasVotecasterLaunchNFT = boostersRep.HasVotecasterLaunchNFT
	rep.IsVotecasterAlphafrensFollower = boostersRep.IsVotecasterAlphafrensFollower
	rep.IsVotecasterFarcasterFollower = boostersRep.IsVotecasterFarcasterFollower
	rep.IsVocdoniFarcasterFollower = boostersRep.IsVocdoniFarcasterFollower
	rep.VotecasterAnnouncementRecasted = boostersRep.VotecasterAnnouncementRecasted
	rep.HasKIWI = boostersRep.HasKIWI
	rep.HasDegenDAONFT = boostersRep.HasDegenDAONFT
	rep.HasHaberdasheryNFT = boostersRep.HasHaberdasheryNFT
	rep.Has10kDegenAtLeast = boostersRep.Has10kDegenAtLeast
	rep.HasTokyoDAONFT = boostersRep.HasTokyoDAONFT
	rep.HasProxyStudioNFT = boostersRep.HasProxyStudioNFT
	rep.Has5ProxyAtLeast = boostersRep.Has5ProxyAtLeast
	rep.HasNameDegen = boostersRep.HasNameDegen
	rep.HasFarcasterOGNFT = boostersRep.HasFarcasterOGNFT
	rep.HasMoxiePass = boostersRep.HasMoxiePass
	// calculate total reputation
	rep.TotalReputation = totalReputation(activityRep, boostersRep)
	return rep, nil
}

// userActivityReputation method fetches the activity data of a given user from
// the database. It returns the activity data as an ActivityReputation struct.
// The activity data includes the number of followers, the number of elections
// created, the number of casted votes, the number of votes casted on elections
// created by the user, and the number of communities where the user is an
// admin. It returns an error if the activity data cannot be fetched.
func (u *Updater) userActivityReputation(user *dbmongo.User) (*ActivityReputationCounts, error) {
	// Fetch the total votes cast on elections created by the user
	totalVotes, err := u.db.TotalVotesForUserElections(user.UserID)
	if err != nil {
		return &ActivityReputationCounts{}, fmt.Errorf("error fetching total votes for user elections: %w", err)
	}
	// Fetch the number of communities where the user is an admin
	communitiesCount, err := u.db.CommunitiesCountForUser(user.UserID)
	if err != nil {
		return &ActivityReputationCounts{}, fmt.Errorf("error fetching communities count for user: %w", err)
	}
	return &ActivityReputationCounts{
		FollowersCount:        user.Followers,
		ElectionsCreatedCount: user.ElectionCount,
		CastVotesCount:        user.CastedVotes,
		ParticipationsCount:   totalVotes,
		CommunitiesCount:      communitiesCount,
	}, nil
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
func (u *Updater) userBoosters(user *dbmongo.User) *Boosters {
	// create new boosters struct and slice for errors
	boosters := &Boosters{}
	// check if user is votecaster alphafrens follower, is vocdoni or votecaster
	// farcaster follower, and if the user has recasted the votecaster launch
	// cast announcement
	u.followersMtx.Lock()
	defer u.followersMtx.Unlock()
	boosters.IsVotecasterAlphafrensFollower = u.alfafrensFollowers[user.UserID]
	boosters.IsVocdoniFarcasterFollower = u.vocdoniFollowers[user.UserID]
	boosters.IsVotecasterFarcasterFollower = u.votecasterFollowers[user.UserID]
	boosters.VotecasterAnnouncementRecasted = u.recasters[user.UserID]
	// for every user address check every booster only if it is not already set
	u.holdersMtx.Lock()
	defer u.holdersMtx.Unlock()
	for _, strAddr := range user.Addresses {
		addr := common.HexToAddress(strAddr)
		// check if user has votecaster nft pass
		if !boosters.HasVotecasterNFTPass {
			_, ok := u.votecasterNFTPassHolders[addr]
			boosters.HasVotecasterNFTPass = ok
		}
		// check if user has votecaster launch nft
		if !boosters.HasVotecasterLaunchNFT {
			_, ok := u.votecasterLaunchNFTHolders[addr]
			boosters.HasVotecasterLaunchNFT = ok
		}
		// check if user has KIWI
		if !boosters.HasKIWI {
			_, ok := u.kiwiHolders[addr]
			boosters.HasKIWI = ok
		}
		// check if user has DegenDAO NFT
		if !boosters.HasDegenDAONFT {
			_, ok := u.degenDAONFTHolders[addr]
			boosters.HasDegenDAONFT = ok
		}
		// check if user has Haberdashery NFT
		if !boosters.HasHaberdasheryNFT {
			_, ok := u.haberdasheryNFTHolders[addr]
			boosters.HasHaberdasheryNFT = ok
		}
		// check if user has 10k Degen
		if !boosters.Has10kDegenAtLeast {
			balance, err := u.census3.TokenHolder(DegenAddress.Hex(), DegenChainID, "", addr.Hex())
			if err != nil {
				log.Warnw("error checking if user has 10k degen", "error", err, "user", user.UserID)
			}
			if balance != nil {
				boosters.Has10kDegenAtLeast = balance.Cmp(big.NewInt(10000)) >= 0
			}
		}
		// check if user has TokyoDAO NFT
		if !boosters.HasTokyoDAONFT {
			_, ok := u.tokyoDAONFTHolders[addr]
			boosters.HasTokyoDAONFT = ok
		}
		// check if user has Proxy and at least 5 Proxies
		if !boosters.Has5ProxyAtLeast {
			if balance, ok := u.proxyHolders[addr]; ok {
				boosters.Has5ProxyAtLeast = balance.Cmp(big.NewInt(5)) >= 0
			}
		}
		// check if user has ProxyStudio NFT
		if !boosters.HasProxyStudioNFT {
			_, ok := u.proxyStudioNFTHolders[addr]
			boosters.HasProxyStudioNFT = ok
		}
		// check if user has NameDegen
		if !boosters.HasNameDegen {
			_, ok := u.nameDegenHolders[addr]
			boosters.HasNameDegen = ok
		}
		// check if user has Farcaster OG NFT
		if !boosters.HasFarcasterOGNFT {
			_, ok := u.farcasterOGNFTHolders[addr]
			boosters.HasFarcasterOGNFT = ok
		}
		// check if user has Moxie Pass
		if !boosters.HasMoxiePass {
			_, ok := u.moxiePassHolders[addr]
			boosters.HasMoxiePass = ok
		}
	}
	return boosters
}

// tokenHolders method fetches the holders of a given token from the Census3 API.
// It returns the token holders as a map of addresses and balances. It returns an
// error if the token holders data cannot be fetched.
func (u *Updater) tokenHolders(address common.Address, chainID uint64) (map[common.Address]*big.Int, error) {
	tokenInfo, err := u.census3.Token(address.Hex(), chainID, "")
	if err != nil {
		return nil, fmt.Errorf("error getting token info: %w", err)
	}
	return u.census3.AllHoldersByStrategy(tokenInfo.DefaultStrategy)
}

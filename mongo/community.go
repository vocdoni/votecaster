package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

func (ms *MongoStorage) AddCommunity(community *Community) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	return ms.addCommunity(community)
}

func (ms *MongoStorage) UpdateCommunity(community *Community) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	return ms.updateCommunity(community)
}

func (ms *MongoStorage) Community(id string) (*Community, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	return ms.community(id)
}

// ListCommunities returns the list of enabled communities.
func (ms *MongoStorage) ListCommunities(limit, offset int64) ([]Community, int64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	// filter by enabled and communities

	communities := []Community{}
	opts := options.Find().SetSort(bson.D{{Key: "featured", Value: -1}, {Key: "_id", Value: -1}})
	total, err := paginatedObjects(ms.communities, bson.M{"disabled": false}, opts, limit, offset, &communities)
	if err != nil {
		return nil, 0, err
	}
	return communities, total, nil
}

// AllCommunities returns the list of all communities.
func (ms *MongoStorage) AllCommunities(limit, offset int64) ([]Community, int64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	// filter by enabled and communities

	communities := []Community{}
	opts := options.Find().SetSort(bson.D{{Key: "featured", Value: -1}, {Key: "_id", Value: -1}})
	total, err := paginatedObjects(ms.communities, nil, opts, limit, offset, &communities)
	if err != nil {
		return nil, 0, err
	}
	return communities, total, nil
}

// ListFeaturedCommunities returns the list of featured communities.
func (ms *MongoStorage) ListFeaturedCommunities(limit, offset int64) ([]Community, int64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	communities := []Community{}
	total, err := paginatedObjects(ms.communities, bson.M{"featured": true}, nil, limit, offset, &communities)
	if err != nil {
		return nil, 0, err
	}
	return communities, total, nil
}

// ListCommunitiesByAdminFID returns the list of communities where the user is an
// admin by FID provided.
func (ms *MongoStorage) ListCommunitiesByAdminFID(fid uint64, limit, offset int64) ([]Community, int64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	communities := []Community{}
	total, err := paginatedObjects(ms.communities, bson.M{"owners": fid}, nil, limit, offset, &communities)
	if err != nil {
		log.Debug("error listing communities by admin FID: ", err)
		return nil, 0, err
	}
	return communities, total, nil
}

// ListCommunitiesByAdminFID returns the list of communities where the user is
// the creator by FID provided.
func (ms *MongoStorage) ListCommunitiesByCreatorFID(fid uint64, limit, offset int64) ([]Community, int64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	communities := []Community{}
	total, err := paginatedObjects(ms.communities, bson.M{"creator": fid}, nil, limit, offset, &communities)
	if err != nil {
		log.Debug("error listing communities by admin FID: ", err)
		return nil, 0, err
	}
	return communities, total, nil
}

// ListCommunitiesByAdminUsername returns the list of communities where the
// user is an admin by username provided. It queries about the user FID first.
func (ms *MongoStorage) ListCommunitiesByAdminUsername(username string, limit, offset int64) ([]Community, int64, error) {
	user, err := ms.UserByUsername(username)
	if err != nil {
		return nil, 0, err
	}
	return ms.ListCommunitiesByAdminFID(user.UserID, limit, offset)
}

func (ms *MongoStorage) LastCommunityID(prefix string) (string, error) {
	// Find documents
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Define the filter to match IDs with the specified prefix
	filter := bson.M{"_id": bson.M{"$regex": fmt.Sprintf("^%s", prefix)}}
	projection := bson.M{"_id": 1}
	options := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}}).SetProjection(projection)

	var community *Community
	if err := ms.communities.FindOne(ctx, filter, options).Decode(&community); err != nil {
		if strings.Contains(err.Error(), "no documents in result") {
			return "", ErrNoResults
		}
		return "", err
	}
	return community.ID, nil
}

// DelCommunity removes the community with the specified ID from the database.
// If an error occurs, it returns the error.
func (ms *MongoStorage) DelCommunity(communityID string) error {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.communities.DeleteOne(ctx, bson.M{"_id": communityID})
	return err
}

// CommunityParticipationMean returns the mean of the participation of the every
// community poll with the given ID. It returns an error if something goes wrong
// with the database.
func (ms *MongoStorage) CommunityParticipationMean(communityID string) (float64, error) {
	elections, err := ms.ElectionsByCommunity(communityID)
	if err != nil {
		return 0, err
	}
	if len(elections) == 0 {
		return 0, nil
	}
	// participation = Σ (sum of votes / sum of voters * 100)
	var totalParticipation float64
	for _, election := range elections {
		// prevent to divide by zero
		if election.FarcasterUserCount == 0 {
			continue
		}
		totalParticipation += float64(election.CastedVotes) / float64(election.FarcasterUserCount) * 100
	}
	// mean participation = total participation / number of elections
	return totalParticipation / float64(len(elections)), nil
}

// CommunitiesByVoter returns the list of communities with elections where the
// user is voter. It returns an error if something goes wrong with the database.
func (ms *MongoStorage) CommunitiesByVoter(userID uint64) ([]Community, error) {
	// get elections with a defined community object and then check if the user
	// id provided is voter for any of those elections, then returns the
	// communities of the elections where the user is voter
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	// get elections with a defined community object
	communityElections, err := ms.electionsWithCommunity()
	if err != nil {
		return nil, err
	}
	// iterate over the elections getting the voters to check if the user is one
	// of them, if so, get the community of the election and add it to the list
	communities := []Community{}
	alreadyIncluded := map[string]bool{}
	for _, election := range communityElections {
		// if the community is already included, skip it to avoid duplicates
		// and unnecessary queries
		if _, ok := alreadyIncluded[election.Community.ID]; ok {
			continue
		}
		// get the voters of the election and check if the user is one of them
		voters, err := ms.votersOfElection(types.HexBytes(election.ElectionID))
		if err != nil {
			if err == ErrElectionUnknown {
				continue
			}
			return nil, err
		}
		// if the user is not a voter, skip the election
		if !ms.isUserVoter(voters, userID) {
			continue
		}
		// get the community of the election and add it to the list
		community, err := ms.community(election.Community.ID)
		if err != nil {
			return nil, err
		}
		alreadyIncluded[election.Community.ID] = true
		communities = append(communities, *community)
	}
	return communities, nil
}

// addCommunity method adds a new community to the database. It returns an error
// if something the census type is invalid or something goes wrong with the
// database.
func (ms *MongoStorage) addCommunity(community *Community) error {
	switch community.Census.Type {
	case TypeCommunityCensusChannel, TypeCommunityCensusERC20, TypeCommunityCensusNFT, TypeCommunityCensusFollowers:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := ms.communities.InsertOne(ctx, community)
		return err
	default:
		return fmt.Errorf("invalid census type")
	}
}

// updateCommunity method updates the community in the database. It returns an
// error if something goes wrong with the database.
func (ms *MongoStorage) updateCommunity(community *Community) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateDoc, err := dynamicUpdateDocument(community, []string{"notifications", "disabled"})
	if err != nil {
		return fmt.Errorf("failed to create update document: %w", err)
	}
	opts := options.Update().SetUpsert(true) // Ensures the document is created if it does not exist
	_, err = ms.communities.UpdateOne(ctx, bson.M{"_id": community.ID}, updateDoc, opts)
	if err != nil {
		return fmt.Errorf("cannot update election: %w", err)
	}
	return nil
}

// community method returns the community with the given id. If something goes
// wrong, it returns an error. If the community does not exist, it returns nil.
func (ms *MongoStorage) community(id string) (*Community, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var community Community
	err := ms.communities.FindOne(ctx, bson.M{"_id": id}).Decode(&community)
	if err != nil {
		if strings.Contains(err.Error(), "no documents in result") {
			return nil, nil
		}
		return nil, err
	}
	return &community, nil
}

// IsCommunityAdmin checks if the user is an admin of the given community by ID.
func (ms *MongoStorage) IsCommunityAdmin(userID uint64, communityID string) bool {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Query to check if the userID is one of the admins in the given communityID
	count, err := ms.communities.CountDocuments(ctx, bson.M{
		"_id":    communityID,
		"owners": bson.M{"$in": []uint64{userID}}, // Correct usage of querying inside an array
	})
	if err != nil {
		log.Warn("Error querying community admins: ", err)
		return false
	}
	return count > 0
}

// IsCommunityDisabled checks if the community with the given ID is disabled.
func (ms *MongoStorage) IsCommunityDisabled(communityID string) bool {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var community Community
	if err := ms.communities.FindOne(ctx, bson.M{"_id": communityID}).Decode(&community); err != nil {
		log.Errorf("error getting community %s: %v", communityID, err)
		return false
	}
	return community.Disabled
}

// SetCommunityStatus sets the disabled status of the community with the given ID.
func (ms *MongoStorage) SetCommunityStatus(communityID string, disabled bool) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.communities.UpdateOne(ctx, bson.M{"_id": communityID}, bson.M{"$set": bson.M{"disabled": disabled}})
	return err
}

// CommunityAllowNotifications checks if the community with the given ID has
// notifications enabled.
func (ms *MongoStorage) CommunityAllowNotifications(communityID string) bool {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var community Community
	if err := ms.communities.FindOne(ctx, bson.M{"_id": communityID}).Decode(&community); err != nil {
		log.Errorf("error getting community %s: %v", communityID, err)
		return false
	}
	return community.Notifications
}

// SetCommunityNotifications sets the disabled status of the community with the given ID.
func (ms *MongoStorage) SetCommunityNotifications(communityID string, enabled bool) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.communities.UpdateOne(ctx, bson.M{"_id": communityID}, bson.M{"$set": bson.M{"notifications": enabled}})
	return err
}

// SetCommunityCensusStrategy sets the census strategy of the community with the given ID.
func (ms *MongoStorage) SetCommunityCensusStrategy(communityID string, strategyID uint64) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.communities.UpdateOne(ctx, bson.M{"_id": communityID}, bson.M{"$set": bson.M{"census.strategy": strategyID}})
	return err
}

// CommunityCensusStrategy returns the census strategy of the community with the
// given ID.
func (ms *MongoStorage) CommunityCensusStrategy(communityID string) (uint64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var community Community
	if err := ms.communities.FindOne(ctx, bson.M{"_id": communityID}).Decode(&community); err != nil {
		log.Errorf("error getting community %s: %v", communityID, err)
		return 0, err
	}
	if community.Census.Strategy == 0 {
		return 0, ErrNoResults
	}
	return community.Census.Strategy, nil
}

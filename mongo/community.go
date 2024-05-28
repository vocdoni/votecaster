package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.vocdoni.io/dvote/log"
)

func (ms *MongoStorage) AddCommunity(id uint64, name, imageUrl, groupChatUrl string,
	census CommunityCensus, channels []string, creator uint64, admins []uint64,
	notifications, disabled bool,
) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	community := Community{
		ID:            id,
		Name:          name,
		Channels:      channels,
		Census:        census,
		ImageURL:      imageUrl,
		GroupChatURL:  groupChatUrl,
		Creator:       creator,
		Admins:        admins,
		Notifications: notifications,
		Disabled:      disabled,
	}
	return ms.addCommunity(&community)
}

func (ms *MongoStorage) UpdateCommunity(community *Community) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	return ms.updateCommunity(community)
}

func (ms *MongoStorage) Community(id uint64) (*Community, error) {
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
	total, err := paginatedObjects(ms.communities, bson.M{"disabled": false}, nil, limit, offset, &communities)
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

// NextCommunityID returns the next community ID which will be assigned to a new
// community. It returns the last community ID + 1. If there are no communities
// in the database, it returns 0. If something goes wrong, it returns an error.
func (ms *MongoStorage) NextCommunityID() (uint64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.FindOne().SetSort(bson.M{"_id": -1})
	// find the last community ID
	var community Community
	err := ms.communities.FindOne(ctx, bson.M{}, opts).Decode(&community)
	if err != nil && !strings.Contains(err.Error(), "no documents in result") {
		// if there is an error and it is not because there are no documents
		// in the result, return the error and 0 (invalid ID)
		return 0, err
	}
	return community.ID + 1, nil
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

// DelCommunity removes the community with the specified ID from the database.
// If an error occurs, it returns the error.
func (ms *MongoStorage) DelCommunity(communityID uint64) error {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.communities.DeleteOne(ctx, bson.M{"_id": communityID})
	return err
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
func (ms *MongoStorage) community(id uint64) (*Community, error) {
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
func (ms *MongoStorage) IsCommunityAdmin(userID, communityID uint64) bool {
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
func (ms *MongoStorage) IsCommunityDisabled(communityID uint64) bool {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var community Community
	if err := ms.communities.FindOne(ctx, bson.M{"_id": communityID}).Decode(&community); err != nil {
		log.Errorf("error getting community %d: %v", communityID, err)
		return false
	}
	return community.Disabled
}

// SetCommunityStatus sets the disabled status of the community with the given ID.
func (ms *MongoStorage) SetCommunityStatus(communityID uint64, disabled bool) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.communities.UpdateOne(ctx, bson.M{"_id": communityID}, bson.M{"$set": bson.M{"disabled": disabled}})
	return err
}

// CommunityAllowNotifications checks if the community with the given ID has
// notifications enabled.
func (ms *MongoStorage) CommunityAllowNotifications(communityID uint64) bool {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var community Community
	if err := ms.communities.FindOne(ctx, bson.M{"_id": communityID}).Decode(&community); err != nil {
		log.Errorf("error getting community %d: %v", communityID, err)
		return false
	}
	return community.Notifications
}

// SetCommunityNotifications sets the disabled status of the community with the given ID.
func (ms *MongoStorage) SetCommunityNotifications(communityID uint64, enabled bool) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.communities.UpdateOne(ctx, bson.M{"_id": communityID}, bson.M{"$set": bson.M{"notifications": enabled}})
	return err
}

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
	census CommunityCensus, channels []string, admins []uint64, notifications, disabled bool,
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
		Admins:        admins,
		Notifications: notifications,
		Disabled:      disabled,
	}
	return ms.addCommunity(&community)
}

func (ms *MongoStorage) Community(id uint64) (*Community, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	return ms.community(id)
}

func (ms *MongoStorage) ListCommunities() ([]Community, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := ms.communitites.Find(ctx, bson.M{"disabled": false})
	if err != nil {
		if strings.Contains(err.Error(), "no documents in result") {
			return nil, nil
		}
		return nil, err
	}
	var communities []Community
	ctx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	for cursor.Next(ctx) {
		var community Community
		if err := cursor.Decode(&community); err != nil {
			log.Warn(err)
			continue
		}
		communities = append(communities, community)
	}
	return communities, nil
}

// ListCommunitiesByAdminFID returns the list of communities where the user is an
// admin by FID provided.
func (ms *MongoStorage) ListCommunitiesByAdminFID(fid uint64) ([]Community, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := ms.communitites.Find(ctx, bson.M{"owners": fid})
	if err != nil {
		if strings.Contains(err.Error(), "no documents in result") {
			return nil, nil
		}
		return nil, err
	}
	var communities []Community
	ctx, cancel2 := context.WithTimeout(ctx, 10*time.Second)
	defer cancel2()
	for cursor.Next(ctx) {
		var community Community
		if err := cursor.Decode(&community); err != nil {
			log.Warn(err)
			continue
		}
		communities = append(communities, community)
	}
	return communities, nil
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
	err := ms.communitites.FindOne(ctx, bson.M{}, opts).Decode(&community)
	if err != nil && !strings.Contains(err.Error(), "no documents in result") {
		// if there is an error and it is not because there are no documents
		// in the result, return the error and 0 (invalid ID)
		return 0, err
	}
	return community.ID + 1, nil
}

// ListCommunitiesByAdminUsername returns the list of communities where the
// user is an admin by username provided. It queries about the user FID first.
func (ms *MongoStorage) ListCommunitiesByAdminUsername(username string) ([]Community, error) {
	user, err := ms.UserByUsername(username)
	if err != nil {
		return nil, err
	}
	return ms.ListCommunitiesByAdminFID(user.UserID)
}

// DelCommunity removes the community with the specified ID from the database.
// If an error occurs, it returns the error.
func (ms *MongoStorage) DelCommunity(communityID uint64) error {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.communitites.DeleteOne(ctx, bson.M{"_id": communityID})
	return err
}

// addCommunity method adds a new community to the database. It returns an error
// if something the census type is invalid or something goes wrong with the
// database.
func (ms *MongoStorage) addCommunity(community *Community) error {
	switch community.Census.Type {
	case TypeCommunityCensusChannel, TypeCommunityCensusERC20, TypeCommunityCensusNFT:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := ms.communitites.InsertOne(ctx, community)
		return err
	default:
		return fmt.Errorf("invalid census type")
	}
}

// community method returns the community with the given id. If something goes
// wrong, it returns an error. If the community does not exist, it returns nil.
func (ms *MongoStorage) community(id uint64) (*Community, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var community Community
	err := ms.communitites.FindOne(ctx, bson.M{"_id": id}).Decode(&community)
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
	count, err := ms.communitites.CountDocuments(ctx, bson.M{
		"_id":    communityID,
		"owners": bson.M{"$in": []uint64{userID}}, // Correct usage of querying inside an array
	})
	if err != nil {
		log.Warn("Error querying community admins: ", err)
		return false
	}
	return count > 0
}

// SetCommunityStatus sets the disabled status of the community with the given ID.
func (ms *MongoStorage) SetCommunityStatus(communityID uint64, disabled bool) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.communitites.UpdateOne(ctx, bson.M{"_id": communityID}, bson.M{"$set": bson.M{"disabled": disabled}})
	return err
}

// SetCommunityNotifications sets the disabled status of the community with the given ID.
func (ms *MongoStorage) SetCommunityNotifications(communityID uint64, enabled bool) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.communitites.UpdateOne(ctx, bson.M{"_id": communityID}, bson.M{"$set": bson.M{"notifications": enabled}})
	return err
}

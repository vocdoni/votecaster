package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// UserAccessProfile retrieves the access profile for a given user ID.
// Returns ErrUserUnknown if the user is not found.
func (ms *MongoStorage) UserAccessProfile(userID uint64) (*UserAccessProfile, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var profile UserAccessProfile
	err := ms.userAccessProfiles.FindOne(ctx, bson.M{"_id": userID}).Decode(&profile)
	if err != nil {
		return nil, ErrUserUnknown
	}

	return &profile, nil
}

// SetReputationForUser updates the reputation for a given user ID.
func (ms *MongoStorage) SetReputationForUser(userID uint64, reputation uint32) error {
	return ms.updateUserAccessProfile(userID, bson.M{"$set": bson.M{"reputation": reputation}})
}

// SetAccessLevelForUser updates the access level for a given user ID.
func (ms *MongoStorage) SetAccessLevelForUser(userID uint64, accessLevel uint32) error {
	return ms.updateUserAccessProfile(userID, bson.M{"$set": bson.M{"accessLevel": accessLevel}})
}

// SetNotificationsAcceptedForUser updates the notifications accepted status for a given user ID.
func (ms *MongoStorage) SetNotificationsAcceptedForUser(userID uint64, accepted bool) error {
	return ms.updateUserAccessProfile(userID, bson.M{"$set": bson.M{"notificationsAccepted": accepted}})
}

// SetNotificationsRequestedForUser updates the notifications requested status for a given user ID.
func (ms *MongoStorage) SetNotificationsRequestedForUser(userID uint64, requested bool) error {
	return ms.updateUserAccessProfile(userID, bson.M{"$set": bson.M{"notificationsRequested": requested}})
}

// SetWhiteListedForUser updates the white listed status for a given user ID.
func (ms *MongoStorage) SetWhiteListedForUser(userID uint64, whiteListed bool) error {
	return ms.updateUserAccessProfile(userID, bson.M{"$set": bson.M{"whiteListed": whiteListed}})
}

// updateUserAccessProfile is a helper function to update fields in the UserAccessProfile document.
func (ms *MongoStorage) updateUserAccessProfile(userID uint64, update bson.M) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ms.userAccessProfiles.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

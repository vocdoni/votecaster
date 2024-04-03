package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
// It now performs an upsert, creating the document if it doesn't already exist.
func (ms *MongoStorage) updateUserAccessProfile(userID uint64, update bson.M) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Update().SetUpsert(true)
	_, err := ms.userAccessProfiles.UpdateOne(ctx, bson.M{"_id": userID}, update, opts)
	return err
}

// AddNotificationMutedUser adds a user ID to the owner user's list of muted notifications users.
func (ms *MongoStorage) AddNotificationMutedUser(ownerUserID, mutedUserID uint64) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{"$addToSet": bson.M{"notificationsMutedUsers": mutedUserID}}
	_, err := ms.userAccessProfiles.UpdateOne(ctx, bson.M{"_id": ownerUserID}, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("error adding muted user to notifications: %w", err)
	}

	return nil
}

// IsUserNotificationMuted checks if a user's notifications are muted by the owner user.
func (ms *MongoStorage) IsUserNotificationMuted(ownerUserID, mutedUserID uint64) (bool, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var profile UserAccessProfile
	err := ms.userAccessProfiles.FindOne(ctx, bson.M{"_id": ownerUserID}).Decode(&profile)
	if err != nil {
		return false, ErrUserUnknown
	}

	for _, userID := range profile.NotificationsMutedUsers {
		if userID == mutedUserID {
			return true, nil
		}
	}

	return false, nil
}

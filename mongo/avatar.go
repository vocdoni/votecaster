package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Avatar returns an avatar image with the given avatarID.
func (ms *MongoStorage) Avatar(avatarID string) (*Avatar, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	return ms.avatar(avatarID)
}

// SetAvatar sets the avatar image data for the given avatarID. If the
// avatar does not exist, it will be created with the given data, otherwise it
// will be updated.
func (ms *MongoStorage) SetAvatar(avatarID string, data []byte, userID, communityID uint64) error {
	avatar := &Avatar{
		ID:          avatarID,
		Data:        data,
		CreatedAt:   time.Now(),
		UserID:      userID,
		CommunityID: communityID,
	}
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	return ms.setAvatar(avatar)
}

// RemoveAvatar removes the avatar image data for the given avatarID.
func (ms *MongoStorage) RemoveAvatar(avatarID string) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.avatars.DeleteOne(ctx, bson.M{"_id": avatarID})
	return err
}

func (ms *MongoStorage) avatar(avatarID string) (*Avatar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := ms.avatars.FindOne(ctx, bson.M{"_id": avatarID})
	var avatar Avatar
	if err := result.Decode(&avatar); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrAvatarUnknown
		}
		return nil, fmt.Errorf("error retrieving avatar %s: %w", avatarID, err)
	}
	return &avatar, nil
}

func (ms *MongoStorage) setAvatar(avatar *Avatar) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.ReplaceOptions{}
	opts.Upsert = new(bool)
	*opts.Upsert = true
	_, err := ms.avatars.ReplaceOne(ctx, bson.M{"_id": avatar.ID}, avatar, &opts)
	if err != nil {
		return fmt.Errorf("cannot update object: %w", err)
	}
	return err
}

package mongo

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.vocdoni.io/dvote/log"
)

func (ms *MongoStorage) AddCommunity(id uint64, name, imageUrl string,
	census CommunityCensus, channels []string, admins []uint64, notifications bool,
) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	community := Community{
		ID:            id,
		Name:          name,
		Channels:      channels,
		Census:        census,
		ImageURL:      imageUrl,
		Admins:        admins,
		Notifications: notifications,
	}
	log.Infow("added new community", "id", id, "name", name, "admins", admins)
	return ms.addCommunity(&community)
}

func (ms *MongoStorage) GetCommunity(id uint64) (*Community, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	return ms.getCommunity(id)
}

func (ms *MongoStorage) ListCommunities() ([]Community, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := ms.communitites.Find(ctx, bson.M{})
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

func (ms *MongoStorage) addCommunity(community *Community) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.communitites.InsertOne(ctx, community)
	return err
}

func (ms *MongoStorage) getCommunity(id uint64) (*Community, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var community Community
	err := ms.communitites.FindOne(ctx, bson.M{"_id": id}).Decode(&community)
	if err != nil {
		return nil, err
	}
	return &community, nil
}

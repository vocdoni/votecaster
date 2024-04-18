package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func (ms *MongoStorage) Metadata(key string) (any, error) {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var metadata any
	err := ms.metadata.FindOne(ctx, bson.M{"key": key}).Decode(&metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func (ms *MongoStorage) SetMetadata(key string, value any) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.metadata.UpdateOne(ctx, bson.M{"key": key}, bson.M{"$set": bson.M{"value": value}})
	if err != nil {
		return err
	}
	return nil
}

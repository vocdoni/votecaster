package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func (ms *MongoStorage) AddNotifications(nType NotificationType, electionID string, userID uint64, username, frameURL string) error {
	notification := Notification{
		Type:       nType,
		UserID:     userID,
		ElectionID: electionID,
		Username:   username,
		FrameUrl:   frameURL,
	}
	return ms.addNotification(&notification)
}

func (ms *MongoStorage) addNotification(notification *Notification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.notifications.InsertOne(ctx, notification)
	return err
}

func (ms *MongoStorage) LastNotifications() ([]Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := ms.notifications.Find(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var notifications []Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (ms *MongoStorage) RemoveNotification(notificationID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.notifications.DeleteOne(ctx, bson.M{"_id": notificationID})
	return err
}

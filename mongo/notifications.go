package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (ms *MongoStorage) AddNotifications(nType NotificationType, electionID string,
	userID, authorID uint64, username, authorUsername, frameURL string,
) (int64, error) {
	notification := Notification{
		Type:           nType,
		ElectionID:     electionID,
		UserID:         userID,
		Username:       username,
		AuthorID:       authorID,
		AuthorUsername: authorUsername,
		FrameUrl:       frameURL,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := ms.notifications.InsertOne(ctx, notification)
	if err != nil {
		return 0, err
	}
	return res.InsertedID.(int64), nil
}

func (ms *MongoStorage) LastNotifications(maxResults int) ([]Notification, error) {
	// Setting options for sorting by _id in ascending order, limiting the
	// results, and projecting only the _id field
	opts := options.Find().SetSort(bson.M{"_id": -1}).SetLimit(int64(maxResults))
	// Creating a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Executing the find operation with the specified filter and options
	cursor, err := ms.notifications.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	// Iterating through the cursor to decode the results
	notifications := []Notification{}
	for cursor.Next(ctx) {
		notification := Notification{}
		if err := cursor.Decode(&notification); err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (ms *MongoStorage) RemoveNotification(notificationID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.notifications.DeleteOne(ctx, bson.M{"_id": notificationID})
	return err
}

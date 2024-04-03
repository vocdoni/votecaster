package mongo

import (
	"context"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func (ms *MongoStorage) AddNotifications(nType NotificationType, electionID string,
	userID, authorID uint64, username, authorUsername, frameURL string, deadline time.Time,
) (int64, error) {
	// create random id for the notification
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)
	randomID := rnd.Int63()
	// create notification
	notification := Notification{
		ID:             randomID,
		Type:           nType,
		ElectionID:     electionID,
		UserID:         userID,
		Username:       username,
		AuthorID:       authorID,
		AuthorUsername: authorUsername,
		FrameUrl:       frameURL,
		Deadline:       deadline,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := ms.notifications.InsertOne(ctx, notification)
	if err != nil {
		return 0, err
	}
	return res.InsertedID.(int64), nil
}

// SetNotificationDeadline sets the deadline for the notification with the
// specified ID. If an error occurs, it returns the error.
func (ms *MongoStorage) SetNotificationDeadline(notificationID int64, deadline time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.notifications.UpdateOne(ctx, bson.M{"_id": notificationID},
		bson.M{"$set": bson.M{"deadline": deadline}})
	return err
}

// LastNotifications returns the registered notifications in the database ordered
// by the _id field in descending order and limited to the specified number of
// results. If an error occurs, it returns the error.
func (ms *MongoStorage) LastNotifications(maxResults int) ([]Notification, error) {
	// Creating a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Executing the find operation with the specified filter and options
	cursor, err := ms.notifications.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	// Iterating through the cursor to decode the results
	notifications := []Notification{}
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return notifications, nil
}

// RemoveNotification removes the notification with the specified ID from the
// database. If an error occurs, it returns the error.
func (ms *MongoStorage) RemoveNotification(notificationID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.notifications.DeleteOne(ctx, bson.M{"_id": notificationID})
	return err
}

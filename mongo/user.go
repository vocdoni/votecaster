package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.vocdoni.io/dvote/log"
)

// UserIDs returns a list of user IDs starting from the given ID and limited to the specified amount.
func (ms *MongoStorage) UserIDs(startId uint64, maxResults int) ([]uint64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	// Setting options for sorting by _id in ascending order, limiting the results, and projecting only the _id field
	opts := options.Find().SetSort(bson.M{"_id": 1}).SetLimit(int64(maxResults)).SetProjection(bson.M{"_id": 1})

	// Creating a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Adjusting the filter to find users with _id greater than or equal to startId
	filter := bson.M{"_id": bson.M{"$gte": startId}}

	// Executing the find operation with the specified filter and options
	cur, err := ms.users.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var ids []uint64
	for cur.Next(ctx) {
		var user struct {
			ID uint64 `bson:"_id"`
		}
		if err := cur.Decode(&user); err != nil {
			log.Warn(err)
			continue
		}
		ids = append(ids, user.ID)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

// CountUsers returns the total number of users in the database.
func (ms *MongoStorage) CountUsers() uint64 {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	// Creating a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Counting the documents in the users collection
	count, err := ms.users.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Warnw("cannot count users", "err", err)
		return 0
	}

	return uint64(count)
}

// AddUser adds a new user to the database. If the user already exists, it returns an error.
func (ms *MongoStorage) AddUser(
	userFID uint64,
	usernanme string,
	displayname string,
	addresses []string,
	signers []string,
	custodyAddr string,
	elections uint64,
) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	user := User{
		UserID:        userFID,
		Username:      usernanme,
		Displayname:   displayname,
		Addresses:     addresses,
		Signers:       signers,
		ElectionCount: elections,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.users.InsertOne(ctx, user)
	return err
}

func (ms *MongoStorage) User(userFID uint64) (*User, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	user, err := ms.getUserData(userFID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ms *MongoStorage) UpdateUser(udata *User) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	return ms.updateUser(udata)
}

func (ms *MongoStorage) UserExists(userFID uint64) bool {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	_, err := ms.getUserData(userFID)
	return err == nil
}

func (ms *MongoStorage) DelUser(userFID uint64) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.users.DeleteOne(ctx, bson.M{"_id": userFID})
	return err
}

// UsersWithPendingProfile returns the list of users that have not set their username yet.
// This call is limited to 32 users.
func (ms *MongoStorage) UsersWithPendingProfile() ([]uint64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	limit := int64(32)
	opts := options.FindOptions{Limit: &limit}
	opts.SetProjection(bson.M{"_id": true})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.users.Find(ctx, bson.M{"username": ""}, &opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	var users []uint64
	for cur.Next(ctx) {
		user := User{}
		err := cur.Decode(&user)
		if err != nil {
			log.Warn(err)
		}
		users = append(users, user.UserID)
	}

	return users, nil
}

// UserByAddress returns the user that has the given address. If the user is not found, it returns an error.
func (ms *MongoStorage) UserByAddress(address string) (*User, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var userByAddress User
	if err := ms.users.FindOne(ctx, bson.M{
		"addresses": bson.M{"$in": []string{address}},
	}).Decode(&userByAddress); err != nil {
		return nil, ErrUserUnknown
	}
	return &userByAddress, nil
}

// UserBySigner returns the user that has the given signer. If the user is not found, it returns an error.
func (ms *MongoStorage) UserBySigner(signer string) (*User, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	// Example query to find a user by a signer
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var userBySigner User
	if err := ms.users.FindOne(ctx, bson.M{
		"signers": bson.M{"$in": []string{signer}},
	}).Decode(&userBySigner); err != nil {
		return nil, ErrUserUnknown
	}
	return &userBySigner, nil
}

// UserByUsername returns the user that has the given username. If the user is
// not found, it returns an error.
func (ms *MongoStorage) UserByUsername(username string) (*User, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var userByUsername User
	if err := ms.users.FindOne(ctx, bson.M{"username": username}).Decode(&userByUsername); err != nil {
		return nil, ErrUserUnknown
	}
	return &userByUsername, nil
}

func (ms *MongoStorage) getUserData(userID uint64) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := ms.users.FindOne(ctx, bson.M{"_id": userID})
	var user User
	if err := result.Decode(&user); err != nil {
		return nil, ErrUserUnknown
	}
	return &user, nil
}

// updateUser makes a conditional update on the user, updating only non-zero fields
func (ms *MongoStorage) updateUser(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateDoc, err := dynamicUpdateDocument(user, nil)
	if err != nil {
		return fmt.Errorf("failed to create update document: %w", err)
	}
	log.Debugw("update user", "updateDoc", updateDoc)
	opts := options.Update().SetUpsert(true) // Ensures the document is created if it does not exist
	_, err = ms.users.UpdateOne(ctx, bson.M{"_id": user.UserID}, updateDoc, opts)
	if err != nil {
		return fmt.Errorf("cannot update user: %w", err)
	}
	return nil
}

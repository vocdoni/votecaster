package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.vocdoni.io/dvote/log"
)

func (ms *MongoStorage) Users() (*Users, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	opts := options.FindOptions{}
	opts.SetProjection(bson.M{"_id": true})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.users.Find(ctx, bson.M{}, &opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	var users Users
	for cur.Next(ctx) {
		user := User{}
		err := cur.Decode(&user)
		if err != nil {
			log.Warn(err)
		}
		users.Users = append(users.Users, user.UserID)
	}

	return &users, nil
}

// AddUser adds a new user to the database. If the user already exists, it returns an error.
func (ms *MongoStorage) AddUser(userFID uint64, usernanme string, addresses []string, elections uint64) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	user := User{
		UserID:        userFID,
		Username:      usernanme,
		Addresses:     addresses,
		ElectionCount: elections,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.users.InsertOne(ctx, user)
	log.Infow("added new user", "userID", userFID, "username", usernanme)
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
	log.Debugw("update user",
		"userID", udata.UserID,
		"username", udata.Username,
		"electionCount", udata.ElectionCount,
		"castedVotes", udata.CastedVotes,
	)
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

// updateUser makes a upsert on the user
func (ms *MongoStorage) updateUser(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.ReplaceOptions{}
	opts.Upsert = new(bool)
	*opts.Upsert = true
	_, err := ms.users.ReplaceOne(ctx, bson.M{"_id": user.UserID}, user, &opts)
	if err != nil {
		return fmt.Errorf("cannot update object: %w", err)
	}
	return nil
}

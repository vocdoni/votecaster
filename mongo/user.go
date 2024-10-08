package mongo

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/vocdoni/vote-frame/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

// UsersIterator iterates over available users and sends them to the provided
// channel.
func (ms *MongoStorage) ReputableUsers() ([]*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"castedVotes": bson.M{"$gt": 0}},
			{"electionCount": bson.M{"$gt": 0}},
		},
	}
	// Executing the find operation with the specified filter and options
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()
	cur, err := ms.users.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var users []*User
	for cur.Next(ctx) {
		user := &User{}
		if err := cur.Decode(user); err != nil {
			log.Warn(err)
			continue
		}
		users = append(users, user)
	}
	return users, nil
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

	user, err := ms.userData(userFID)
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
	_, err := ms.userData(userFID)
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

// UserByAddress returns the user that has the given address (case insensitive). If the user is not found, it returns an error.
// Warning, this is expensive and should be used with caution.
func (ms *MongoStorage) UserByAddressCaseInsensitive(address string) (*User, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var userByAddress User
	if err := ms.users.FindOne(ctx, bson.M{
		"addresses": bson.M{
			"$regex":   "^" + regexp.QuoteMeta(address) + "$",
			"$options": "i",
		},
	}).Decode(&userByAddress); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserUnknown
		}
		return nil, err
	}
	return &userByAddress, nil
}

// UserByAddress returns the user that has the given address. If the user is not found, it returns an error.
func (ms *MongoStorage) UserByAddress(address string) (*User, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var userByAddress User
	if err := ms.users.FindOne(ctx, bson.M{
		"addresses": bson.M{"$in": []string{helpers.NormalizeAddressString(address)}},
	}).Decode(&userByAddress); err != nil {
		return nil, ErrUserUnknown
	}
	return &userByAddress, nil
}

// UserByAddressBulk returns a map of users that have the given addresses. If the user is not found, it returns an error.
func (ms *MongoStorage) UserByAddressBulk(addresses []string) (map[string]*User, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	// MongoDB query to find all users whose addresses match any in the list
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query to find users with any of the provided addresses
	filter := bson.M{"addresses": bson.M{"$in": helpers.NormalizeAddressStringSlice(addresses)}}
	cur, err := ms.users.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	// Mapping addresses to users
	userMap := make(map[string]*User)
	for cur.Next(ctx) {
		var user User
		if err := cur.Decode(&user); err != nil {
			return nil, err
		}
		for _, addr := range user.Addresses {
			userMap[addr] = &user
		}
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return userMap, nil
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
	return ms.userDataByUsername(username)
}

// userData retrieves the user data based on the user ID (FID).
func (ms *MongoStorage) userData(userID uint64) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := ms.users.FindOne(ctx, bson.M{"_id": userID})
	var user User
	if err := result.Decode(&user); err != nil {
		return nil, ErrUserUnknown
	}
	return &user, nil
}

// userDataByUsername retrieves the user data based on the username.
func (ms *MongoStorage) userDataByUsername(username string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := ms.users.FindOne(ctx, bson.M{"username": username})
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

	opts := options.Update().SetUpsert(true) // Ensures the document is created if it does not exist
	_, err = ms.users.UpdateOne(ctx, bson.M{"_id": user.UserID}, updateDoc, opts)
	if err != nil {
		return fmt.Errorf("cannot update user: %w", err)
	}
	return nil
}

func (ms *MongoStorage) NormalizeUserAddresses() error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	// Context with a reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Cursor to iterate over all users
	cur, err := ms.users.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("error fetching users: %w", err)
	}
	defer cur.Close(ctx)

	// Prepare bulk update operations
	var bulkOps []mongo.WriteModel

	for cur.Next(ctx) {
		var user User
		if err := cur.Decode(&user); err != nil {
			return fmt.Errorf("error decoding user: %w", err)
		}

		// Normalize addresses
		normalizedAddresses := make([]string, len(user.Addresses))
		for i, addr := range user.Addresses {
			normalizedAddresses[i] = helpers.NormalizeAddressString(addr)
		}

		// Check if an update is necessary
		if !equalStringSlices(normalizedAddresses, user.Addresses) {
			// Create update operation if normalization changed the addresses
			update := mongo.NewUpdateOneModel()
			update.SetFilter(bson.M{"_id": user.UserID})
			update.SetUpdate(bson.M{"$set": bson.M{"addresses": normalizedAddresses}})
			bulkOps = append(bulkOps, update)
		}
	}

	if err := cur.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	// Perform bulk update if there are any updates
	if len(bulkOps) > 0 {
		bulkOpts := options.BulkWrite().SetOrdered(true)
		_, err := ms.users.BulkWrite(ctx, bulkOps, bulkOpts)
		if err != nil {
			return fmt.Errorf("error performing bulk update: %w", err)
		}
		log.Infof("Normalized %d user addresses in the database", len(bulkOps))
	} else {
		log.Info("No user addresses needed normalization")
	}

	return nil
}

// Helper function to compare two slices of strings
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

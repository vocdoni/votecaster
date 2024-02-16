package mongo

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

// MongoStorage uses an external MongoDB service for stoting the user data and election details.
type MongoStorage struct {
	users     *mongo.Collection
	elections *mongo.Collection
	keysLock  sync.RWMutex
	election  funcGetElection
}

type Options struct {
	MongoURL string
	Database string
}

// funcGetElection is a function that returns an election by its ID.
type funcGetElection = func(electionID types.HexBytes) (*api.Election, error)

// AddElectionCallback adds a callback function to get the election details by its ID.
func (ms *MongoStorage) AddElectionCallback(f funcGetElection) {
	ms.election = f
}

func New(url, database string) (*MongoStorage, error) {
	var err error
	ms := &MongoStorage{}
	if url == "" {
		return nil, fmt.Errorf("mongo URL is not defined")
	}
	if database == "" {
		return nil, fmt.Errorf("mongo database is not defined")
	}
	log.Infof("connecting to mongodb %s@%s", url, database)
	opts := options.Client()
	opts.ApplyURI(url)
	opts.SetMaxConnecting(200)
	timeout := time.Second * 10
	opts.ConnectTimeout = &timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	defer cancel()
	if err != nil {
		return nil, err
	}
	// Shutdown database connection when SIGTERM received
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Warnf("received SIGTERM, disconnecting mongo database")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		err := client.Disconnect(ctx)
		if err != nil {
			log.Warn(err)
		}
		cancel()
	}()

	ctx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("cannot connect to mongodb: %w", err)
	}

	ms.users = client.Database(database).Collection("users")
	ms.elections = client.Database(database).Collection("elections")

	// If reset flag is enabled, Reset drops the database documents and recreates indexes
	// else, just createIndexes
	if reset := os.Getenv("VOCDONI_MONGO_RESET_DB"); reset != "" {
		err := ms.Reset()
		if err != nil {
			return nil, err
		}
	} else {
		err := ms.createIndexes()
		if err != nil {
			return nil, err
		}
	}

	return ms, nil
}

func (ms *MongoStorage) createIndexes() error {
	ctx, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel3()
	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: "username", Value: "text"},
		},
	}
	_, err := ms.users.Indexes().CreateOne(ctx, index)
	if err != nil {
		return err
	}
	return nil
}

func (ms *MongoStorage) Reset() error {
	log.Infof("resetting database")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ms.users.Drop(ctx); err != nil {
		return err
	}
	if err := ms.elections.Drop(ctx); err != nil {
		return err
	}
	if err := ms.createIndexes(); err != nil {
		return err
	}
	return nil
}

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

func (ms *MongoStorage) AddElection(electionID types.HexBytes, userFID uint64) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	election := Election{
		UserID:      userFID,
		ElectionID:  electionID.String(),
		CreatedTime: time.Now(),
	}
	log.Infow("added new election", "electionID", electionID.String(), "userID", userFID)
	return ms.addElection(&election)
}

func (ms *MongoStorage) addElection(election *Election) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ms.elections.InsertOne(ctx, election)
	return err
}

func (ms *MongoStorage) Election(electionID types.HexBytes) (*Election, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	election, err := ms.getElection(electionID)
	if err != nil {
		return nil, err
	}
	return election, nil
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

func (ms *MongoStorage) getElection(electionID types.HexBytes) (*Election, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := ms.elections.FindOne(ctx, bson.M{"_id": electionID.String()})
	var election Election
	if err := result.Decode(&election); err != nil {
		log.Warn(err)
		return nil, ErrElectionUnknown
	}
	return &election, nil
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

// updateElection makes a upsert on the election
func (ms *MongoStorage) updateElection(election *Election) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.ReplaceOptions{}
	opts.Upsert = new(bool)
	*opts.Upsert = true
	_, err := ms.elections.ReplaceOne(ctx, bson.M{"_id": election.ElectionID}, election, &opts)
	if err != nil {
		return fmt.Errorf("cannot update object: %w", err)
	}
	return nil
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

func (ms *MongoStorage) IncreaseVoteCount(userFID uint64, electionID types.HexBytes) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	log.Debugw("increase vote count", "userID", userFID, "electionID", electionID.String())

	user, err := ms.getUserData(userFID)
	if err != nil {
		return err
	}
	user.CastedVotes++

	election, err := ms.getElection(electionID)
	if err != nil {
		if errors.Is(err, ErrElectionUnknown) {
			log.Warnw("creating fallback election", "electionID", electionID.String(), "userFID", userFID)
			election = &Election{
				UserID:      userFID,
				CastedVotes: 0,
				ElectionID:  electionID.String(),
				CreatedTime: time.Now(),
			}
			if err := ms.addElection(election); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	election.CastedVotes++
	election.LastVoteTime = time.Now()

	if err := ms.updateUser(user); err != nil {
		return err
	}
	return ms.updateElection(election)
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

// UsersByElectionNumber returns the list of users ordered by the number of elections they have created.
func (ms *MongoStorage) UsersByElectionNumber() ([]UserRanking, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	limit := int64(32)
	opts := options.FindOptions{Limit: &limit}
	opts.SetSort(bson.M{"electionCount": -1})
	opts.SetProjection(bson.M{"_id": true, "username": true, "electionCount": true})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.users.Find(ctx, bson.M{}, &opts)
	if err != nil {
		return nil, err
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	var ranking []UserRanking
	for cur.Next(ctx2) {
		user := User{}
		err := cur.Decode(&user)
		if err != nil {
			log.Warn(err)
		}
		ranking = append(ranking, UserRanking{
			FID:      user.UserID,
			Username: user.Username,
			Count:    user.ElectionCount,
		})
	}
	return ranking, nil
}

// UsersByVoteNumber returns the list of users ordered by the number of votes they have casted.
func (ms *MongoStorage) UsersByVoteNumber() ([]UserRanking, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	limit := int64(32)
	opts := options.FindOptions{Limit: &limit}
	opts.SetSort(bson.M{"castedVotes": -1})
	opts.SetProjection(bson.M{"_id": true, "username": true, "castedVotes": true})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.users.Find(ctx, bson.M{}, &opts)
	if err != nil {
		return nil, err
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	var ranking []UserRanking
	for cur.Next(ctx2) {
		user := User{}
		err := cur.Decode(&user)
		if err != nil {
			log.Warn(err)
		}
		ranking = append(ranking, UserRanking{
			FID:      user.UserID,
			Username: user.Username,
			Count:    user.CastedVotes,
		})
	}
	return ranking, nil
}

// ElectionsByVoteNumber returns the list elections ordered by the number of votes casted.
func (ms *MongoStorage) ElectionsByVoteNumber() ([]ElectionRanking, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	limit := int64(12)
	opts := options.FindOptions{Limit: &limit}
	opts.SetSort(bson.M{"castedVotes": -1})
	opts.SetProjection(bson.M{"_id": true, "castedVotes": true, "userId": true})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.elections.Find(ctx, bson.M{}, &opts)
	if err != nil {
		return nil, err
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	var ranking []ElectionRanking
	for cur.Next(ctx2) {
		election := Election{}
		err := cur.Decode(&election)
		if err != nil {
			log.Warn(err)
		}
		if election.CastedVotes == 0 {
			continue
		}
		eID, err := hex.DecodeString(election.ElectionID)
		if err != nil {
			log.Warn(err)
			continue
		}

		title := ""
		electionInfo, err := ms.election(eID)
		if err != nil {
			log.Warn(err)
		} else {
			if electionInfo != nil && electionInfo.Metadata != nil {
				title = electionInfo.Metadata.Title["default"]
			}
		}

		username := ""
		user, err := ms.getUserData(election.UserID)
		if err != nil {
			log.Warn(err)
		} else {
			username = user.Username
		}

		ranking = append(ranking, ElectionRanking{
			ElectionID:        election.ElectionID,
			VoteCount:         election.CastedVotes,
			CreatedByFID:      election.UserID,
			Title:             title,
			CreatedByUsername: username,
		})

	}
	return ranking, nil
}

func (ms *MongoStorage) Import(data []byte) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()
	var collection UserCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return err
	}
	for _, u := range collection.Users {
		if err := ms.updateUser(&u); err != nil {
			log.Warnf("cannot upsert %d", u.UserID)
		}
	}
	return nil
}

func (ms *MongoStorage) String() string {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := ms.users.Find(ctx, bson.D{{}})
	if err != nil {
		log.Warn(err)
		return "{}"
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	var users UserCollection
	for cur.Next(ctx2) {
		var user User
		err := cur.Decode(&user)
		if err != nil {
			log.Warn(err)
		}
		users.Users = append(users.Users, user)
	}

	ctx3, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel3()
	cur, err = ms.elections.Find(ctx3, bson.D{{}})
	if err != nil {
		log.Warn(err)
		return "{}"
	}

	ctx4, cancel4 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel4()
	var elections ElectionCollection
	for cur.Next(ctx4) {
		var election Election
		err := cur.Decode(&election)
		if err != nil {
			log.Warn(err)
		}
		elections.Elections = append(elections.Elections, election)
	}

	data, err := json.MarshalIndent(struct {
		UserCollection
		ElectionCollection
	}{users, elections},
		"", " ")
	if err != nil {
		log.Warn(err)
	}
	return string(data)
}

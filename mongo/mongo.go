package mongo

import (
	"context"
	"encoding/json"
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
	results   *mongo.Collection
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
	ms.results = client.Database(database).Collection("results")

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

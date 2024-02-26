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
	voters    *mongo.Collection
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
	ms.voters = client.Database(database).Collection("voters")

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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Index model for the 'addresses' field
	addressesIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "addresses", Value: 1}}, // 1 for ascending order
		Options: nil,
	}

	// Index model for the 'signers' field
	signersIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "signers", Value: 1}}, // 1 for ascending order
		Options: nil,
	}

	// Create both indexes
	_, err := ms.users.Indexes().CreateMany(ctx, []mongo.IndexModel{addressesIndexModel, signersIndexModel})
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

	ctx5, cancel5 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel5()
	cur, err = ms.results.Find(ctx5, bson.D{{}})
	if err != nil {
		log.Warn(err)
		return "{}"
	}

	ctx6, cancel6 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel6()
	var results ResultsCollection
	for cur.Next(ctx6) {
		var result Results
		err := cur.Decode(&result)
		if err != nil {
			log.Warn(err)
		}
		results.Results = append(results.Results, result)
	}

	data, err := json.Marshal(&Collection{users, elections, results})
	if err != nil {
		log.Warn(err)
	}
	return string(data)
}

// Import imports a JSON dataset produced by String() into the database.
func (ms *MongoStorage) Import(jsonData []byte) error {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	log.Infof("importing database")
	var collection Collection
	err := json.Unmarshal(jsonData, &collection)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	// Upsert Users
	log.Infow("importing users", "count", len(collection.Users))
	for _, user := range collection.Users {
		filter := bson.M{"_id": user.UserID}
		update := bson.M{"$set": user}
		opts := options.Update().SetUpsert(true)
		_, err := ms.users.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Warnw("Error upserting user", "err", err, "user", user.UserID)
		}
	}

	// Upsert Elections
	log.Infow("importing elections", "count", len(collection.Elections))
	for _, election := range collection.Elections {
		filter := bson.M{"_id": election.ElectionID}
		update := bson.M{"$set": election}
		opts := options.Update().SetUpsert(true)
		_, err := ms.elections.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Warnw("Error upserting election", "err", err, "election", election.ElectionID)
		}
	}

	// Upsert Results
	log.Infow("importing results", "count", len(collection.Results))
	for _, result := range collection.Results {
		filter := bson.M{"_id": result.ElectionID}
		update := bson.M{"$set": result}
		opts := options.Update().SetUpsert(true)
		_, err := ms.results.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Warnw("Error upserting result", "err", err, "election", result.ElectionID)
		}
	}

	log.Infof("imported database!")
	return nil
}

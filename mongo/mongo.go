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

const (
	authenticationExpirationNoActivitySeconds = 15 * 24 * 60 * 60 // 15 days
)

// MongoStorage uses an external MongoDB service for stoting the user data and election details.
type MongoStorage struct {
	client   *mongo.Client
	election funcGetElection
	keysLock sync.RWMutex

	users              *mongo.Collection
	elections          *mongo.Collection
	census             *mongo.Collection
	results            *mongo.Collection
	voters             *mongo.Collection
	authentications    *mongo.Collection
	notifications      *mongo.Collection
	userAccessProfiles *mongo.Collection
	communitites       *mongo.Collection
	metadata           *mongo.Collection
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
	client, err := mongo.Connect(ctx, opts)
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

	ms.client = client
	ms.users = client.Database(database).Collection("users")
	ms.elections = client.Database(database).Collection("elections")
	ms.census = client.Database(database).Collection("census")
	ms.results = client.Database(database).Collection("results")
	ms.voters = client.Database(database).Collection("voters")
	ms.authentications = client.Database(database).Collection("authentications")
	ms.notifications = client.Database(database).Collection("notifications")
	ms.userAccessProfiles = client.Database(database).Collection("userAccessProfiles")
	ms.communitites = client.Database(database).Collection("communities")
	ms.metadata = client.Database(database).Collection("metadata")

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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	// Create an index model for the 'castedVotes' field on users (ranking)
	userCastedVotesIndexModel := mongo.IndexModel{
		Keys:    bson.M{"castedVotes": -1}, // -1 for descending order
		Options: options.Index().SetName("castedVotesIndex"),
	}

	_, err = ms.users.Indexes().CreateOne(ctx, userCastedVotesIndexModel)
	if err != nil {
		return err
	}

	// Create index for authentication collection
	authIndexModel := mongo.IndexModel{
		Keys: bson.M{"authTokens": 1},
	}
	if _, err := ms.authentications.Indexes().CreateOne(ctx, authIndexModel); err != nil {
		return err
	}

	// Create the TTL index for the 'createdAt' field in the authentications collection.
	// With this index, the auth entries will be automatically deleted after N days of no activity.
	ttlIndexModel := mongo.IndexModel{
		Keys:    bson.M{"updatedAt": 1}, // Index on the updatedAt field
		Options: options.Index().SetExpireAfterSeconds(authenticationExpirationNoActivitySeconds),
	}

	if _, err := ms.authentications.Indexes().CreateOne(ctx, ttlIndexModel); err != nil {
		return err
	}

	// Create index for Census Root
	rootIndexModel := mongo.IndexModel{
		Keys:    bson.M{"root": 1}, // 1 for ascending order
		Options: options.Index().SetUnique(false),
	}

	if _, err := ms.census.Indexes().CreateOne(ctx, rootIndexModel); err != nil {
		return fmt.Errorf("failed to create index on root field: %w", err)
	}

	// Create index for Census ElectionID
	electionIDIndexModel := mongo.IndexModel{
		Keys:    bson.M{"electionId": 1}, // 1 for ascending order
		Options: options.Index().SetUnique(false),
	}

	if _, err := ms.census.Indexes().CreateOne(ctx, electionIDIndexModel); err != nil {
		return fmt.Errorf("failed to create index on electionId field: %w", err)
	}

	// Create index for election creation time (ranking)
	electionCreationIndexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "createdTime", Value: -1}}, // -1 for descending order
	}

	_, err = ms.elections.Indexes().CreateOne(ctx, electionCreationIndexModel)
	if err != nil {
		return fmt.Errorf("failed to create index on createdTime: %w", err)
	}

	// Create an index model for the 'castedVotes' field on election (ranking)
	electionCastedVotesIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "castedVotes", Value: -1}}, // -1 for descending order
		Options: options.Index().SetName("castedVotesIndex"),
	}

	_, err = ms.elections.Indexes().CreateOne(ctx, electionCastedVotesIndexModel)
	if err != nil {
		return fmt.Errorf("failed to create index on castedVotes for elections: %w", err)
	}

	// Create an index model for the 'key' field on metadata
	metadataIndex := mongo.IndexModel{
		Keys:    bson.M{"key": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err = ms.metadata.Indexes().CreateOne(ctx, metadataIndex)
	if err != nil {
		return fmt.Errorf("failed to create index on metadata key: %w", err)
	}

	// Create an index for the 'owners' field on communities
	ownersIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "owners", Value: 1}}, // 1 for ascending order
		Options: nil,
	}
	if _, err := ms.communitites.Indexes().CreateOne(ctx, ownersIndexModel); err != nil {
		return fmt.Errorf("failed to create index on owners for communities: %w", err)
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
	const contextTimeout = 20 * time.Second
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()
	cur, err := ms.users.Find(ctx, bson.D{{}})
	if err != nil {
		log.Warn(err)
		return "{}"
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), contextTimeout)
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

	ctx3, cancel3 := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel3()
	cur, err = ms.elections.Find(ctx3, bson.D{{}})
	if err != nil {
		log.Warn(err)
		return "{}"
	}

	ctx4, cancel4 := context.WithTimeout(context.Background(), contextTimeout)
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

	ctx5, cancel5 := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel5()
	cur, err = ms.results.Find(ctx5, bson.D{{}})
	if err != nil {
		log.Warn(err)
		return "{}"
	}

	ctx6, cancel6 := context.WithTimeout(context.Background(), contextTimeout)
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

	ctx7, cancel7 := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel7()
	var votersOfElection VotersOfElectionCollection
	cur, err = ms.voters.Find(ctx7, bson.D{{}})
	if err != nil {
		log.Warn(err)
		return "{}"
	}
	for cur.Next(ctx7) {
		var voter VotersOfElection
		err := cur.Decode(&voter)
		if err != nil {
			log.Warn(err)
		}
		votersOfElection.VotersOfElection = append(votersOfElection.VotersOfElection, voter)
	}

	ctx8, cancel8 := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel8()
	var censuses CensusCollection
	cur, err = ms.census.Find(ctx8, bson.D{{}})
	if err != nil {
		log.Warn(err)
	}
	for cur.Next(ctx8) {
		var census Census
		err := cur.Decode(&census)
		if err != nil {
			log.Warn(err)
		}
		censuses.Censuses = append(censuses.Censuses, census)
	}

	ctx9, cancel9 := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel9()
	var communitites CommunitiesCollection
	cur, err = ms.communitites.Find(ctx9, bson.D{{}})
	if err != nil {
		log.Warn(err)
	}
	for cur.Next(ctx8) {
		var community Community
		err := cur.Decode(&community)
		if err != nil {
			log.Warn(err)
		}
		communitites.Communities = append(communitites.Communities, community)
	}

	data, err := json.Marshal(&Collection{users, elections, results, votersOfElection, censuses, communitites})
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

	// Upsert VotersOfElection
	log.Infow("importing votersOfElection", "count", len(collection.VotersOfElection))
	for _, voter := range collection.VotersOfElection {
		filter := bson.M{"_id": voter.ElectionID}
		update := bson.M{"$set": voter}
		opts := options.Update().SetUpsert(true)
		_, err := ms.voters.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Warnw("Error upserting votersOfElection", "err", err, "election", voter.ElectionID)
		}
	}

	// Upsert Censuses
	log.Infow("importing censuses", "count", len(collection.Censuses))
	for _, census := range collection.Censuses {
		filter := bson.M{"_id": census.CensusID}
		update := bson.M{"$set": census}
		opts := options.Update().SetUpsert(true)
		_, err := ms.census.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Warnw("Error upserting census", "err", err, "census", census.CensusID)
		}
	}

	// Upsert Communities
	log.Infow("importing communitites", "count", len(collection.Communities))
	for _, community := range collection.Communities {
		filter := bson.M{"_id": community.ID}
		update := bson.M{"$set": community}
		opts := options.Update().SetUpsert(true)
		_, err := ms.census.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Warnw("Error upserting community", "err", err, "community", community.ID)
		}
	}

	log.Infof("imported database!")
	return nil
}

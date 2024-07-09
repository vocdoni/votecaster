package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type migrationFunc func(context.Context, *mongo.Database) error

var migrations = map[string]migrationFunc{
	"migrateCommunityID": migrateID,
}

func main() {
	connectionURI := flag.String("uri", "", "MongoDB connection URI")
	databaseName := flag.String("db", "", "Database name")
	migrationName := flag.String("migration", "", "Migration name")
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Replace with your MongoDB connection string
	clientOptions := options.Client().ApplyURI(*connectionURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	// Replace 'your_database' with your actual database name
	db := client.Database(*databaseName)
	// Define the migration function
	fn, ok := migrations[*migrationName]
	if !ok {
		availableMigrations := make([]string, 0, len(migrations))
		for migration := range migrations {
			availableMigrations = append(availableMigrations, migration)
		}
		fmt.Printf("Invalid migration name. Available migrations: %s\n", strings.Join(availableMigrations, ", "))
		return
	}
	// Run the migration
	if err := fn(ctx, db); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Migration completed successfully!")
}

func migrateID(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("elections")
	// Define a context with a timeout to ensure the migration doesn't run indefinitely
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Fetch all documents from the elections collection
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc bson.M
		if err = cursor.Decode(&doc); err != nil {
			return err
		}
		// Check if the 'community' sub-object and its 'id' attribute exist
		if community, ok := doc["community"].(bson.M); ok {
			if oldID, ok := community["id"].(int32); ok {
				newID := "degen:" + strconv.Itoa(int(oldID))
				// Update the document with the new id value
				filter := bson.M{"_id": doc["_id"]}
				update := bson.M{"$set": bson.M{"community.id": newID}}
				_, err := collection.UpdateOne(ctx, filter, update)
				if err != nil {
					return err
				}
			}
		}
	}
	return cursor.Err()
}

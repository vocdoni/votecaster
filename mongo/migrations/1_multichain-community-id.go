package migrations

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	migrate.MustRegister(upNewCommunityID, downNewCommunityID)
}

func upNewCommunityID(ctx context.Context, db *mongo.Database) error {
	// fetch all documents from the elections collection
	electionsCollection := db.Collection("elections")
	electionsCursor, err := electionsCollection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer electionsCursor.Close(ctx)
	// iterate over all documents
	for electionsCursor.Next(ctx) {
		var doc bson.M
		if err = electionsCursor.Decode(&doc); err != nil {
			return err
		}
		// check if the 'community' sub-object and its 'id' attribute exist
		if community, ok := doc["community"].(bson.M); ok {
			if oldID, ok := community["id"].(uint); ok {
				newID := fmt.Sprintf("degen:%d", oldID)
				// update the document with the new id value
				filter := bson.M{"_id": doc["_id"]}
				update := bson.M{"$set": bson.M{"community.id": newID}}
				_, err := electionsCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					return err
				}
			}
		}
	}
	if err := electionsCursor.Err(); err != nil {
		return err
	}
	// fetch all documents from the avatars collection
	avatarsCollection := db.Collection("avatars")
	avatarsCursor, err := avatarsCollection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer avatarsCursor.Close(ctx)
	// iterate over all documents
	for avatarsCursor.Next(ctx) {
		var doc bson.M
		if err = avatarsCursor.Decode(&doc); err != nil {
			return err
		}
		// check if the 'communityId' attribute exist
		if oldID, ok := doc["communityId"].(uint); ok {
			newID := fmt.Sprintf("degen:%d", oldID)
			// update the document with the new id value
			filter := bson.M{"_id": doc["_id"]}
			update := bson.M{"$set": bson.M{"communityId": newID}}
			_, err := avatarsCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return err
			}
		}
	}
	return avatarsCursor.Err()
}

func downNewCommunityID(ctx context.Context, db *mongo.Database) error {
	newIDRgx := regexp.MustCompile(`^degen:(\d+)$`)

	// fetch all documents from the elections collection
	electionsCollection := db.Collection("elections")
	electionsCursor, err := electionsCollection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer electionsCursor.Close(ctx)
	// iterate over all documents
	for electionsCursor.Next(ctx) {
		var doc bson.M
		if err = electionsCursor.Decode(&doc); err != nil {
			return err
		}
		// check if the 'community' sub-object and its 'id' attribute exist
		if community, ok := doc["community"].(bson.M); ok {
			if oldID, ok := community["id"].(string); ok {
				// by default, unset the community attribute if the id does not
				// match the regex
				update := bson.M{"$unset": bson.M{"community": ""}}
				filter := bson.M{"_id": doc["_id"]}
				// if the id matches the regex, parse the id and update the
				// document with the new id value
				if newIDRgx.MatchString(oldID) {
					newID, err := strconv.ParseUint(newIDRgx.FindStringSubmatch(oldID)[1], 10, 64)
					if err != nil {
						return err
					}
					update = bson.M{"$set": bson.M{"community.id": newID}}
				}
				if _, err := electionsCollection.UpdateOne(ctx, filter, update); err != nil {
					return err
				}
			}
		}
	}
	if err := electionsCursor.Err(); err != nil {
		return err
	}
	// fetch all documents from the avatars collection
	avatarsCollection := db.Collection("avatars")
	avatarsCursor, err := avatarsCollection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer avatarsCursor.Close(ctx)
	// iterate over all documents
	for avatarsCursor.Next(ctx) {
		var doc bson.M
		if err = avatarsCursor.Decode(&doc); err != nil {
			return err
		}
		// check if the 'communityId' attribute exist
		if oldID, ok := doc["communityId"].(string); ok {
			// by default, unset the communityId attribute if the id does not
			// match the regex
			update := bson.M{"$unset": bson.M{"communityId": ""}}
			filter := bson.M{"_id": doc["_id"]}
			// if the id matches the regex, parse the id and update the
			// document with the new id value
			if newIDRgx.MatchString(oldID) {
				newID, err := strconv.ParseUint(newIDRgx.FindStringSubmatch(oldID)[1], 10, 64)
				if err != nil {
					return err
				}
				update = bson.M{"$set": bson.M{"communityId": newID}}
			}
			if _, err := avatarsCollection.UpdateOne(ctx, filter, update); err != nil {
				return err
			}
		}
	}
	return avatarsCursor.Err()
}

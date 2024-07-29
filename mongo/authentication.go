package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AddAuthentication adds an authentication token for a user and updates the CreatedAt field to the current time.
func (ms *MongoStorage) AddAuthentication(userFID uint64, authToken string) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare the update with $addToSet for the authToken and $set for the UpdatedAt field
	update := bson.M{
		"$addToSet": bson.M{"authTokens": authToken},
		"$set":      bson.M{"updatedAt": time.Now()},
	}

	// Execute the update operation
	_, err := ms.authentications.UpdateOne(
		ctx,
		bson.M{"_id": userFID},
		update,
		options.Update().SetUpsert(true), // Upsert if the document doesn't exist
	)

	return err
}

// UpdateActivityAndGetData updates the activity timer and retrieves the Authentication data for a given authToken.
func (ms *MongoStorage) UpdateActivityAndGetData(authToken string) (*Authentication, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	authentication := &Authentication{}

	// Start a session for transaction
	session, err := ms.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	// Use a transaction to find the document and update it atomically
	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		// Search for the document where authTokens array contains the authToken
		if err := ms.authentications.FindOne(sc, bson.M{"authTokens": authToken}).Decode(authentication); err != nil {
			if err == mongo.ErrNoDocuments {
				return ErrUserUnknown
			}
			return err
		}

		// Update the updatedAt field for the found document
		update := bson.M{"$set": bson.M{"updatedAt": time.Now()}}
		_, err := ms.authentications.UpdateOne(
			sc,
			bson.M{"_id": authentication.UserID},
			update,
		)
		return err
	})

	if err != nil {
		return nil, err
	}

	// Return the found and updated Authentication data
	return authentication, nil
}

// CheckAuthentication checks if the authToken is valid and returns the corresponding userID.
// If the token is not found, it returns ErrUserUnknown.
func (ms *MongoStorage) UserFromAuthToken(authToken string) (uint64, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	var authData Authentication
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ms.authentications.FindOne(ctx, bson.M{"authTokens": authToken}).Decode(&authData)
	if err != nil {
		return 0, ErrUserUnknown
	}

	return authData.UserID, nil
}

// UserAuthorizations method returns the tokens of a user for the fid provider.
// If the user is not found, it returns ErrUserUnknown.
func (ms *MongoStorage) UserAuthorizations(userFID uint64) ([]string, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var authData Authentication
	err := ms.authentications.FindOne(ctx, bson.M{"_id": userFID}).Decode(&authData)
	if err != nil {
		return nil, ErrUserUnknown
	}
	return authData.AuthTokens, nil
}

// Authentications returns the full list of authTokens.
func (ms *MongoStorage) Authentications() ([]string, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Find all documents without any specific filter
	cur, err := ms.authentications.Find(ctx, bson.D{{}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var allTokens []string
	for cur.Next(ctx) {
		var authData Authentication
		if err := cur.Decode(&authData); err != nil {
			continue
		}
		allTokens = append(allTokens, authData.AuthTokens...)
	}

	return allTokens, nil
}

package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetDelegation inserts a delegation into the database and returns the ID of
// the inserted delegation
func (ms *MongoStorage) SetDelegation(delegation Delegation) (string, error) {
	// Insert the delegation into the database and retrieve the ID
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	delegation.ID = primitive.NewObjectID()
	if _, err := ms.delegations.InsertOne(ctx, delegation); err != nil {
		return "", err
	}
	return delegation.ID.Hex(), nil
}

// Delegation retrieves a delegation from the database by its ID
func (ms *MongoStorage) Delegation(id string) (Delegation, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Delegation{}, err
	}

	var delegation Delegation
	if err = ms.delegations.FindOne(ctx, bson.M{"_id": _id}).Decode(&delegation); err != nil {
		return Delegation{}, err
	}
	return delegation, nil
}

// DelegationsFrom retrieves all delegations from a user by their user ID provided
func (ms *MongoStorage) DelegationsFrom(userID uint64) ([]*Delegation, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return ms.filterDelegations(ctx, bson.M{"from": userID})
}

// DelegationsTo retrieves all delegations to a user by their user ID provided
func (ms *MongoStorage) DelegationsTo(userID uint64) ([]*Delegation, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return ms.filterDelegations(ctx, bson.M{"to": userID})
}

// DelegationsByCommunity retrieves all delegations to a community by the
// community ID provided
func (ms *MongoStorage) DelegationsByCommunity(communityID string) ([]*Delegation, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return ms.filterDelegations(ctx, bson.M{"communityId": communityID})
}

// FinalDelegationsByCommunity retrieves all delegations to a community by the
// community ID provided, solving nested delegations
func (ms *MongoStorage) FinalDelegationsByCommunity(communityID string) ([]*Delegation, error) {
	communityDelegations, err := ms.DelegationsByCommunity(communityID)
	if err != nil {
		return nil, err
	}
	return solveNestedDelegations(communityDelegations, nil), nil
}

// DelegationsByCommunityFrom retrieves all delegations from a user to a
// community by the community ID and user ID provided
func (ms *MongoStorage) DelegationsByCommunityFrom(communityID string, userID uint64) ([]*Delegation, error) {
	ms.keysLock.RLock()
	defer ms.keysLock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	communityDelegations, err := ms.filterDelegations(ctx, bson.M{"communityId": communityID})
	if err != nil {
		return nil, err
	}
	userDelegations := []*Delegation{}
	for _, delegation := range communityDelegations {
		if delegation.From == userID {
			userDelegations = append(userDelegations, delegation)
		}
	}
	return solveNestedDelegations(communityDelegations, userDelegations), nil
}

// DeleteDelegation deletes a delegation from the database by its ID
func (ms *MongoStorage) DeleteDelegation(id string) error {
	ms.keysLock.Lock()
	defer ms.keysLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = ms.delegations.DeleteOne(ctx, bson.M{"_id": _id})
	return err
}

func (ms *MongoStorage) filterDelegations(ctx context.Context, filter bson.M) ([]*Delegation, error) {
	cursor, err := ms.delegations.Find(ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	var delegations []*Delegation
	err = cursor.All(ctx, &delegations)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return delegations, nil
}

// solveNestedDelegations itereates over the list of delegations and solves
// chains of delegations, for example, if user A delegates to user B and user B
// delegates to user C, the function will return a list of delegations where
// user A delegates to user C and user B delegates to user C.
func solveNestedDelegations(original, filtered []*Delegation) []*Delegation {
	if filtered == nil {
		filtered = append([]*Delegation{}, original...)
	}
	finalDelegations := []*Delegation{}
	for _, delegation := range filtered {
		// check if the delegation is to a user that has already delegated to
		// another user
		delegateDelegations := []*Delegation{}
		for _, originalDelegation := range original {
			if originalDelegation.From == delegation.To {
				delegateDelegations = append(delegateDelegations, originalDelegation)
			}
		}
		if len(delegateDelegations) == 0 {
			finalDelegations = append(finalDelegations, delegation)
			continue
		}
		// solve the nested delegations for the current delegation and append
		// them to the final list
		for _, nestedDelegation := range solveNestedDelegations(original, delegateDelegations) {
			// keep the original delegation ID and from user
			nestedDelegation.ID = delegation.ID
			nestedDelegation.From = delegation.From
			finalDelegations = append(finalDelegations, nestedDelegation)
		}
	}
	return finalDelegations
}

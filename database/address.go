package database

import (
	"context"
	"errors"
	"time"

	"github/akhil/ecommerce-yt/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindUser      = errors.New("can't find user")
	ErrCantUpdateAddress = errors.New("can't update address")
	ErrCantDeleteAddress = errors.New("can't delete address")
	ErrAddressNotFound   = errors.New("address not found")
	ErrNoAddresses       = errors.New("no addresses found")
)

// AddAddress adds a new address to the user's address list
func AddAddress(userCollection *mongo.Collection, userID string, address models.Address) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the user
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", ErrCantFindUser
		}
		return "", err
	}

	// Generate address ID if not provided
	if address.Address_ID.IsZero() {
		address.Address_ID = primitive.NewObjectID()
	}

	// Add address to user's address list
	user.Address_Details = append(user.Address_Details, address)

	// Update user in database
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "address_details", Value: user.Address_Details},
		{Key: "updated_at", Value: time.Now()},
	}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return "", ErrCantUpdateAddress
	}

	return address.Address_ID.Hex(), nil
}

// EditHomeAddress updates the home address (first address in the list)
func EditHomeAddress(userCollection *mongo.Collection, userID string, address models.Address) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the user
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", ErrCantFindUser
		}
		return "", err
	}

	// Check if address list is empty
	if len(user.Address_Details) == 0 {
		return "", ErrNoAddresses
	}

	// Update the first address (home address)
	if address.Address_ID.IsZero() {
		address.Address_ID = user.Address_Details[0].Address_ID
	}
	user.Address_Details[0] = address

	// Update user in database
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "address_details", Value: user.Address_Details},
		{Key: "updated_at", Value: time.Now()},
	}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return "", ErrCantUpdateAddress
	}

	return address.Address_ID.Hex(), nil
}

// EditWorkAddress updates the work address (second address in the list, or creates if doesn't exist)
func EditWorkAddress(userCollection *mongo.Collection, userID string, address models.Address) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the user
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", ErrCantFindUser
		}
		return "", err
	}

	// If work address doesn't exist (less than 2 addresses), add it
	if len(user.Address_Details) < 2 {
		if address.Address_ID.IsZero() {
			address.Address_ID = primitive.NewObjectID()
		}
		user.Address_Details = append(user.Address_Details, address)
	} else {
		// Update the second address (work address)
		if address.Address_ID.IsZero() {
			address.Address_ID = user.Address_Details[1].Address_ID
		}
		user.Address_Details[1] = address
	}

	// Update user in database
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "address_details", Value: user.Address_Details},
		{Key: "updated_at", Value: time.Now()},
	}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return "", ErrCantUpdateAddress
	}

	return address.Address_ID.Hex(), nil
}

// DeleteAddress deletes an address by address_id
func DeleteAddress(userCollection *mongo.Collection, userID string, addressID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the user
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrCantFindUser
		}
		return err
	}

	// Remove address from list
	var updatedAddresses []models.Address
	found := false
	for _, addr := range user.Address_Details {
		if addr.Address_ID != addressID {
			updatedAddresses = append(updatedAddresses, addr)
		} else {
			found = true
		}
	}

	if !found {
		return ErrAddressNotFound
	}

	// Update user in database
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "address_details", Value: updatedAddresses},
		{Key: "updated_at", Value: time.Now()},
	}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return ErrCantDeleteAddress
	}

	return nil
}

// GetAddresses retrieves all addresses for a user
func GetAddresses(userCollection *mongo.Collection, userID string) ([]models.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the user
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrCantFindUser
		}
		return nil, err
	}

	return user.Address_Details, nil
}

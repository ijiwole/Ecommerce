package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DBSet initializes and returns a connected MongoDB client.
func DBSet() *mongo.Client {
	mongoURI := os.Getenv("MONGODB_URL")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// Create context with timeout for connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
	}

	log.Println("Successfully connected to MongoDB!")
	return client
}

var Client *mongo.Client = DBSet()

// UserData returns a handle to the specified user-related collection.
func UserData(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("Ecommerce").Collection(collectionName)
	return collection
}

// ProductData returns a handle to the specified product-related collection.
func ProductData(client *mongo.Client, collectionName string) *mongo.Collection {
	var productCollection *mongo.Collection = client.Database("Ecommerce").Collection(collectionName)
	return productCollection
}

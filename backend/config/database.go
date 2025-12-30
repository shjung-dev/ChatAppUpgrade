package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func ConnectDatabase(uri string) {
	log.Println("Connecting to MongoDB....")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	err = c.Ping(ctx, nil)

	if err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}

	log.Println("Successfully connected to MongoDB")

	client = c
}

func OpenCollection(collectionName string) *mongo.Collection {
	if client == nil {
		log.Fatal("MongoDB Client is not initialized. Please connect DB first")
	}

	return client.Database("ChatApplication").Collection(collectionName)
}

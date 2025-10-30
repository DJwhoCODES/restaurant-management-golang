package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBinstance() *mongo.Client {
	MongoURI := "mongodb://localhost:27017"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}

	fmt.Println("✅ Connected to MongoDB successfully!")
	return client
}

var MongoClient *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("restaurant_management").Collection(collectionName)
}

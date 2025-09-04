package initializers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// ConnectDB initializes MongoDB connection
func ConnectDB() {
	uri := os.Getenv("MONGODB_URI")

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Create a short-lived context for connecting
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ MongoDB connection error:", err)
	}

	// Ping the database to test connection
	if err = Client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ MongoDB ping failed:", err)
	}

	fmt.Println("✅ Connected to MongoDB!")
}

// GetCollection returns a MongoDB collection
func GetCollection(dbName, colName string) *mongo.Collection {
	return Client.Database(dbName).Collection(colName)
}

// CloseDB disconnects MongoDB safely
func CloseDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Disconnect(ctx); err != nil {
		log.Fatal("❌ Error disconnecting MongoDB:", err)
	}
	fmt.Println("🔌 MongoDB connection closed.")
}

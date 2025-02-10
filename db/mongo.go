package db

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

var client *mongo.Client
var Database *mongo.Database
var customerCollection *mongo.Collection

func InitMongoDB() {
	// Load environment
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Msg("Error loading .env file")
	}

	// Initialize Mongo DB client
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set Global customer collection
	var err2 error
	client, err2 = mongo.Connect(ctx, clientOptions)
	if err2 != nil {
		log.Fatal().Err(err2).Msg("Failed to connect to MongoDB: %v")
	}
	customerCollection = client.Database("onboarding").Collection("customer")
}

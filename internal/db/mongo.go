package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client *mongo.Client
}

func NewMongoDB() (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:example@localhost:27017"))
	if err != nil {
		return nil, err
	}

	return &MongoDB{Client: client}, nil
}

func (m *MongoDB) Close(ctx context.Context) {
	log.Println("Closing MongoDB connection...")
	if m.Client != nil {
		if err := m.Client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}
	log.Println("MongoDB shutdown complete.")
}

func (m *MongoDB) IsConnected() bool {
	if m.Client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := m.Client.Ping(ctx, nil)
	if err != nil {
		log.Printf("MongoDB Ping failed: %v", err)
		return false
	}
	return true
}

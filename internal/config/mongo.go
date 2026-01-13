package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	URI      string
	Database string
}

func NewMongoConnection(cfg MongoConfig) *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		log.Fatal("mongo connect error:", err)
	}

	// ping เพื่อเช็กการเชื่อมต่อ
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("mongo ping error:", err)
	}

	log.Println("✅ MongoDB connected")

	return client.Database(cfg.Database)
}

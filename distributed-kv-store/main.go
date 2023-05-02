package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connect() (*mongo.Database, error) {
	clientOptions := options.Client().ApplyURI(
		"mongodb://mongo1:27017,mongo2:27017,mongo3:27017/?replicaSet=rs",
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo. [%w]", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping. [%w]", err)
	}
	log.Info().Msg("connected to mongo")

	return client.Database("db"), nil
}

func main() {
	db, err := connect()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect")
	}
	_, err = db.Collection("c").InsertOne(context.Background(), bson.M{"key": "val"})
	if err != nil {
		log.Error().Err(err).Msg("failed to insert")
	}
}

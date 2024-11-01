package model

import (
	"GoCrawl/internal/log"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	Setup()
}

var DB *mongo.Database
var content *mongo.Collection
var ctx = context.TODO()

func Setup() {
	log.Info("Initializing MongoDB")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		panic(err)
	}

	DB = client.Database("crawler")
	content = DB.Collection("content")

	log.Info("MongoDB Initialized")
	buildIndex()
}

func buildIndex() {
	var err error
	_, err = content.Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"domain": 1,
			},
			Options: options.Index().SetName("domains_index"),
		})

	if err != nil {
		log.Error("failed to create index for domain, err:%s", err.Error())
		panic(err)
	}
}

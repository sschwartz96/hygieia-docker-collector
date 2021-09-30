package internal

import "go.mongodb.org/mongo-driver/mongo"

type Store interface {
}

type MongoStore struct {
	client *mongo.Client
}

func NewMongoStore(client *mongo.Client) *MongoStore {
	return &MongoStore{client: client}
}

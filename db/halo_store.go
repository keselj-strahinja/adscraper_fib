package db

import (
	"context"

	models "github.com/keselj-strahinja/halo_scraper/type_models"
	"go.mongodb.org/mongo-driver/mongo"
)

type HaloStore interface {
	InsertListing(context.Context, *models.Apartment) (*models.Apartment, error)
}

type MongoHaloStore struct {
	client *mongo.Client
	col    *mongo.Collection
}

func NewMongoHaloStore(client *mongo.Client) *MongoHaloStore {

	return &MongoHaloStore{
		client: client,
		col:    client.Database("oglasi").Collection("halo"),
	}
}

func (h *MongoHaloStore) InsertListing(context.Context, *models.Apartment) (*models.Apartment, error) {

	return nil, nil
}

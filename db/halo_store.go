package db

import (
	"context"
	"log"

	models "github.com/keselj-strahinja/halo_scraper/type_models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type HaloStore interface {
	InsertListing(context.Context, *models.Apartment) (*models.Apartment, error)
	URLExists(context.Context, string) (bool, error)
	GetApartmant(context.Context, string) (*models.Apartment, error)
	GetActiveListings(context.Context, bson.M) ([]*models.Apartment, error)
	SetAllInactive(context.Context) error
	SetActive(context.Context, string) error
	SetScraped(context.Context, string, bool) error
	GetUnscrapedURLs(context.Context) ([]string, error)
	UpdateListing(context.Context, *models.Apartment) error
}

type MongoHaloStore struct {
	client *mongo.Client
	col    *mongo.Collection
}

func NewMongoHaloStore(client *mongo.Client) *MongoHaloStore {

	return &MongoHaloStore{
		client: client,
		col:    client.Database("oglasi-scraper").Collection("halo-oglasi"),
	}
}

func (h *MongoHaloStore) InsertListing(ctx context.Context, apt *models.Apartment) (*models.Apartment, error) {

	res, err := h.col.InsertOne(ctx, apt)

	if err != nil {
		return nil, err
	}
	apt.ID = res.InsertedID.(primitive.ObjectID)
	return apt, nil
}

func (h *MongoHaloStore) URLExists(ctx context.Context, url string) (bool, error) {
	res := h.col.FindOne(ctx, bson.M{"url": url})

	var result models.Apartment
	err := res.Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// The URL does not exist in the database.
			return false, nil
		}

		return false, err
	}
	// The URL exists in the database.
	return true, nil
}

func (h *MongoHaloStore) GetActiveListings(ctx context.Context, filter bson.M) ([]*models.Apartment, error) {
	resp, err := h.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var listings []*models.Apartment
	if err := resp.All(ctx, &listings); err != nil {
		return nil, err
	}
	return listings, nil
}

func (h *MongoHaloStore) GetApartmant(ctx context.Context, url string) (*models.Apartment, error) {
	res := h.col.FindOne(ctx, bson.M{"url": url})

	var result models.Apartment
	err := res.Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, err
		}

		return nil, err
	}
	// The URL exists in the database.
	return &result, nil
}
func (h *MongoHaloStore) SetAllInactive(ctx context.Context) error {
	// Create a filter that matches all documents
	filter := bson.M{}

	// Create an update that sets "active" to false for all documents
	update := bson.M{
		"$set": bson.M{"active": false},
	}

	// Update all documents
	_, err := h.col.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println("No docs in mongo")
	}
	return nil
}

func (h *MongoHaloStore) SetActive(ctx context.Context, url string) error {
	// Create a filter that matches the specific document with the given URL
	filter := bson.M{"url": url}

	// Create an update that sets "active" to true for the matched document
	update := bson.M{
		"$set": bson.M{"active": true},
	}

	// Update the document
	_, err := h.col.UpdateOne(ctx, filter, update)
	return err
}

func (h *MongoHaloStore) SetScraped(ctx context.Context, url string, scraped bool) error {
	// Create a filter that matches the specific document with the given URL
	filter := bson.M{"url": url}

	// Create an update that sets "active" to true for the matched document
	update := bson.M{
		"$set": bson.M{"scraped": scraped},
	}

	// Update the document
	_, err := h.col.UpdateOne(ctx, filter, update)
	return err
}

func (h *MongoHaloStore) GetUnscrapedURLs(ctx context.Context) ([]string, error) {
	// Define a filter for documents where "scraped" is false
	filter := bson.M{"scraped": false}

	// Use Find to get a cursor for all matching documents
	cur, err := h.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var urls []string
	for cur.Next(ctx) {
		var apartment models.Apartment
		err := cur.Decode(&apartment)
		if err != nil {
			log.Println("error decoding document:", err)
			continue
		}
		urls = append(urls, apartment.URL)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (h *MongoHaloStore) UpdateListing(ctx context.Context, a *models.Apartment) error {
	logrus.Info("Updating listing")
	filter := bson.M{"_id": a.ID}
	update := bson.M{
		"$set": bson.M{
			"name":            a.Name,
			"city":            a.City,
			"location":        a.Location,
			"microloc":        a.Microlocation,
			"street":          a.Street,
			"realEstateType":  a.RealEstateType,
			"buildingType":    a.BuildingType,
			"state":           a.State,
			"furnishing":      a.Furnishing,
			"heating":         a.Heating,
			"floor":           a.Floor,
			"buildingFloors":  a.BuildingFloors,
			"monthlyBills":    a.MonthlyBills,
			"paymentType":     a.PaymentType,
			"additionalPerks": a.AdditionalPerks,
			"otherPerks":      a.OtherPerks,
			"owner":           a.Owner,
			"phone":           a.Phone,
			"url":             a.URL,
			"squareSize":      a.SquareSize,
			"rooms":           a.Rooms,
			"desc":            a.Description,
			"postingDate":     a.PostingDate,
			"priceValue":      a.PriceValue,
			"priceUnit":       a.PriceUnit,
			"active":          a.Active,
			"scraped":         a.Scraped,
			"imageURLS":       a.ImageURL,
			"contactName":     a.ContactName,
		},
	}
	_, err := h.col.UpdateOne(ctx, filter, update)
	return err
}

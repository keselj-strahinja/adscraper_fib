package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Apartment struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `json:"name" bson:"name"`
	City            string             `json:"city" bson:"city"`
	Location        string             `json:"location" bson:"location"`
	Microlocation   string             `json:"microloc" bson:"microloc"`
	Street          string             `json:"street" bson:"street"`
	RealEstateType  string             `json:"realEstateType" bson:"realEstateType"`
	BuildingType    string             `json:"buildingType" bson:"buildingType"`
	State           string             `json:"state" bson:"state"`
	Furnishing      string             `json:"furnishing" bson:"furnishing"`
	Heating         string             `json:"heating" bson:"heating"`
	Floor           string             `json:"floor" bson:"floor"`
	BuildingFloors  int64              `json:"buildingFloors" bson:"buildingFloors"`
	MonthlyBills    int64              `json:"monthlyBills" bson:"monthlyBills"`
	PaymentType     string             `json:"paymentType" bson:"paymentType"`
	AdditionalPerks []string           `json:"additionalPerks" bson:"additionalPerks"`
	OtherPerks      []string           `json:"otherPerks" bson:"otherPerks"`
	Owner           string             `json:"owner" bson:"owner"`
	Phone           string             `json:"phone" bson:"phone"`
	ContactName     string             `json:"contactName" bson:"contactName"`
	URL             string             `json:"url" bson:"url"`
	SquareSize      int64              `json:"squareSize" bson:"squareSize"`
	Rooms           float64            `json:"rooms" bson:"rooms"`
	Description     []string           `json:"desc" bson:"desc"`
	PostingDate     int64              `json:"postingDate" bson:"postingDate"`
	PriceValue      int64              `json:"priceValue" bson:"priceValue"`
	PriceUnit       string             `json:"priceUnit" bson:"priceUnit"`
	Active          bool               `json:"active" bson:"active"`
	Scraped         bool               `json:"scraped" bson:"scraped"`
	Unavailable         bool               `json:"unavailable" bson:"unavailable"`

	ImageURL []string `json:"imageURLS" bson:"imageURLS"`
}

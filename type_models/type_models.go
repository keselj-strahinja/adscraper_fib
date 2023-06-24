package models

type Apartment struct {
	Name        string  `json:"name" bson:"name"`
	URL         string  `json:"url" bson:"url"`
	Size        int64   `json:"size,omitempty" bson:"size,omitempty"`
	Rooms       float64 `json:"rooms,omitempty" bson:"rooms,omitempty"`
	Description string  `json:"desc,omitempty" bson:"desc,omitempty"`
	PostingDate string  `json:"postingDate,omitempty" bson:"postingDate,omitempty"`
	Price       int64   `json:"price" bson:"price"`
	Floor       string  `json:"floor" bson:"floor"`
	ImageURL    string  `json:"imageURL" bson:"imageURL"`
	Location    string  `json:"location" bson:"location"`
	Agency      string  `json:"agency" bson:"agency"`
}

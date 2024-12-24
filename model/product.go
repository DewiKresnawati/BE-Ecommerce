package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// Product model represents the product schema for MongoDB
type Product struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	Price    float64            `json:"price" bson:"price"`
	Discount float64            `json:"discount" bson:"discount"`
	Category string             `json:"category" bson:"category"`
	Image    string             `json:"image" bson:"image"`
	Rating   float64            `json:"rating" bson:"rating"`
	Reviews  int                `json:"reviews" bson:"reviews"`
}

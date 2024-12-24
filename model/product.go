package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// Product model represents the product schema for MongoDB
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Price       float64            `json:"price" bson:"price"`
	Category    string             `json:"category" bson:"category"`
	Stock       int                `json:"stock" bson:"stock"`
	ImageURL    string             `json:"image_url" bson:"image_url"`
}

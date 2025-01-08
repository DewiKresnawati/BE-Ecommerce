package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// Review represents the review schema for a product
type Review struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ProductID primitive.ObjectID `json:"product_id" bson:"product_id"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Rating    float64            `json:"rating" bson:"rating"`
	Comment   string             `json:"comment" bson:"comment"`
	CreatedAt int64              `json:"created_at" bson:"created_at"`
}

package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// Category represents the schema for storing product categories
type Category struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name       string             `json:"name" bson:"name"`
	SubCategories []SubCategory   `json:"sub_categories" bson:"sub_categories"`
}

// SubCategory represents a sub-category under a category
type SubCategory struct {
	ID   primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name"`
}

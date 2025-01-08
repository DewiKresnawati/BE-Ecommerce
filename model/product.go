package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// Product model represents the product schema for MongoDB
type Product struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name          string             `json:"name" bson:"name"`
	Price         float64            `json:"price" bson:"price"`
	Discount      float64            `json:"discount" bson:"discount"`
	Image         string             `json:"image" bson:"image"`
	Rating        float64            `json:"rating" bson:"rating"`
	Reviews       int                `json:"reviews" bson:"reviews"`
	Description   string             `json:"description" bson:"description"`
	SellerID      primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	CategoryID    primitive.ObjectID `json:"category_id" bson:"category_id"`
	SubCategoryID primitive.ObjectID `json:"sub_category_id" bson:"sub_category_id"`
}

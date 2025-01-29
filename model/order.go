package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Order model untuk menyimpan data pesanan
type Order struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	Items          []OrderItem        `bson:"items" json:"items"`
	TotalAmount    int                `bson:"total_amount" json:"total_amount"`
	ShippingCost   int                `bson:"shipping_cost" json:"shipping_cost"`
	ShippingAddress string            `bson:"shipping_address" json:"shipping_address"`
	Status        string              `bson:"status" json:"status"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	PaymentToken   string             `bson:"payment_token,omitempty" json:"payment_token"`
}

// OrderItem menyimpan item dalam sebuah order
type OrderItem struct {
	Name     string `bson:"name" json:"name"`
	Quantity int    `bson:"quantity" json:"quantity"`
	Price    int    `bson:"price" json:"price"`
}

package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Order model untuk menyimpan data pesanan
type Order struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	SellerID	   primitive.ObjectID `bson:"seller_id" json:"seller_id"`
	Items          []OrderItem        `bson:"items" json:"items"`
	TotalAmount    int                `bson:"total_amount" json:"total_amount"`
	ShippingCost   int                `bson:"shipping_cost" json:"shipping_cost"`
	ShippingAddress string            `bson:"shipping_address" json:"shipping_address"`
	Status string 					  `bson:"status" json:"status" validate:"oneof=Pending Processing Shipped Delivered"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	PaymentDate time.Time 			  `bson:"payment_date,omitempty" json:"payment_date"`
	PaymentToken   string             `bson:"payment_token,omitempty" json:"payment_token"`
}

// OrderItem menyimpan item dalam sebuah order
type OrderItem struct {
	ProductID primitive.ObjectID `bson:"product_id,omitempty" json:"product_id"`
	Name      string             `bson:"name" json:"name"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Price     int                `bson:"price" json:"price"`
	SellerID  primitive.ObjectID `bson:"seller_id,omitempty" json:"seller_id"`
	Product   *Product           `bson:"product,omitempty" json:"product"`  // Tambahkan informasi produk langsung di OrderItem
}

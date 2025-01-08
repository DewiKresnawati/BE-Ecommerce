package model

type Cart struct {
	UserID   string     `json:"user_id" bson:"user_id"`
	Products []CartItem `json:"products" bson:"products"`
}

type CartItem struct {
	UserID    string  `json:"user_id" bson:"user_id"`
	ProductID string  `json:"product_id" bson:"product_id"`
	Quantity  int     `json:"quantity" bson:"quantity"`
	Price     int 	  `json:"price" bson:"price"`
}


package model

type Cart struct {
	UserID   string     `json:"user_id" bson:"user_id"`
	Products []CartItem `json:"products" bson:"products"`
}

type CartItem struct {
	ProductID string  `json:"product_id" bson:"product_id"`
	Quantity  int     `json:"quantity" bson:"quantity"`
	Price     float64 `json:"price" bson:"price"`
}

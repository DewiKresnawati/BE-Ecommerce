package model

type CartItem struct {
    ProductID   string `json:"product_id" bson:"product_id"`
    ProductName string `json:"product_name" bson:"product_name"`
    Price       int    `json:"price" bson:"price"`
    Quantity    int    `json:"quantity" bson:"quantity"`
    UserID      string `json:"user_id" bson:"user_id"` // Tambahkan field UserID
}


type Cart struct {
	UserID   string     `json:"user_id" bson:"user_id"`
	Products []CartItem `json:"products" bson:"products"`
}
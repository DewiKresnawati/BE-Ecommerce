package model

type Favorite struct {
	UserID    string   `json:"user_id" bson:"user_id"`
	ProductIDs []string `json:"product_ids" bson:"product_ids"`
}

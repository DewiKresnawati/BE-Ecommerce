package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// LoginRequest represents a user login payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User represents the user schema for MongoDB
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Username    string             `json:"username" bson:"username"`
	Email       string             `json:"email" bson:"email"`
	Password    string             `json:"password" bson:"password"`
	Roles       []string           `json:"roles" bson:"roles"`                 // Example: ["customer"], ["customer", "seller"]
	StoreStatus *string            `json:"store_status,omitempty" bson:"store_status,omitempty"` // "pending", "approved", "rejected"
	StoreInfo   *StoreInfo         `json:"store_info,omitempty" bson:"store_info,omitempty"`
}

// StoreInfo represents additional information for becoming a seller
type StoreInfo struct {
	StoreName    string `json:"store_name" bson:"store_name"`
	FullAddress  string `json:"full_address" bson:"full_address"`
	NIK          string `json:"nik" bson:"nik"` // National Identity Number
}
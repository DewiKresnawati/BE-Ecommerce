package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	Password    string             `json:"password" bson:"password"` // Tambahkan field Password
	Roles       []string           `json:"roles" bson:"roles"`
	SellerID    *primitive.ObjectID  `bson:"seller_id,omitempty" json:"seller_id,omitempty"`
	StoreStatus *string            `json:"store_status,omitempty" bson:"store_status,omitempty"`
	StoreInfo   *StoreInfo         `json:"store_info,omitempty" bson:"store_info,omitempty"`
	ResetToken       string             `json:"reset_token,omitempty" bson:"reset_token,omitempty"`
	ResetTokenExpiry time.Time          `json:"reset_token_expiry,omitempty" bson:"reset_token_expiry,omitempty"`
}
type StoreInfo struct {
	StoreName   string `json:"store_name" bson:"store_name"`
	FullAddress string `json:"full_address" bson:"full_address"`
	NIK         string `json:"nik" bson:"nik"`
	PhotoSelfie string `json:"photo_selfie" bson:"photo_selfie"`
}

// StoreInfo represents additional information for becoming a seller
type RequestPayload struct {
	StoreName   string `json:"store_name"`
	FullAddress string `json:"full_address"`
	NIK         string `json:"nik"`
	PhotoBase64 string `json:"photo_base64"`
}

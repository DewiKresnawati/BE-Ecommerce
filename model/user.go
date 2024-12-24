package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User model represents the user schema for MongoDB
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username"`
	Email    string             `json:"email" bson:"email"`
	Password string             `json:"password" bson:"password"`
	Role     string             `json:"role" bson:"role"` // For example: "admin", "customer"
}

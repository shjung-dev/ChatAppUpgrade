package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Username      *string            `json:"username" validate:"required,min=2,max=100"`
	Password      *string            `json:"password" validate:"required,min=6"`
	Token         *string            `json:"token,omitempty"`
	Refresh_token *string            `json:"refresh_token,omitempty"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
	User_id       string             `json:"user_id"`
}
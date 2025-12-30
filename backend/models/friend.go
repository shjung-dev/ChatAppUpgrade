package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Friend struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Username       *string            `bson:"username"`
	FriendUsername *string            `bson:"friendusername"`
}

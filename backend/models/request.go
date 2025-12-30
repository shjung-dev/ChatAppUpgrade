package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Request struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	From   string             `bson:"from"`
	To     string             `bson:"to"`
	Status string             `bson:"status"`
}

package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Group struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Hostname *string             `bson:host`
	GroupID   *string             `bson:"groupID"`
	Members  []string           `bson:"members"`
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	From   *string            `json:"from"`
	To     *string            `json:"to"`
	Text   *string            `json:"text"`
	SentAt time.Time         `json:"sent_at`
}

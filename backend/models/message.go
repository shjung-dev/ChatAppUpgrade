package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	ConversationID string             `bson:"conversationID"`
	SenderUserName string             `bson:"senderUserName"`
	Content        string             `bson:"content"`
	CreatedAt      time.Time          `bson:"created_at"`
}

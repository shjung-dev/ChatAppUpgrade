package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Conversation struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	ConversationID   string             `bson:"conversationID"`
	ConversationName *string            `bson:"conversationName,omitempty"`
	Participants     []string           `bson:"participants"`
	CreatedAt        time.Time          `bson:"created_at"`
	LastMessageAt    time.Time          `bson:"lastMessageAt"`
}

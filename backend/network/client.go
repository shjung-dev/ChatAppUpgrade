package network

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/shjung-dev/ChatApplication/backend/config"
	"github.com/shjung-dev/ChatApplication/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Client struct {
	Socket   *websocket.Conn
	Receive  chan []byte
	Room     *Room
	Username string
}

type WSMessage struct {
	Type           string   `json:"type"`
	To             string   `json:"to"`
	ConvoID        string   `json:"convoID"`
	GroupName      string   `json:"groupName"`
	Members        []string `json:"members"`
	MessageContent string   `json:"messageContent"`
}

type OutgoingMessage struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

func (c *Client) Read() {

	defer func() {
		c.Room.Leave <- c
		close(c.Receive)
		c.Socket.Close()
	}()

	for {
		_, raw, err := c.Socket.ReadMessage()
		if err != nil {
			return
		}

		var msg WSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Println(err)
			continue
		}

		from := c.Username

		switch msg.Type {
		case "friend_request":
			requestCollection := config.OpenCollection("request")
			_, err := requestCollection.InsertOne(context.Background(), bson.M{
				"from":   from,
				"to":     msg.To,
				"status": "pending",
			})
			if err != nil {
				log.Println("Failed to insert friend request:", err)
				continue
			}
			//Check if the recipient client is also online
			toClient, online := onlineClients[msg.To]
			if online {
				//Client is online too so immediately send over websocket
				jsMsg, err := json.Marshal(OutgoingMessage{
					From: from,
					To:   msg.To,
					Type: msg.Type,
				})
				if err != nil {
					log.Println("Failed to marshal message:", err)
					continue
				}
				toClient.Receive <- jsMsg
				continue
			}
		case "friend_list_update":
			friendCollection := config.OpenCollection("friend")
			cursor, err := friendCollection.Find(context.Background(), bson.M{
				"username": from,
			})

			if err != nil {
				log.Println("Failed to fetch pending requests:", err)
				return
			}

			defer cursor.Close(context.Background())

			var friends []bson.M

			if err := cursor.All(context.Background(), &friends); err != nil {
				log.Println(err.Error())
				return
			}

			response, err := json.Marshal(map[string]interface{}{
				"type":    "friend_list_update",
				"friends": friends,
			})
			c.Receive <- response
			continue
		case "message":
			currentUser := c.Username
			announceMessage := false

			//Check if this convo exists
			convoCollection := config.OpenCollection("conversation")
			var convo models.Conversation
			err := convoCollection.FindOne(context.Background(), bson.M{"conversationID": msg.ConvoID}).Decode(&convo)

			if err != nil {
				//This convo whether it is 1to1 or group doesn't exist

				announceMessage = true
				convoID := uuid.NewString()
				participants := append(msg.Members, currentUser)
				participants = unique(participants)

				/*Update Convo in database
				-> This convo will now persist for all users
				*/
				convo = models.Conversation{
					ID:               primitive.NewObjectID(),
					ConversationID:   convoID,
					ConversationName: &msg.GroupName, //If it is 1 to 1 chat, this will be null
					Participants:     participants,
					CreatedAt:        time.Now(),
				}
				_, insertErr1 := convoCollection.InsertOne(context.Background(), convo)
				if insertErr1 != nil {
					log.Println("Failed to insert convo:", insertErr1.Error())
					continue
				}
			}

			//This convo already exists OR Finished adding new convo
			//We still need to update Convo whether it is a new or existing because received new message
			filter := bson.M{
				"conversationID": convo.ConversationID,
			}

			update := bson.M{
				"$set": bson.M{
					"lastMessageAt": time.Now(),
				},
			}

			result, err := convoCollection.UpdateOne(context.Background(), filter, update)

			if err != nil {
				log.Println("Failed to update convo:", err.Error())
				continue
			}

			if result.MatchedCount == 0 {
				log.Println("Convo is not found")
				return
			}

			//Update Message
			messageCollection := config.OpenCollection("message")
			var m models.Message

			//Differentiate between 1-to-1 and group chat
			if msg.GroupName == "" {
				//This is 1-to-1
				m = models.Message{
					ID:             primitive.NewObjectID(),
					ConversationID: convo.ConversationID,
					SenderUserName: currentUser,
					Content:        msg.MessageContent,
					CreatedAt:      time.Now(),
				}
			} else {
				//This is group chat

				//Check whether this is a newly created group chat
				if announceMessage { //this will hold true if the convo is newly created
					m = models.Message{
						ID:             primitive.NewObjectID(),
						ConversationID: convo.ConversationID,
						SenderUserName: "Announce", //This will be an announcement that this group chat is created to all the participants
						Content:        msg.MessageContent,
						CreatedAt:      time.Now(),
					}
				} else {
					//This is already an existing group chat
					m = models.Message{
						ID:             primitive.NewObjectID(),
						ConversationID: convo.ConversationID,
						SenderUserName: currentUser,
						Content:        msg.MessageContent,
						CreatedAt:      time.Now(),
					}
				}
			}
			_, insertErr := messageCollection.InsertOne(context.Background(), m)
			if insertErr != nil {
				log.Println("Failed to insert message:", insertErr.Error())
				continue
			}

			//Send back payload to all the participants of this convo
			response, err := json.Marshal(map[string]interface{}{
				"type":    "message",
				"convo":   convo,
				"message": m,
			})
			if err != nil {
				log.Println("Failed to marshal message:", err)
				continue
			}

			for _, p := range convo.Participants {
				//Check if the participant is online
				toClient, online := onlineClients[p]
				if online {
					//Immediately send the payload
					toClient.Receive <- response
				}
			}
			continue
		}
	}
}

/*-----------------------------------------------------------------------------------------------*/

func unique(input []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(input))

	for _, v := range input {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}

	return result
}

/*-----------------------------------------------------------------------------------------------*/



func (c *Client) LoadAllMessage() {
	//Get all related convo to this client
	convoCollection := config.OpenCollection("conversation")
	filter := bson.M{
		"participants": c.Username,
	}
	ConvoCursor, err := convoCollection.Find(context.Background(), filter)
	
	if err != nil {
		log.Println("Failed to fetch convo:", err)
		return
	}

	var convos []models.Conversation

	if err := ConvoCursor.All(context.Background(), &convos); err != nil {
		log.Println(err.Error())
		return
	}
	defer ConvoCursor.Close(context.Background())

	messagesByConversationID := make(map[string][]models.Message)

	messageCollection := config.OpenCollection("message")
	MessageCursor, err := messageCollection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Println("Error retrieving all the message documents")
		return
	}

	defer MessageCursor.Close(context.Background())

	//Group ALL the messages by its unique ConversationID
	for MessageCursor.Next(context.Background()) {
		var message models.Message
		if err := MessageCursor.Decode(&message); err != nil {
			log.Println(err.Error())
			return
		}
		messagesByConversationID[message.ConversationID] = append(messagesByConversationID[message.ConversationID], message)
	}

	//Sort Messages with oldest message -> latest message for each ConversationID
	for _, msgs := range messagesByConversationID {
		sort.Slice(msgs, func(i, j int) bool {
			return msgs[i].CreatedAt.Before(msgs[i].CreatedAt)
		})
	}

	type ConvoAndMessagesItem struct {
		Conversation models.Conversation `json:"conversation"`
		Messages     []models.Message    `json:"Messages"`
	}

	var convoAndMessages []ConvoAndMessagesItem
	//Filter the Grouped messages by the ConversationID that this current user is inside only
	for _, c := range convos {
		if msgs, ok := messagesByConversationID[c.ConversationID]; ok {
			convoAndMessages = append(convoAndMessages, ConvoAndMessagesItem{
				Conversation: c,
				Messages:     msgs,
			})
		}
	}

	//Send back all the Conversations + related Messages for each Conversation back to the client
	payload, err := json.Marshal(map[string]interface{}{
		"type":             "allMessages",
		"convoAndMessages": convoAndMessages,
	})
	if err != nil {
		log.Println("Failed to marshal message:", err)
		return
	}
	c.Receive <- payload
}

/*-----------------------------------------------------------------------------------------------*/

func (c *Client) LoadAllFriends() {
	friendCollection := config.OpenCollection("friend")
	cursor, err := friendCollection.Find(context.Background(), bson.M{
		"username": c.Username,
	})

	if err != nil {
		log.Println("Failed to fetch pending requests:", err)
		return
	}

	defer cursor.Close(context.Background())

	var friends []bson.M

	if err := cursor.All(context.Background(), &friends); err != nil {
		log.Println(err.Error())
		return
	}

	response, err := json.Marshal(map[string]interface{}{
		"type":    "friend_list_update",
		"friends": friends,
	})
	c.Receive <- response
}

func (c *Client) ReceivePendingFriendRequest() {
	requestCollection := config.OpenCollection("request")
	cursor, err := requestCollection.Find(context.Background(), bson.M{
		"to":     c.Username,
		"status": "pending",
	})
	if err != nil {
		log.Println("Failed to fetch pending requests:", err)
		return
	}

	defer cursor.Close(context.Background())

	var requests []bson.M

	if err := cursor.All(context.Background(), &requests); err != nil {
		log.Println(err.Error())
		return
	}

	for _, r := range requests {
		jsMsg, err := json.Marshal(OutgoingMessage{
			From: r["from"].(string),
			To:   c.Username,
			Type: "friend_request",
		})
		if err != nil {
			log.Println("Failed to marshal message:", err)
			continue
		}
		c.Receive <- jsMsg
	}
}

func (c *Client) Write() {
	for msg := range c.Receive {
		if err := c.Socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

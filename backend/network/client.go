package network

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/shjung-dev/ChatApplication/backend/config"
	"go.mongodb.org/mongo-driver/bson"
)

type Client struct {
	Socket   *websocket.Conn
	Receive  chan []byte
	Room     *Room
	Username string
}

type WSMessage struct {
	Type    string `json:"type"`
	To      string `json:"to"`
	Content string `json:"content"`
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

		if msg.Type == "friend_request" {
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
			toClient , online := onlineClients[msg.To]
			if online {
				//Client is online too so immediately send over websocket
				jsMsg , err := json.Marshal(OutgoingMessage{
					From: from,
					To: msg.To,
					Type: msg.Type,
				})
				if err != nil {
					log.Println("Failed to marshal message:", err)
					continue
				}
				toClient.Receive <- jsMsg
				continue
			}
		}
	}
}

func (c *Client) ReceivePendingFriendRequest(){
	requestCollection := config.OpenCollection("request")
	cursor, err := requestCollection.Find(context.Background(), bson.M{
        "to": c.Username,
        "status": "pending",
    })
	if err != nil {
        log.Println("Failed to fetch pending requests:", err)
        return
    }

	var requests []bson.M

	if err:=cursor.All(context.Background() , &requests); err != nil {
		log.Println(err.Error())
		return
	}
	
	for _ , r := range requests{
		jsMsg , err := json.Marshal(OutgoingMessage{
			From: r["from"].(string),
			To: c.Username,
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

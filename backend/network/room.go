package network

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

//Global Online Client tracker variable
var onlineClients = make(map[string]*Client)
var onlineMu sync.Mutex

type Room struct {
	Name      string
	Clients   map[*Client]bool
	Join      chan *Client
	Leave     chan *Client
	BroadCast chan OutgoingMessage
}

func NewRoom(name string) *Room {
	return &Room{
		Name:      name,
		Clients:   make(map[*Client]bool),
		Join:      make(chan *Client),
		Leave:     make(chan *Client),
		BroadCast: make(chan OutgoingMessage),
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.Join:
			r.Clients[client] = true
			log.Println(client.Username + " has joined the room")
		case client := <-r.Leave:
			delete(r.Clients, client)

			if len(r.Clients) == 0 {
				mu.Lock()
				delete(rooms, r.Name)
				log.Println("Room", r.Name, "deleted due to no clients")
				mu.Unlock()
			}
		case msg := <-r.BroadCast:
			jsMsg, err := json.Marshal(msg)
			if err != nil {
				log.Println("Failed to marshal message:", err)
				continue
			}
			for c := range r.Clients {
				c.Receive <- jsMsg
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var rooms = make(map[string]*Room)
var mu sync.Mutex

func GetRoom(name string) *Room {
	mu.Lock()
	defer mu.Unlock()

	if r, ok := rooms[name]; ok {
		return r
	}

	r := NewRoom(name)
	rooms[name] = r
	go r.Run()
	return r
}

func GenerateRoomName(username string) string {
	return "Room_" + username
}

func (r *Room) ServeHttp(w http.ResponseWriter, req *http.Request) {
	username := req.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		Socket:  socket,
		Receive: make(chan []byte, messageBufferSize),
		Room:    r,
		Username: username,
	}

	onlineMu.Lock()
	onlineClients[client.Username] = client
	onlineMu.Unlock()

	r.Join <- client
	defer func(){
		r.Leave <- client
		onlineMu.Lock()
		delete(onlineClients , username)
		onlineMu.Unlock()
	}()
	
	go client.Write()

	client.ReceivePendingFriendRequest()

	client.Read()
}

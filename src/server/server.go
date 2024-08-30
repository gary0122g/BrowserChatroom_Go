package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Server struct {
	clients          map[*Client]*websocket.Conn
	rooms            map[string][]*Client
	randomMatchQueue []*Client
}

type Client struct {
	conn *websocket.Conn
	room []string
	id   string
}

type ControlMessage struct {
	Type             string `json:"type"`
	Room             string `json:"room"`
	BroadcastMessage string `json:"broadcast_message"`
	RandomMatch      bool   `json:"random_match"`
}

func NewWsServer() *Server {
	return &Server{
		clients: make(map[*Client]*websocket.Conn),
		rooms:   make(map[string][]*Client),
	}
}

func handler(w http.ResponseWriter, r *http.Request, s *Server) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{conn: conn, id: uuid.New().String()}
	fmt.Println(client.id)
	s.clients[client] = conn
	fmt.Println(s.clients)
	rooms := GetRooms(s)
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Welcome to chat room. There are %d rooms available: %v", len(rooms), rooms)))
	go handleControlMessage(client, s)
}

func handleControlMessage(client *Client, s *Server) {
	defer client.conn.Close()
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var controlMsg ControlMessage
		err = json.Unmarshal(message, &controlMsg)
		if err != nil {
			log.Println("Invalid message format:", err)
			continue
		}
		switch controlMsg.Type {
		case "create":
			createRoom(client, s, controlMsg.Room)
		case "join":
			joinRoom(client, s, controlMsg.Room)
		case "broadcast":
			broadcastMessage(s, client, controlMsg.Room, controlMsg.BroadcastMessage)
		case "getRooms":
			sendRoomList(client, s)
		case "randomMatch":
			randonMatch(client, s)
		default:
			log.Println("Unknown control message type:", controlMsg.Type)
		}
	}
}

func broadcastMessage(s *Server, sender *Client, roomName string, message string) {
	fullMessage := fmt.Sprintf("%s: %s", sender.id, message)
	for _, client := range s.rooms[roomName] {
		if err := client.conn.WriteMessage(websocket.TextMessage, []byte(fullMessage)); err != nil {
			log.Println("Write error:", err)
		}
	}
}

func (s *Server) Run() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, s)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

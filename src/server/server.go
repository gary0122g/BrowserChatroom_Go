package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Server struct {
	clients          map[*Client]*websocket.Conn
	rooms            map[string][]*Client
	randomMatchQueue []*Client
	mu               sync.Mutex
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
}

func NewWsServer() *Server {
	return &Server{
		clients: make(map[*Client]*websocket.Conn),
		rooms:   make(map[string][]*Client),
	}
}

func (s *Server) Run() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, s)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
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
	conn.WriteMessage(websocket.TextMessage, []byte(client.id))
	s.mu.Lock()
	s.clients[client] = conn
	s.mu.Unlock()
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
			randomMatch(client, s)
		case "disconnect":
			disconnect(client, s)

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

func disconnect(client *Client, s *Server) {
	s.mu.Lock()
	delete(s.clients, client)
	for i, c := range s.randomMatchQueue {
		if c == client {
			s.randomMatchQueue = append(s.randomMatchQueue[:i], s.randomMatchQueue[i+1:]...)
			fmt.Println("remove from random match queue")
		}
	}

	for _, room := range client.room {
		for i, c := range s.rooms[room] {
			if c == client {
				s.rooms[room] = append(s.rooms[room][:i], s.rooms[room][i+1:]...)
				if len(s.rooms[room]) == 0 {
					delete(s.rooms, room)
					fmt.Println("delete room")
					fmt.Println(s.rooms)
				}
			}
		}
	}
	s.mu.Unlock()
}

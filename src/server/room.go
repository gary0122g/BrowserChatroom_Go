package server

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

func createRoom(client *Client, s *Server, roomName string) {
	s.rooms[roomName] = append(s.rooms[roomName], client)
	client.room = append(client.room, roomName)
	client.conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Room created: %s", roomName)))
	fmt.Println(s.rooms)
}

func joinRoom(client *Client, s *Server, roomName string) {
	s.rooms[roomName] = append(s.rooms[roomName], client)
	client.room = append(client.room, roomName)
	broadcastMessage(s, client, roomName, "New user joined the room")
	fmt.Println(s.rooms)
}

func GetRooms(s *Server) []string {
	rooms := make([]string, 0)
	for room := range s.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
func sendRoomList(client *Client, s *Server) {
	rooms := GetRooms(s)
	roomsJSON, err := json.Marshal(rooms)
	if err != nil {
		log.Println("Error marshalling room list:", err)
		return
	}

	message := fmt.Sprintf("ROOMLIST: %s", string(roomsJSON))
	if err := client.conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		log.Println("Write error:", err)
	}
}

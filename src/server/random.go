package server

import (
	"fmt"
	"math/rand"

	"github.com/gorilla/websocket"
)

func randonMatch(client *Client, s *Server) {
	s.randomMatchQueue = append(s.randomMatchQueue, client)
	if len(s.randomMatchQueue) >= 2 {
		match := s.randomMatchQueue[:2]
		// Generate a random room name for the match
		randomRoomName := fmt.Sprintf("random_%s", generateRandomString(8))

		// Create the random room
		s.rooms[randomRoomName] = []*Client{match[0], match[1]}
		fmt.Println(s.rooms[randomRoomName])
		match[0].room = append(match[0].room, randomRoomName)
		match[1].room = append(match[1].room, randomRoomName)
		s.randomMatchQueue = s.randomMatchQueue[2:]

		// Send the room name to both matched clients
		matchMessage := fmt.Sprintf("RANDOM_MATCH: %s", randomRoomName)
		match[0].conn.WriteMessage(websocket.TextMessage, []byte(matchMessage))
		match[1].conn.WriteMessage(websocket.TextMessage, []byte(matchMessage))

		// Broadcast the match information to both clients
		broadcastMessage(s, match[0], randomRoomName, fmt.Sprintf("You have been matched with %s", match[1].id))
		broadcastMessage(s, match[1], randomRoomName, fmt.Sprintf("You have been matched with %s", match[0].id))
	} else {
		// If there's only one client in the queue, inform them they're waiting
		client.conn.WriteMessage(websocket.TextMessage, []byte("Waiting for a match..."))
	}
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

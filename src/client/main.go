package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
)

type ControlMessage struct {
	Type             string `json:"type"`
	Room             string `json:"room"`
	BroadcastMessage string `json:"broadcast_message"`
}

func main() {
	var currentRoom string
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)

	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Println(string(message))
		}
	}()
	go func() {
		for {
			if currentRoom == "" {
				fmt.Print("Enter command (create/join): ")
				command, _ := reader.ReadString('\n')
				command = strings.TrimSpace(command)
				var msg ControlMessage

				switch command {
				case "create":

					fmt.Print("Enter room name to create: ")
					roomName, _ := reader.ReadString('\n')
					currentRoom = strings.TrimSpace(roomName)
					msg.Type = "create"
					msg.Room = currentRoom

				case "join":
					fmt.Print("Enter room name to join: ")
					roomName, _ := reader.ReadString('\n')
					currentRoom = strings.TrimSpace(roomName)
					msg.Type = "join"
					msg.Room = currentRoom

				default:
					fmt.Println("Invalid command. Use 'create' or 'join'.")
					continue
				}

				err = c.WriteJSON(msg)
				if err != nil {
					log.Println("write:", err)
					return
				}
			} else {
				fmt.Print("Enter message to broadcast: ")
				message, _ := reader.ReadString('\n')
				message = strings.TrimSpace(message)
				var msg ControlMessage
				msg.Type = "broadcast"
				msg.Room = currentRoom
				msg.BroadcastMessage = message

				err = c.WriteJSON(msg)
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
		}
	}()

	<-interrupt
	log.Println("Closing connection...")
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
	}
}

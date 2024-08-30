package main

import (
	"chatroom/src/server"
)

func main() {
	server := server.NewWsServer()
	server.Run()
}

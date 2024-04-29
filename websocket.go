package main

import "github.com/gorilla/websocket"

var clients map[string][]*websocket.Conn

func init() {
	clients = make(map[string][]*websocket.Conn)
}

func sendMsg(clientendpoint string, msg string) {
	for _, conn := range clients[clientendpoint] {
		conn.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}

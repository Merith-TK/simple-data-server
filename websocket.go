package main

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients      map[string][]*websocket.Conn
	clientsMutex sync.RWMutex
)

func init() {
	clients = make(map[string][]*websocket.Conn)
}

// sendMsg sends a message to all connected clients for a specific endpoint
func sendMsg(clientendpoint string, msg string) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	// Create a copy of the slice to avoid issues with concurrent modification
	var connsCopy []*websocket.Conn
	if conns, exists := clients[clientendpoint]; exists {
		connsCopy = make([]*websocket.Conn, len(conns))
		copy(connsCopy, conns)
	}

	// Send to all connections, removing any that are closed
	var validConns []*websocket.Conn
	for _, conn := range connsCopy {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Printf("Error sending message to client: %v", err)
			conn.Close()
		} else {
			validConns = append(validConns, conn)
		}
	}

	// Update the clients map with only valid connections
	clientsMutex.Lock()
	clients[clientendpoint] = validConns
	clientsMutex.Unlock()
}

// addClient adds a new websocket connection to the clients map
func addClient(endpoint string, conn *websocket.Conn) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	clients[endpoint] = append(clients[endpoint], conn)
}

// removeClient removes a websocket connection from the clients map
func removeClient(endpoint string, conn *websocket.Conn) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if conns, exists := clients[endpoint]; exists {
		for i, c := range conns {
			if c == conn {
				clients[endpoint] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
	}
}

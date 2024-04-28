package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	router := mux.NewRouter()

	// handle /api/<datapath>/<uuid>/ws
	router.HandleFunc("/api/{datapath}/{uuid}", handleUUID)
	router.HandleFunc("/api/{datapath}/{uuid}/ws", handleWebSocket)
	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	datapath := vars["datapath"]
	uuid := vars["uuid"]

	verifyFile(datapath, uuid)
	data, err := readEntireData(datapath, uuid)
	if err != nil {
		log.Println("Failed to read data:", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)

}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Println("Received websocket connection")
	//  GET mux vars
	vars := mux.Vars(r)
	datapath := vars["datapath"]
	uuid := vars["uuid"]
	log.Printf("Received websocket connection for %s/%s\n", datapath, uuid)
	verifyFile(datapath, uuid)
	// ping pong
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()
	for {
		// client can send one of two commands
		// 1. set key value
		// 2. get key
		// 3. del key
		// 4. ping
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Failed to read message:", err)
			break
		}
		log.Printf("Received message: %s\n", message)
		msg := string(message)
		// split message into words
		words := strings.Split(msg, " ")
		command := words[0]

		switch command {
		case "set":
			if len(words) < 3 {
				log.Println("Invalid message")
				continue
			}
			key := words[1]
			value := words[2]
			err := writeData(datapath, uuid, key, value)
			if err != nil {
				log.Println("Failed to write data:", err)
				continue
			}
			log.Printf("Set %s to %s\n", key, value)
		case "get":
			data, err := readData(datapath, uuid)
			if err != nil {
				log.Println("Failed to read data:", err)
				continue
			}
			if len(words) < 2 {
				log.Println("Invalid message")
				continue
			}
			key := words[1]
			value, ok := data.Data[key]
			if !ok {
				log.Printf("Key %s not found\n", key)
				continue
			}
			log.Printf("Get %s: %s\n", key, value)
			err = conn.WriteMessage(websocket.TextMessage, []byte(value))
			if err != nil {
				log.Println("Failed to write message:", err)
				break
			}
		case "del":
			if len(words) < 2 {
				log.Println("Invalid message")
				continue
			}
			key := words[1]
			err := deleteData(datapath, uuid, key)
			if err != nil {
				log.Println("Failed to delete data:", err)
				continue
			}
			log.Printf("Deleted %s\n", key)
		case "ping":
			err := conn.WriteMessage(websocket.TextMessage, []byte("pong"))
			if err != nil {
				log.Println("Failed to write message:", err)
				break
			}
		default:
			log.Println("Invalid command")

		}

	}
}

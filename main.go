package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	"crypto/sha256"
	"encoding/hex"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

/*
	FUNCTIONS:
	* userHash
		* get the user and password from the request and hash them to create a unique key
		* return true and the user hash if the user and password are found
		* return false and an empty string if the user and password are not found
	* getData (from data.go)
		* get the userhash, object, table, and key from the request
		* if userhash is empty, return key from ./data/default/object/table.json
		* if userhash is not  empty, return key from ./data/userhash/object/table.json
		* if key is empty, return all keys from ./data/userhash/object/table.json or ./data/default/object/table.json if userhash is empty
	* setData  (from data.go)
		* get the userhash, object, table, key, and data from the request
		* if userhash is empty, write data to ./data/default/object/table.json
		* if userhash is not empty, write data to ./data/userhash/object/table.json
		* if key is empty, do nothing
	* deleteData  (from data.go)
		* get the userhash, object, table, and key from the request
		* if userhash is empty, delete key from ./data/default/object/table.json
		* if userhash is not empty, delete key from ./data/userhash/object/table.json

	WEB SOCKET FUNCTIONS:
	* cmd: set key value
		* sets the key to the value in the current object/table
	* cmd: get key
		* gets the value of the key in the current object/table
	* cmd: del key
		* deletes the key from the current object/table


	USAGE:
	* GET /api/object/table/key
		* get the value of the key from the current object/table
		* if userhash is empty, get the value from ./data/default/object/table.json
		* if userhash is not empty, get the value from ./data/userhash/object/table.json
	* GET /api/object/table
		* get all keys from the current object/table
		* if userhash is empty, get all keys from ./data/default/object/table.json
		* if userhash is not empty, get all keys from ./data/userhash/object/table.json
	* POST /api/object/table/key
		* set the value of the key in the current object/table
		* if userhash is empty, set the value in ./data/default/object/table.json
		* if userhash is not empty, set the value in ./data/userhash/object/table.json
	* DELETE /api/object/table/key
		* delete the key from the current object/table
		* if userhash is empty, delete the key from ./data/default/object/table.json
		* if userhash is not empty, delete the key from ./data/userhash/object/table.json
	* GET /api/object/table/ws
		* upgrade the connection to a websocket
		* send and receive commands to set, get, and delete keys

*/

func userHash(w http.ResponseWriter, r *http.Request) (bool, string) {
	// get the user and password from the request
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false, "default"
	}

	// hash the user and password to create a unique key
	data := []byte(user + pass)
	hash := sha256.Sum256(data)
	return true, hex.EncodeToString(hash[:])
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/{object}/{table}/ws", handleWS)
	router.HandleFunc("/api/{object}/{table}/{key}", handleAPI).Methods("GET", "POST", "DELETE")
	router.HandleFunc("/api/{object}/{table}", handleAPI).Methods("GET")

	http.Handle("/", router)
	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func handleWS(w http.ResponseWriter, r *http.Request) {
	log.Println("Websocket connection")

	// get the userhash, object, and table from the request
	_, userhash := userHash(w, r)

	vars := mux.Vars(r)
	object := vars["object"]
	table := vars["table"]

	// upgrade the connection to a websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Received websocket connection for %s/%s\n", object, table)

	// add client to the list of clients
	endpointString := userhash + "/" + object + "/" + table
	clients[endpointString] = append(clients[endpointString], conn)

	// send the userhash to the client
	conn.WriteMessage(websocket.TextMessage, []byte("uuid: "+userhash))
	for {
		// read the command from the websocket
		_, cmd, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Client disconnected: %v\n", err)
			// find the client in the list of clients and remove it
			for i, c := range clients[endpointString] {
				if c == conn {
					clients[endpointString] = append(clients[endpointString][:i], clients[endpointString][i+1:]...)
					break
				}
			}

			return
		}

		// get the key and value from the command
		parts := strings.Split(string(cmd), " ")
		key := parts[1]
		cmd1 := parts[0]
		cmdstr := strings.ToLower(string(cmd1))
		// handle the command
		switch cmdstr {
		case "set":
			value := strings.Join(parts[2:], " ")
			if len(parts) == 2 {
				value = ""
			}
			err := setData(userhash, object, table, key, value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// conn.WriteMessage(websocket.TextMessage, []byte("SET: "+key))
			sendMsg(endpointString, "UPDATE: "+key+": "+value)
		case "get":
			data, err := getData(userhash, object, table, key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			conn.WriteMessage(websocket.TextMessage, []byte(data))
		case "del":
			err := deleteData(userhash, object, table, key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			conn.WriteMessage(websocket.TextMessage, []byte("Deleted "+key))
		case "exit":
			conn.WriteMessage(websocket.TextMessage, []byte("Exiting"))
			conn.Close()
		default:
			conn.WriteMessage(websocket.TextMessage, []byte("Unknown command: "+parts[0]))
		}
	}
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	// get the userhash, object, table, and key from the request
	_, userhash := userHash(w, r)

	vars := mux.Vars(r)
	object := vars["object"]
	table := vars["table"]
	key := vars["key"]

	switch r.Method {
	case "GET":
		data, err := getData(userhash, object, table, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if key == "" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.Write([]byte(data))
	case "POST":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = setData(userhash, object, table, key, string(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("SET: " + key))
	case "DELETE":
		err := deleteData(userhash, object, table, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Deleted " + key))
	}
}

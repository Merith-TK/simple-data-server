package main

import (
	"encoding/json"
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
		return true // TODO: Make this configurable for production
	},
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Helper function to send JSON error responses
func sendErrorResponse(w http.ResponseWriter, statusCode int, errMsg string, details ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: errMsg,
	}
	if len(details) > 0 {
		response.Message = details[0]
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function to send success response
func sendSuccessResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "success", "message": message}
	json.NewEncoder(w).Encode(response)
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
	config := loadConfig()

	router := mux.NewRouter()
	router.HandleFunc("/api/{object}/{table}/ws", handleWS)
	router.HandleFunc("/api/{object}/{table}/{key}", handleAPI).Methods("GET", "POST", "DELETE")
	router.HandleFunc("/api/{object}/{table}", handleAPI).Methods("GET")

	// Add a health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		sendSuccessResponse(w, "Server is healthy")
	})

	// Add middleware
	handler := LoggingMiddleware(CORSMiddleware(router))

	http.Handle("/", handler)
	log.Printf("Server started on port %s", config.Port)
	log.Printf("Data directory: %s", config.DataDir)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
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
		log.Printf("Failed to upgrade connection: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to upgrade connection", err.Error())
		return
	}
	defer conn.Close()

	log.Printf("Received websocket connection for %s/%s\n", object, table)

	// add client to the list of clients
	endpointString := userhash + "/" + object + "/" + table
	addClient(endpointString, conn)
	defer removeClient(endpointString, conn)

	// send the userhash to the client
	if err := conn.WriteMessage(websocket.TextMessage, []byte("uuid: "+userhash)); err != nil {
		log.Printf("Error sending initial message: %v", err)
		return
	}

	for {
		// read the command from the websocket
		_, cmd, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Client disconnected: %v\n", err)
			return
		}

		// get the key and value from the command
		parts := strings.Split(string(cmd), " ")
		if len(parts) < 2 {
			conn.WriteMessage(websocket.TextMessage, []byte("Error: Invalid command format"))
			continue
		}

		key := parts[1]
		cmd1 := parts[0]
		cmdstr := strings.ToLower(string(cmd1))

		// handle the command
		switch cmdstr {
		case "set":
			value := ""
			if len(parts) > 2 {
				value = strings.Join(parts[2:], " ")
			}

			err := setData(userhash, object, table, key, value)
			if err != nil {
				log.Printf("Error setting data via websocket: %v", err)
				conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
				continue
			}
			sendMsg(endpointString, "UPDATE: "+key+": "+value)

		case "get":
			data, err := getData(userhash, object, table, key)
			if err != nil {
				log.Printf("Error getting data via websocket: %v", err)
				conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
				continue
			}
			conn.WriteMessage(websocket.TextMessage, []byte(key+": "+data))

		case "del":
			err := deleteData(userhash, object, table, key)
			if err != nil {
				log.Printf("Error deleting data via websocket: %v", err)
				conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
				continue
			}
			conn.WriteMessage(websocket.TextMessage, []byte("Deleted: "+key))
			sendMsg(endpointString, "DELETED: "+key)

		case "exit":
			conn.WriteMessage(websocket.TextMessage, []byte("Goodbye"))
			return

		default:
			conn.WriteMessage(websocket.TextMessage, []byte("Error: Unknown command: "+parts[0]))
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
			log.Printf("Error getting data: %v", err)
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve data", err.Error())
			return
		}
		if key == "" {
			w.Header().Set("Content-Type", "application/json")
		} else {
			w.Header().Set("Content-Type", "text/plain")
		}
		w.Write([]byte(data))

	case "POST":
		// Validate that key is provided for POST
		if key == "" {
			sendErrorResponse(w, http.StatusBadRequest, "Key is required for POST requests")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			sendErrorResponse(w, http.StatusBadRequest, "Failed to read request body", err.Error())
			return
		}

		err = setData(userhash, object, table, key, string(body))
		if err != nil {
			log.Printf("Error setting data: %v", err)
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to set data", err.Error())
			return
		}
		sendSuccessResponse(w, "Data set successfully for key: "+key)

	case "DELETE":
		// Validate that key is provided for DELETE
		if key == "" {
			sendErrorResponse(w, http.StatusBadRequest, "Key is required for DELETE requests")
			return
		}

		err := deleteData(userhash, object, table, key)
		if err != nil {
			log.Printf("Error deleting data: %v", err)
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to delete data", err.Error())
			return
		}
		sendSuccessResponse(w, "Data deleted successfully for key: "+key)
	}
}

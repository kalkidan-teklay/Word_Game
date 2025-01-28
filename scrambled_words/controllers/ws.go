package controllers

import (
	"log"
	"net/http"
	"scrambled_words/shared"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Register client
	shared.Mu.Lock()
	shared.Clients[conn] = true
	shared.Mu.Unlock()
	log.Printf("New WebSocket connection. Total clients: %d\n", len(shared.Clients))

	// Listen for messages and respond (optional)
	for {
		var msg shared.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}
		// Example: Broadcast the received message
		shared.Broadcast <- msg
	}

	// Unregister client
	shared.Mu.Lock()
	delete(shared.Clients, conn)
	shared.Mu.Unlock()
	log.Printf("WebSocket disconnected. Total clients: %d\n", len(shared.Clients))
}

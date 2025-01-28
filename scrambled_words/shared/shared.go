package shared

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

var (
	Clients   = make(map[*websocket.Conn]bool) // Track connected clients
	Mu        sync.Mutex                       // Synchronize access to Clients
	Broadcast = make(chan Message)             // Broadcast channel
)

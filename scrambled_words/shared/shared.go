package shared

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
type Player struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

var (
	Clients   = make(map[*websocket.Conn]bool) // Track connected clients
	Players   = make(map[*websocket.Conn]Player)
	Mu        sync.Mutex           // Synchronize access to Clients
	Broadcast = make(chan Message) // Broadcast channel
)

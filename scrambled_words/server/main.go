package main

import (
	"log"
	"net/http"
	"scrambled_words/controllers"
	"scrambled_words/db"
	"scrambled_words/routes"
	"scrambled_words/shared"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) // Track connected clients

func broadcastMessages() {
	for {
		msg := <-shared.Broadcast
		for client := range shared.Clients {
			err := client.WriteJSON(msg)
			log.Printf("Broadcasting message: %+v\n", msg)
			log.Printf("Connected clients: %d\n", len(clients))
			if err != nil {
				log.Println("WebSocket write error:", err)
				client.Close()
				shared.Mu.Lock()
				delete(shared.Clients, client)
				shared.Mu.Unlock()
			}
		}
	}
}

func main() {
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	go broadcastMessages()
	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	routes.RegisterRoutes(r)
	db.InitRedis()
	controllers.LoadGameState()

	// Start the server
	log.Println("Server is running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

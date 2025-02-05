package controllers

import (
	"context"
	"fmt"
	"net/http"
	"third_server/db"
	"third_server/models"

	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Leaderboard endpoint
func GetLeaderboard(c *gin.Context) {
	collection := db.GetCollection("scrambled_words", "users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch users sorted by wins in descending order
	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"wins": -1}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}
	defer cursor.Close(ctx)

	var leaderboard []map[string]interface{}
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			fmt.Println("Error decoding user:", err)
			continue
		}
		leaderboard = append(leaderboard, map[string]interface{}{
			"username": user.Username,
			"wins":     user.Wins,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
	})
}

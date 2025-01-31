package controllers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"scrambled_words/db"
	"scrambled_words/models"
	"scrambled_words/shared"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

var gameState = models.GameState{}
var mu sync.Mutex

func init() {
	rand.Seed(time.Now().UnixNano())
}

func LoadGameState() {
	mu.Lock()
	defer mu.Unlock()

	storedState, err := db.LoadGameState()
	if err == nil {
		gameState = *storedState
	} else {
		gameState = models.GameState{}
	}
}

// Generate a random word and shuffle it
func generateWord() string {
	words := []string{"apple", "banana", "cherry", "grape", "orange"}
	word := words[rand.Intn(len(words))]
	shuffled := shuffleString(word)
	gameState.Word = word
	gameState.Shuffled = shuffled
	return word
}

func shuffleString(s string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	chars := strings.Split(s, "")
	r.Shuffle(len(chars), func(i, j int) { chars[i], chars[j] = chars[j], chars[i] })
	return strings.Join(chars, "")
}

// Join the game
func JoinGame(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	var player models.Player
	if err := c.ShouldBindJSON(&player); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	player.Score = 0
	gameState.Players = append(gameState.Players, player)

	playerNames := []string{}
	for _, p := range gameState.Players {
		playerNames = append(playerNames, p.Name)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Player joined",
		"player":       player,
		"joined_users": playerNames, // List of player names
	})
	db.SaveGameState(&gameState)
}

func CheckMenu(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	var request struct {
		PlayerID string `json:"player_id"`
		Type     string `json:"type"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collection := db.GetCollection("scrambled_words", "users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(request.PlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Player ID"})
		return
	}

	if request.Type == "new" {
		// Reset only this player's score
		_, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"score": 0}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset score"})
			return
		}

		for i := range gameState.Players {
			if gameState.Players[i].ID == request.PlayerID {
				gameState.Players[i].Score = 0
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
func StartGame(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	for i := range gameState.Players {
		gameState.Players[i].Score = 0
	}

	if !gameState.Started {
		generateWord()
		gameState.Started = true
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,

		"word": gameState.Shuffled,
	})
	db.SaveGameState(&gameState)
}

// Submit answer
func SubmitAnswer(c *gin.Context) {
	log.Println("Acquiring lock in SubmitAnswer()")
	shared.Mu.Lock()
	defer shared.Mu.Unlock()
	log.Println("Releasing lock in SubmitAnswer()")

	var request struct {
		PlayerID string `json:"player_id"`
		Guess    string `json:"guess"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get the player from the database
	collection := db.GetCollection("scrambled_words", "users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var player models.Player
	objID, err := primitive.ObjectIDFromHex(request.PlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Player ID"})
		return
	}
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&player)

	// Check the answer
	if strings.ToLower(request.Guess) == strings.ToLower(gameState.Word) {
		// Update the player's score in the database
		player.Score++
		_, err := collection.UpdateOne(
			ctx,
			bson.M{"_id": objID},
			bson.M{"$set": bson.M{"score": player.Score}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update score"})
			return
		}

		for conn, p := range shared.Players {
			if p.Name == player.Name {
				p.Score = player.Score
				shared.Players[conn] = p
				break
			}
		}

		go broadcastPlayerList()

		if player.Score == 3 {
			gameState.Winner = &player
			gameState.Started = false
			log.Println("Broadcasting game over for winner:", player.Name)

			shared.Broadcast <- shared.Message{
				Type: "game_over",
				Payload: gin.H{
					"winner":  player.Name,
					"message": fmt.Sprintf("%s won the game!", player.Name),
				},
			}

			_, err := collection.UpdateOne(
				ctx,
				bson.M{"_id": objID},
				bson.M{"$inc": bson.M{"wins": 1}}, // Increment the "wins" field
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wins"})
				return
			}
		} else {
			generateWord()

		}

		c.JSON(http.StatusOK, gin.H{
			"message":  fmt.Sprintf("%s won the game!", player.Name),
			"correct":  true,
			"player":   player,
			"new_word": gameState.Shuffled,
			"scores":   getScores(), // Include scores in the response
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Incorrect, try again!",
			"correct": false,
			"scores":  getScores(), // Include scores in the response
		})
	}

	db.SaveGameState(&gameState)

}

// Helper function to get scores of all online players
func getScores() []map[string]interface{} {
	mu.Lock()
	defer mu.Unlock()

	var scores []map[string]interface{}

	// Iterate through the gameState.Players to get online players
	for _, player := range gameState.Players {
		scores = append(scores, map[string]interface{}{
			"name":   player.Name,
			"points": player.Score,
		})
	}
	if len(scores) == 0 {
		return []map[string]interface{}{}
	}

	return scores
}

func getPlayerByID(id string) *models.Player {
	for i, p := range gameState.Players {
		if p.ID == id {
			return &gameState.Players[i]
		}
	}
	return nil
}

func LeaveGame(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	var request struct {
		PlayerID string `json:"player_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Remove the player from the gameState.Players slice
	for i, player := range gameState.Players {
		if player.ID == request.PlayerID {
			gameState.Players = append(gameState.Players[:i], gameState.Players[i+1:]...)
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Player left the game",
	})
	db.SaveGameState(&gameState)
}

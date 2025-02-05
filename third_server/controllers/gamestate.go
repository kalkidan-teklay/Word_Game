package controllers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"third_server/db"
	"third_server/models"
	"third_server/shared"
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
	shared.Mu.Lock()
	defer shared.Mu.Unlock()

	// Extract the player ID from the request
	var request struct {
		PlayerID string `json:"player_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Convert the PlayerID string to primitive.ObjectID
	playerID, err := primitive.ObjectIDFromHex(request.PlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Player ID"})
		return
	}

	// Get the users collection (not players, because we are modifying user data)
	userCollection := db.GetCollection("scrambled_words", "users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the player in the "users" collection
	var targetPlayer shared.Player
	err = userCollection.FindOne(ctx, bson.M{"_id": playerID}).Decode(&targetPlayer)
	if err != nil {
		log.Println("Player not found in users collection:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	// Generate a new word for the game
	newWord := generateWord()

	// Update the player's word in the database
	_, err = userCollection.UpdateOne(
		ctx,
		bson.M{"_id": playerID},
		bson.M{"$set": bson.M{"word": newWord}},
	)
	if err != nil {
		log.Printf("Failed to update player word: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update player word"})
		return
	}

	// Update the in-memory player state if the player is currently connected
	for conn, player := range shared.Players {
		if player.ID == playerID {
			player.Word = newWord
			shared.Players[conn] = player
			break
		}
	}

	// Broadcast the new word only to the player who called the function
	message := shared.Message{
		Type:    "start_game",
		Payload: gin.H{"word": newWord},
	}

	// Find the connection for the target player and send the message
	for conn, player := range shared.Players {
		if player.ID == playerID {
			err := conn.WriteJSON(message)
			if err != nil {
				log.Println("Error sending message to client:", err)
				conn.Close()
				delete(shared.Clients, conn)
				delete(shared.Players, conn)
			}
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"word":    newWord,
	})
}

func SubmitAnswer(c *gin.Context) {
	log.Println("Acquiring lock in SubmitAnswer()")
	shared.Mu.Lock()
	defer shared.Mu.Unlock()

	var request struct {
		PlayerID string `json:"player_id"`
		Guess    string `json:"guess"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

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
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Player not found"})
		return
	}

	// **Always Use the Word from the Database**
	log.Printf("Player ID: %s - Retrieved Word from DB: %s", request.PlayerID, player.Word)

	// If the word in the database is empty or missing, assign a new one
	if player.Word == "" {
		player.Word = generateWord()
		updateResult, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"word": player.Word}})
		if err != nil {
			log.Println("Error assigning word to player:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign word"})
			return
		}
		log.Printf("New word assigned from DB: %s (Update result: %v)", player.Word, updateResult)
	}

	// **Check if the guess is correct**
	normalizedWord := strings.ToLower(player.Word)
	normalizedGuess := strings.ToLower(request.Guess)

	log.Printf("Normalized Word from DB: %s, Normalized Guess: %s", normalizedWord, normalizedGuess)

	if normalizedGuess == normalizedWord {
		player.Score++

		// Assign a new word
		newWord := generateWord()
		updateResult, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"word": newWord, "score": player.Score}})
		if err != nil {
			log.Println("Error updating word in DB:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update word"})
			return
		}
		log.Printf("New word updated in DB: %s (Update result: %v)", newWord, updateResult)

		// Update in-memory player list
		for conn, p := range shared.Players {
			if p.Name == player.Name {
				p.Score = player.Score
				p.Word = newWord
				shared.Players[conn] = p
				break
			}
		}

		go broadcastPlayerList()

		if player.Score == 3 {
			gameState.Winner = &player
			gameState.Started = false
			log.Println("Broadcasting game over for winner:", player.Name)

			select {
			case shared.Broadcast <- shared.Message{
				Type: "game_over",
				Payload: gin.H{
					"winner":  player.Name,
					"message": fmt.Sprintf("%s won the game!", player.Name),
				},
			}:
				log.Println("Game over broadcast sent.")
			default:
				log.Println("Broadcast channel is full, dropping message!")
			}

			_, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$inc": bson.M{"wins": 1}})
			if err != nil {
				log.Printf("Failed to update wins: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wins"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":  fmt.Sprintf("%s won the game!", player.Name),
				"correct":  true,
				"player":   player,
				"new_word": shuffleString(newWord),
				"scores":   getScores(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Correct! New word assigned.",
				"correct": true,
				"player": gin.H{
					"name":  player.Name,
					"score": player.Score,
				},
				"new_word": shuffleString(newWord),
				"scores":   getScores(),
			})
		}

	} else {
		log.Println("Incorrect guess. Try again.")
		c.JSON(http.StatusOK, gin.H{
			"message": "Incorrect, try again!",
			"correct": false,
			"scores":  getScores(),
		})
	}
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

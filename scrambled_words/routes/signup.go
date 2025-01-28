package routes

import (
	"scrambled_words/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)
	r.POST("/join", controllers.JoinGame)
	r.GET("/start", controllers.StartGame)
	r.POST("/submit", controllers.SubmitAnswer)
	r.GET("/leaderboard", controllers.GetLeaderboard)
}

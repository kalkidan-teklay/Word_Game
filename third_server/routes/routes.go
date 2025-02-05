package routes

import (
	"third_server/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {

	r.POST("/join", controllers.JoinGame)
	r.POST("/start", controllers.StartGame)
	r.POST("/menu", controllers.CheckMenu)
	r.POST("/submit", controllers.SubmitAnswer)
	// r.GET("/leaderboard", controllers.GetLeaderboard)
	r.GET("/ws", func(c *gin.Context) {
		controllers.HandleWebSocket(c.Writer, c.Request)
	})

}

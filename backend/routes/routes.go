package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shjung-dev/ChatApplication/backend/controllers"
	"github.com/shjung-dev/ChatApplication/backend/middleware"
)

func SetUpRoutes(r *gin.Engine) {
	r.POST("/login", controllers.Login())
	r.POST("/signup", controllers.Signup())
	r.POST("/refresh", controllers.RefreshTokenHandler())

	protected := r.Group("/")

	protected.Use(middleware.Authenticate())
	{
		protected.GET("/user/:receiver", controllers.SearchUser())
		protected.POST("/accept/:username", controllers.Accept())
		protected.POST("/reject/:receiver", controllers.Reject())
		protected.POST("/remove/:username", controllers.Remove())
	}
}

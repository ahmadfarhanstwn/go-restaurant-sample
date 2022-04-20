package routes

import (
	"github.com/ahmadfarhanstwn/go-restaurant-management/controllers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/users", controllers.GetUsers())
	incomingRoutes.GET("/users/:user_id", controllers.GetUser())
	incomingRoutes.POST("/users/sign-up", controllers.SignUp())
	incomingRoutes.POST("/users/sign-in", controllers.SignIn())
}
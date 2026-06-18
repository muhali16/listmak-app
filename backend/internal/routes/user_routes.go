package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
	"github.com/muhali16/listmak-service/internal/handlers/middlewares"
)

func UserRoutes(r *gin.RouterGroup, uc controllers.UserController) {
	userRoute := r.Group("/users")
	userRoute.Use(middlewares.AuthMiddleware())
	{
		userRoute.GET("", uc.GetUsers)
		userRoute.POST("", uc.CreateUser)
	}
}

package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
	"github.com/muhali16/listmak-service/internal/handlers/middlewares"
)

func AuthRoute(r *gin.RouterGroup, ac controllers.AuthController) {
	authRoute := r.Group("/auth")
	{
		authRoute.GET("/google/login", ac.GoogleLogin)
		authRoute.GET("/google/callback", ac.GoogleCallback)
		authRoute.Use(middlewares.AuthMiddleware()).GET("/logout", ac.Logout)
		authRoute.Use(middlewares.AuthMiddleware()).GET("/user", ac.GetUserAuthenticated)
	}
}

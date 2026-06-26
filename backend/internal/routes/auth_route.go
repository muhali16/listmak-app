package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
	"github.com/muhali16/listmak-service/internal/handlers/middlewares"
	"golang.org/x/time/rate"
)

func AuthRoute(r *gin.RouterGroup, ac controllers.AuthController) {
	authRoute := r.Group("/auth")
	authLoginLimiter := middlewares.RateLimiter(rate.Every(10*time.Second), 3) // 6 req/min, burst 3
	{
		authRoute.GET("/google/login", authLoginLimiter, ac.GoogleLogin)
		authRoute.GET("/google/callback", authLoginLimiter, ac.GoogleCallback)
		authRoute.Use(middlewares.AuthMiddleware()).GET("/logout", ac.Logout)
		authRoute.Use(middlewares.AuthMiddleware()).GET("/user", ac.GetUserAuthenticated)
	}
}

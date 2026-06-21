package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
	"github.com/muhali16/listmak-service/internal/handlers/middlewares"
)

func AIRoutes(r *gin.RouterGroup, ac controllers.AIController) {
	ai := r.Group("/ai")
	ai.Use(middlewares.AuthMiddleware())
	{
		ai.POST("/parse-orders", ac.ParseOrders)
	}
}

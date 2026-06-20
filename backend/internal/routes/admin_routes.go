package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
	"github.com/muhali16/listmak-service/internal/handlers/middlewares"
)

func AdminRoutes(r *gin.RouterGroup, ac controllers.AdminController) {
	admin := r.Group("/admin")
	admin.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
	{
		admin.GET("/ai-logs", ac.GetAILogs)
		admin.PATCH("/users/:id/role", ac.UpdateUserRole)
	}
}

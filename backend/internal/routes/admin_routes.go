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
		admin.GET("/system-logs", ac.GetSystemLogs)
		admin.PATCH("/users/:id/role", ac.UpdateUserRole)
		admin.GET("/listmaks", ac.GetAllListmaks)
		admin.GET("/price-catalog", ac.GetPriceCatalog)
		admin.POST("/price-catalog", ac.UpsertPriceCatalog)
		admin.DELETE("/price-catalog/:id", ac.DeletePriceCatalog)
		admin.DELETE("/view-shares/:id", ac.DeleteViewShare)
		admin.DELETE("/summaries/listmak/:listmakId", ac.DeleteSummary)
	}
}

package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
	"github.com/muhali16/listmak-service/internal/handlers/middlewares"
)

func SummaryRoutes(r *gin.RouterGroup, sc controllers.SummaryController) {
	// Nested under listmaks
	listmakGroup := r.Group("/listmaks")
	listmakGroup.Use(middlewares.AuthMiddleware())
	{
		listmakGroup.GET("/:id/summary", sc.GetSummary)
		listmakGroup.PUT("/:id/summary/confirm", sc.ConfirmPrices)
	}

	// Price catalog estimate (auth required)
	catalogGroup := r.Group("/price-catalog")
	catalogGroup.Use(middlewares.AuthMiddleware())
	{
		catalogGroup.GET("/estimate", sc.EstimatePrice)
	}
}

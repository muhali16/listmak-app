package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
	"github.com/muhali16/listmak-service/internal/handlers/middlewares"
)

func ShareRoutes(r *gin.RouterGroup, sc controllers.ShareController) {
	// Share Links (Private management)
	shareLinks := r.Group("/share-links")
	{
		// Management endpoints need auth
		private := shareLinks.Group("")
		private.Use(middlewares.AuthMiddleware())
		{
			private.POST("", sc.CreateShareLink)
			private.DELETE("/:id", sc.DeleteShareLink)
		}

		// Public endpoints (no auth middleware, but maybe valid shareId check inside controller)
		// Important: Conflict with auth routes? No, paths differ.
		// GET /share-links/:shareId -> Public
		shareLinks.GET("/:shareId", sc.GetShareLink)
		shareLinks.GET("/:shareId/orders", sc.GetOrdersViaShare)
		shareLinks.POST("/:shareId/orders", sc.SubmitOrderViaShare)
		shareLinks.GET("/:shareId/food-suggestions", sc.GetFoodSuggestions)
	}

	// View Shares
	viewShares := r.Group("/view-shares")
	{
		// Management
		private := viewShares.Group("")
		private.Use(middlewares.AuthMiddleware())
		{
			private.POST("", sc.CreateViewShare)
		}

		// Public
		viewShares.GET("/:viewId", sc.GetViewShare)
	}

	// Active shares per listmak (auth required)
	listmakShares := r.Group("/listmaks")
	listmakShares.Use(middlewares.AuthMiddleware())
	{
		listmakShares.GET("/:id/active-shares", sc.GetActiveSharesForListmak)
	}
}

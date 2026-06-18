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
		// GET /share-links/:shareId/orders -> Public
		shareLinks.GET("/:shareId/orders", sc.GetOrdersViaShare)
		// POST /share-links/:shareId/orders -> Public
		shareLinks.POST("/:shareId/orders", sc.SubmitOrderViaShare)
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
}

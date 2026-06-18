package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
	"github.com/muhali16/listmak-service/internal/handlers/middlewares"
)

func ListmakRoutes(r *gin.RouterGroup, lc controllers.ListmakController, oc controllers.OrderController) {
	// Group listmaks requiring auth
	listmakGroup := r.Group("/listmaks")
	listmakGroup.Use(middlewares.AuthMiddleware())
	{
		listmakGroup.GET("", lc.GetListmaks)
		listmakGroup.GET("/:id", lc.GetListmakById)
		listmakGroup.GET("/date/:date", lc.GetListmakByDate)
		listmakGroup.POST("", lc.CreateListmak)
		listmakGroup.PUT("/:id", lc.UpdateListmak)
		listmakGroup.DELETE("/:id", lc.DeleteListmak)

		// Nested orders in listmaks
		// Nested orders in listmaks
		// Use :id consistently to avoid wild card conflict with /:id endpoint
		listmakGroup.GET("/:id/orders", oc.GetOrders)
		listmakGroup.POST("/:id/orders", oc.CreateOrder)
		listmakGroup.POST("/:id/orders/bulk", oc.CreateOrdersBulk)
	}

	// Group orders requiring auth
	orderGroup := r.Group("/orders")
	orderGroup.Use(middlewares.AuthMiddleware())
	{
		orderGroup.PUT("/:id", oc.UpdateOrder)
		orderGroup.PATCH("/:id/paid", oc.UpdateOrderPaid)
		orderGroup.DELETE("/:id", oc.DeleteOrder)
	}
}

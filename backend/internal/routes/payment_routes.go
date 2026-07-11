package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
)

// PaymentRoutes registers the guest payment endpoints. All are public (no auth):
// guests are unauthenticated and identify themselves via name + WhatsApp at
// checkout. Security lives in the service layer (share validation, server-side
// amount, gateway status verification, idempotency).
func PaymentRoutes(r *gin.RouterGroup, pc controllers.PaymentController) {
	payments := r.Group("/payments")
	{
		payments.POST("/checkout", pc.Checkout)
		payments.POST("/webhook", pc.Webhook)
		payments.GET("/:orderId/status", pc.GetStatus)
		payments.POST("/:orderId/cancel", pc.Cancel)
	}
}

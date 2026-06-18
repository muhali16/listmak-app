package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
)

func LogRoutes(r *gin.RouterGroup) {
	r.GET("/logs", controllers.GetAllLogs)
	r.GET("/logs/:request_id", controllers.GetLogByRequestID)
}

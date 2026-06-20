package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/pkg/utils"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			utils.SendResponse(c, http.StatusForbidden, false, "Admin access required", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

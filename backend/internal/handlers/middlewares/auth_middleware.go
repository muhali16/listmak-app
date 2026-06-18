package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/pkg/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var authToken string

		// Pertama coba ambil dari cookie
		authToken, err := c.Cookie("X-User-Authentication-Token")

		// Jika tidak ada di cookie, coba ambil dari Authorization header
		if err != nil || authToken == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				authToken = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Jika masih tidak ada token, return unauthorized
		if authToken == "" {
			utils.SendResponse(c, http.StatusUnauthorized, false, "Missing user authentication token", nil)
			c.Abort()
			return
		}

		claims, err := utils.ValidateJWT(authToken)
		if err != nil {
			utils.SendResponse(c, http.StatusUnauthorized, false, "Invalid user authentication token", nil)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

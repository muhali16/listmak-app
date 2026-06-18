package middlewares

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	origins := []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		origins = append(origins, frontendURL)
	}
	return cors.New(cors.Config{
		AllowOrigins:  origins,
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		// SANGAT PENTING: Izinkan pengiriman cookie
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

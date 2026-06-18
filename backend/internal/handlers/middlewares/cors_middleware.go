package middlewares

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		// Ganti dengan URL frontend kamu (React/Vue/Next.js)
		AllowOrigins:  []string{"http://localhost:5173", "http://127.0.0.1:5173", os.Getenv("FRONTEND_URL")},
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		// SANGAT PENTING: Izinkan pengiriman cookie
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/muhali16/listmak-service/internal/configs"
	"github.com/muhali16/listmak-service/internal/handlers/middlewares"
	"github.com/muhali16/listmak-service/internal/repository"
	"github.com/muhali16/listmak-service/internal/routes"
)

// @title           Listmak Service API
// @version         1.0
// @description     API Server untuk manajemen Listmak
// @host            localhost:9001
// @BasePath        /api/v1
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configs.InitDB()
	db := configs.GetDB()
	configs.AutoMigrate(db)

	systemLogRepo := repository.NewSystemLogRepository(db)

	// background cleanup: delete AI logs older than 3 months, runs daily
	go func() {
		aiLogRepo := repository.NewAILogRepository(db)
		for {
			cutoff := time.Now().AddDate(0, -3, 0)
			n, err := aiLogRepo.DeleteOlderThan(cutoff)
			if err != nil {
				log.Printf("AI log cleanup error: %v", err)
			} else if n > 0 {
				log.Printf("AI log cleanup: deleted %d records older than %s", n, cutoff.Format("2006-01-02"))
			}
			time.Sleep(24 * time.Hour)
		}
	}()

	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1", "192.168.1.8.sslip.io", "192.168.1.8"})

	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.LoggerWithID(systemLogRepo))

	routes.Routes(r, systemLogRepo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on localhost:%s", port)
	r.Run(":" + port)
}

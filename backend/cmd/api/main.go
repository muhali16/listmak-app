package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/muhali16/listmak-service/internal/configs"
	"github.com/muhali16/listmak-service/internal/handlers/middlewares"
	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/routes"
)

// @title           Listmak Service API
// @version         1.0
// @description     API Server untuk manajemen Listmak
// @host            localhost:9001
// @BasePath        /api/v1
func main() {
	// 1. Load ENV & Database pertama kali
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// inisialisasi database
	models.InitLogDB() // log DB
	configs.InitDB()   // primary DB
	db := configs.GetDB()
	configs.AutoMigrate(db)

	// inisialisasi gin
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 2. Inisialisasi Gin
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1", "192.168.1.8.sslip.io", "192.168.1.8"})

	// 3. Pasang Middleware (WAJIB sebelum routes)
	// CORS harus dipasang pertama kali untuk menangani preflight request
	r.Use(middlewares.CORSMiddleware())
	// Semua request yang masuk setelah baris ini akan memiliki RequestID
	r.Use(middlewares.LoggerWithID())

	// 4. Setup HTML Template removed (using JSON API now)
	// r.SetHTMLTemplate(t)

	// 5. Routes are handled in routes/routes.go

	routes.Routes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on localhost:%s", port)
	r.Run(":" + port)
}

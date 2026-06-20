package controllers

import (
	"os"

	"github.com/muhali16/listmak-service/internal/repository"
	"github.com/muhali16/listmak-service/internal/services"
	"gorm.io/gorm"
)

type Container struct {
	UserController    UserController
	AuthController    AuthController
	ListmakController ListmakController
	OrderController   OrderController
	ShareController   ShareController
	AdminController   AdminController
}

func InitContainer(db *gorm.DB, systemLogRepo repository.SystemLogRepository) *Container {
	// init repositories
	userRepo := repository.NewUserRepository(db)
	listmakRepo := repository.NewListmakRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	shareRepo := repository.NewShareLinkRepository(db)
	viewShareRepo := repository.NewViewShareRepository(db)
	aiLogRepo := repository.NewAILogRepository(db)

	// init AI service
	var aiService services.AIService
	apiKey := os.Getenv("FIREWORKS_API_KEY")
	model := os.Getenv("FIREWORKS_MODEL")
	if apiKey != "" && model != "" {
		aiService = services.NewFireworksAIService(apiKey, model, aiLogRepo)
	} else {
		aiService = services.NewNoopAIService()
	}

	// init services
	userService := services.NewUserService(userRepo)
	listmakService := services.NewListmakService(listmakRepo)
	orderService := services.NewOrderService(orderRepo, listmakRepo, aiService)
	shareService := services.NewShareService(shareRepo, viewShareRepo, listmakRepo)

	// init controllers
	userController := NewUserController(userService)
	authController := NewAuthController(userService)
	listmakController := NewListmakController(listmakService)
	orderController := NewOrderController(orderService)
	shareController := NewShareController(shareService, orderService)
	adminController := NewAdminController(aiLogRepo, systemLogRepo, userRepo)

	return &Container{
		UserController:    userController,
		AuthController:    authController,
		ListmakController: listmakController,
		OrderController:   orderController,
		ShareController:   shareController,
		AdminController:   adminController,
	}
}

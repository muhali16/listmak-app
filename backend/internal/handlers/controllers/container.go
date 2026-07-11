package controllers

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/muhali16/listmak-service/internal/models"
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
	SummaryController SummaryController
	AIController      AIController
	PaymentController PaymentController
	ConfigController  ConfigController
}

func InitContainer(db *gorm.DB, systemLogRepo repository.SystemLogRepository) *Container {
	// init repositories
	userRepo := repository.NewUserRepository(db)
	listmakRepo := repository.NewListmakRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	shareRepo := repository.NewShareLinkRepository(db)
	viewShareRepo := repository.NewViewShareRepository(db)
	aiLogRepo := repository.NewAILogRepository(db)
	summaryRepo := repository.NewSummaryRepository(db)
	catalogRepo := repository.NewPriceCatalogRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	paymentLogRepo := repository.NewPaymentLogRepository(db)
	appSettingRepo := repository.NewAppSettingRepository(db)

	// Runtime config (admin-togglable testing mode). Seed default from env.
	appConfig := services.NewAppConfig(appSettingRepo, os.Getenv("TESTING_MODE") == "true")

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
	listmakService := services.NewListmakService(listmakRepo, appConfig)
	orderService := services.NewOrderService(orderRepo, listmakRepo, aiService)
	shareService := services.NewShareService(shareRepo, viewShareRepo, listmakRepo)
	summaryService := services.NewSummaryService(summaryRepo, catalogRepo, orderRepo, aiService)

	// Gateway audit log: append one row per Pakasir interaction (async, best-effort).
	paymentLogFn := func(orderID, action string, statusCode int, response []byte, errMsg string) {
		resp := response
		if len(resp) == 0 {
			resp = nil
		} else if !json.Valid(resp) {
			// Non-JSON body (rare error case) — wrap so the jsonb column stays valid.
			resp, _ = json.Marshal(map[string]string{"raw": string(response)})
		}
		go paymentLogRepo.Create(&models.PaymentLog{
			OrderID:    orderID,
			Action:     action,
			StatusCode: statusCode,
			Success:    errMsg == "",
			Response:   resp,
			Error:      errMsg,
		})
	}

	// init payment gateway (Pakasir). Disabled gracefully if env not set —
	// Checkout returns ErrPaymentNotConfigured -> HTTP 503.
	pakasirClient := services.NewPakasirClient(
		os.Getenv("PAKASIR_BASE_URL"),
		os.Getenv("PAKASIR_PROJECT"),
		os.Getenv("PAKASIR_API_KEY"),
		paymentLogFn,
	)
	paymentService := services.NewPaymentService(paymentRepo, orderService, shareService, pakasirClient, paymentLogFn, appConfig)

	// Background reconciler: settle payments still pending after 10 min (complete
	// if actually paid, otherwise cancel so a saved/screenshot QR can't be paid
	// late). Only runs when the gateway is configured.
	if pakasirClient.Configured() {
		go func() {
			for {
				if n, err := paymentService.ReconcileStalePending(10 * time.Minute); err != nil {
					log.Printf("payment reconcile error: %v", err)
				} else if n > 0 {
					log.Printf("payment reconcile: handled %d stale pending", n)
				}
				time.Sleep(2 * time.Minute)
			}
		}()
	}

	// init controllers
	userController := NewUserController(userService)
	authController := NewAuthController(userService)
	listmakController := NewListmakController(listmakService)
	orderController := NewOrderController(orderService)
	shareController := NewShareController(shareService, orderService, aiService)
	adminController := NewAdminController(aiLogRepo, systemLogRepo, userRepo, listmakRepo, catalogRepo, viewShareRepo, summaryRepo, paymentRepo, paymentLogRepo, orderRepo, paymentService)
	summaryController := NewSummaryController(summaryService)
	aiController := NewAIController(aiService)
	paymentController := NewPaymentController(paymentService)
	configController := NewConfigController(appConfig)

	return &Container{
		UserController:    userController,
		AuthController:    authController,
		ListmakController: listmakController,
		OrderController:   orderController,
		ShareController:   shareController,
		AdminController:   adminController,
		SummaryController: summaryController,
		AIController:      aiController,
		PaymentController: paymentController,
		ConfigController:  configController,
	}
}

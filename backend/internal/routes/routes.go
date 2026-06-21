package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/muhali16/listmak-service/docs"
	"github.com/muhali16/listmak-service/internal/configs"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
	"github.com/muhali16/listmak-service/internal/repository"
	"github.com/muhali16/listmak-service/pkg/utils"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Routes(r *gin.Engine, systemLogRepo repository.SystemLogRepository) {
	container := controllers.InitContainer(configs.GetDB(), systemLogRepo)

	r.NoRoute(func(c *gin.Context) {
		utils.SendResponse(c, http.StatusNotFound, false, "What you looking for?", nil)
	})
	r.HandleMethodNotAllowed = true
	r.NoMethod(func(c *gin.Context) {
		utils.SendResponse(c, http.StatusMethodNotAllowed, false, "Illegal method", nil)
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	UserRoutes(v1, container.UserController)
	AuthRoute(v1, container.AuthController)
	ListmakRoutes(v1, container.ListmakController, container.OrderController)
	ShareRoutes(v1, container.ShareController)
	AdminRoutes(v1, container.AdminController)
	SummaryRoutes(v1, container.SummaryController)
	AIRoutes(v1, container.AIController)
}

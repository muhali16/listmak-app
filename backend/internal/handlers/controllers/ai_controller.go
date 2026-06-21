package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/services"
	"github.com/muhali16/listmak-service/pkg/utils"
)

type AIController interface {
	ParseOrders(c *gin.Context)
}

type aiController struct {
	aiService services.AIService
}

func NewAIController(aiService services.AIService) AIController {
	return &aiController{aiService: aiService}
}

func (ac *aiController) ParseOrders(c *gin.Context) {
	var payload struct {
		Orders   []services.ParseOrderInput `json:"orders"`
		Location string                     `json:"location"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil || len(payload.Orders) == 0 {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	items, err := ac.aiService.ParseOrders(c.GetString("RequestID"), payload.Orders, payload.Location)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal parsing pesanan: "+err.Error(), nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "OK", items)
}

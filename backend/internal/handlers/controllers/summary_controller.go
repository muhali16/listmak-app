package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/services"
	"github.com/muhali16/listmak-service/pkg/utils"
)

type SummaryController interface {
	GetSummary(c *gin.Context)
	ConfirmPrices(c *gin.Context)
	EstimatePrice(c *gin.Context)
}

type summaryController struct {
	svc services.SummaryService
}

func NewSummaryController(svc services.SummaryService) SummaryController {
	return &summaryController{svc: svc}
}

func (h *summaryController) GetSummary(c *gin.Context) {
	listmakID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid listmak ID", nil)
		return
	}

	requestID := c.GetString("RequestID")
	location := c.Query("location")
	if location == "" {
		if lat, lng := c.Query("lat"), c.Query("lng"); lat != "" && lng != "" {
			location = fmt.Sprintf("koordinat lat %s, lng %s", lat, lng)
		}
	}

	summary, err := h.svc.GetOrGenerateSummary(requestID, uint(listmakID), location)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, err.Error(), nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "Ringkasan berhasil diambil", summary)
}

func (h *summaryController) ConfirmPrices(c *gin.Context) {
	listmakID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid listmak ID", nil)
		return
	}

	var items []services.ConfirmItem
	if err := c.ShouldBindJSON(&items); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid request body", nil)
		return
	}

	summary, err := h.svc.ConfirmPrices(uint(listmakID), items)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, err.Error(), nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "Harga berhasil dikonfirmasi", summary)
}

func (h *summaryController) EstimatePrice(c *gin.Context) {
	itemDetail := c.Query("item")
	if itemDetail == "" {
		utils.SendResponse(c, http.StatusBadRequest, false, "Query parameter 'item' wajib diisi", nil)
		return
	}

	requestID := c.GetString("RequestID")
	location := c.Query("location")
	if location == "" {
		if lat, lng := c.Query("lat"), c.Query("lng"); lat != "" && lng != "" {
			location = fmt.Sprintf("koordinat lat %s, lng %s", lat, lng)
		}
	}

	price, isEstimated, err := h.svc.EstimatePrice(requestID, itemDetail, location)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, err.Error(), nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "Estimasi harga berhasil diambil", gin.H{
		"price":        price,
		"is_estimated": isEstimated,
	})
}

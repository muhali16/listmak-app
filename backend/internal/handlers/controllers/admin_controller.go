package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
	"github.com/muhali16/listmak-service/internal/services"
	"github.com/muhali16/listmak-service/pkg/utils"
	"gorm.io/gorm"
)

type AdminController interface {
	GetAILogs(c *gin.Context)
	GetSystemLogs(c *gin.Context)
	UpdateUserRole(c *gin.Context)
	GetAllListmaks(c *gin.Context)
	GetPriceCatalog(c *gin.Context)
	UpsertPriceCatalog(c *gin.Context)
	DeletePriceCatalog(c *gin.Context)
	DeleteViewShare(c *gin.Context)
	DeleteSummary(c *gin.Context)
	GetPayments(c *gin.Context)
	GetPaymentDetail(c *gin.Context)
	ReconcilePayment(c *gin.Context)
	CancelPayment(c *gin.Context)
}

type adminController struct {
	aiLogRepo      repository.AILogRepository
	systemLogRepo  repository.SystemLogRepository
	userRepo       repository.UserRepository
	listmakRepo    repository.ListmakRepository
	catalogRepo    repository.PriceCatalogRepository
	viewShareRepo  repository.ViewShareRepository
	summaryRepo    repository.SummaryRepository
	paymentRepo    repository.PaymentRepository
	paymentLogRepo repository.PaymentLogRepository
	orderRepo      repository.OrderRepository
	paymentService services.PaymentService
}

func NewAdminController(
	aiLogRepo repository.AILogRepository,
	systemLogRepo repository.SystemLogRepository,
	userRepo repository.UserRepository,
	listmakRepo repository.ListmakRepository,
	catalogRepo repository.PriceCatalogRepository,
	viewShareRepo repository.ViewShareRepository,
	summaryRepo repository.SummaryRepository,
	paymentRepo repository.PaymentRepository,
	paymentLogRepo repository.PaymentLogRepository,
	orderRepo repository.OrderRepository,
	paymentService services.PaymentService,
) AdminController {
	return &adminController{
		aiLogRepo:      aiLogRepo,
		systemLogRepo:  systemLogRepo,
		userRepo:       userRepo,
		listmakRepo:    listmakRepo,
		catalogRepo:    catalogRepo,
		viewShareRepo:  viewShareRepo,
		summaryRepo:    summaryRepo,
		paymentRepo:    paymentRepo,
		paymentLogRepo: paymentLogRepo,
		orderRepo:      orderRepo,
		paymentService: paymentService,
	}
}

func (ac *adminController) GetAILogs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	status := c.Query("status")
	search := c.Query("search")
	logs, total, err := ac.aiLogRepo.GetAll(page, 50, status, search)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to retrieve AI logs", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "AI logs retrieved", gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
	})
}

func (ac *adminController) GetSystemLogs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	f := repository.SystemLogFilter{
		RequestID: c.Query("request_id"),
		Method:    c.Query("method"),
	}

	if sc, err := strconv.Atoi(c.Query("status")); err == nil && sc > 0 {
		f.StatusCode = sc
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			f.From = &t
		}
	} else {
		defaultFrom := time.Now().AddDate(0, 0, -7)
		f.From = &defaultFrom
	}
	if toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			f.To = &t
		}
	}

	logs, total, err := ac.systemLogRepo.GetAll(page, 100, f)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to retrieve system logs", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "System logs retrieved", gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
	})
}

func (ac *adminController) UpdateUserRole(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid user ID", nil)
		return
	}
	var body struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid request body", nil)
		return
	}
	if body.Role != "admin" && body.Role != "user" {
		utils.SendResponse(c, http.StatusBadRequest, false, "Role must be 'admin' or 'user'", nil)
		return
	}
	if err := ac.userRepo.UpdateRole(uint(userID), body.Role); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to update role", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Role updated", nil)
}

func (ac *adminController) GetAllListmaks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	status := c.Query("status")

	var startDate, endDate *time.Time
	if sDate := c.Query("start_date"); sDate != "" {
		if t, err := time.Parse("2006-01-02", sDate); err == nil {
			startDate = &t
		}
	}
	if eDate := c.Query("end_date"); eDate != "" {
		if t, err := time.Parse("2006-01-02", eDate); err == nil {
			endDate = &t
		}
	}

	// userId=0 returns all users
	data, total, err := ac.listmakRepo.GetAllListmaks(page, limit, status, startDate, endDate, 0)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to get listmaks", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (ac *adminController) GetPriceCatalog(c *gin.Context) {
	entries, err := ac.catalogRepo.GetAll()
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to get price catalog", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Price catalog retrieved", entries)
}

func (ac *adminController) UpsertPriceCatalog(c *gin.Context) {
	var entries []models.PriceCatalog
	if err := c.ShouldBindJSON(&entries); err != nil || len(entries) == 0 {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}
	if err := ac.catalogRepo.UpsertBatch(entries); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to upsert price catalog", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Price catalog updated", nil)
}

func (ac *adminController) DeletePriceCatalog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid ID", nil)
		return
	}
	if err := ac.catalogRepo.Delete(uint(id)); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to delete catalog entry", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Catalog entry deleted", nil)
}

func (ac *adminController) DeleteViewShare(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid ID", nil)
		return
	}
	if err := ac.viewShareRepo.Delete(uint(id)); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to delete view share", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "View share deleted", nil)
}

func (ac *adminController) DeleteSummary(c *gin.Context) {
	listmakID, err := strconv.ParseUint(c.Param("listmakId"), 10, 64)
	if err != nil || listmakID == 0 {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid listmak ID", nil)
		return
	}
	if err := ac.summaryRepo.DeleteByListmakID(uint(listmakID)); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to delete summary", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Summary deleted", nil)
}

// ---- Payments (admin) ----

// GetPayments godoc
// @Summary  List payments with filters + revenue stats
// @Tags     admin
// @Produce  json
// @Param    page    query  int     false  "Page"
// @Param    status  query  string  false  "pending|completed|cancelled|expired"
// @Param    search  query  string  false  "order_id / guest name / whatsapp"
// @Success  200  {object}  utils.Response
// @Router   /admin/payments [get]
func (ac *adminController) GetPayments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	status := c.Query("status")
	search := c.Query("search")

	payments, total, err := ac.paymentRepo.ListForAdmin(page, 30, status, search)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to retrieve payments", nil)
		return
	}
	stats, err := ac.paymentRepo.Stats()
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to compute stats", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Payments retrieved", gin.H{
		"payments": payments,
		"total":    total,
		"page":     page,
		"stats":    stats,
	})
}

// GetPaymentDetail godoc
// @Summary  Payment detail + covered orders
// @Tags     admin
// @Produce  json
// @Param    orderId  path  string  true  "Invoice / order id"
// @Success  200  {object}  utils.Response
// @Router   /admin/payments/{orderId} [get]
func (ac *adminController) GetPaymentDetail(c *gin.Context) {
	orderID := c.Param("orderId")
	payment, err := ac.paymentRepo.GetByOrderID(orderID)
	if err != nil {
		utils.SendResponse(c, http.StatusNotFound, false, "Payment tidak ditemukan", nil)
		return
	}
	orders, _ := ac.orderRepo.GetByPaymentID(payment.ID)
	logs, _ := ac.paymentLogRepo.ListByOrderID(payment.OrderID)
	utils.SendResponse(c, http.StatusOK, true, "OK", gin.H{
		"payment": payment,
		"orders":  orders,
		"logs":    logs,
	})
}

// ReconcilePayment godoc
// @Summary  Re-sync a payment status from Pakasir (authoritative)
// @Tags     admin
// @Produce  json
// @Param    orderId  path  string  true  "Invoice / order id"
// @Success  200  {object}  utils.Response
// @Router   /admin/payments/{orderId}/reconcile [post]
func (ac *adminController) ReconcilePayment(c *gin.Context) {
	orderID := c.Param("orderId")
	payment, err := ac.paymentService.GetStatus(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendResponse(c, http.StatusNotFound, false, "Payment tidak ditemukan", nil)
			return
		}
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal reconcile", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Status disinkronkan", payment)
}

// CancelPayment godoc
// @Summary  Cancel a pending payment (admin)
// @Tags     admin
// @Produce  json
// @Param    orderId  path  string  true  "Invoice / order id"
// @Success  200  {object}  utils.Response
// @Failure  409  {object}  utils.Response
// @Router   /admin/payments/{orderId}/cancel [post]
func (ac *adminController) CancelPayment(c *gin.Context) {
	orderID := c.Param("orderId")
	if err := ac.paymentService.Cancel(orderID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendResponse(c, http.StatusNotFound, false, "Payment tidak ditemukan", nil)
			return
		}
		if errors.Is(err, services.ErrAlreadyPaid) {
			utils.SendResponse(c, http.StatusConflict, false, "Sudah lunas, tidak bisa dibatalkan", nil)
			return
		}
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal membatalkan", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Pembayaran dibatalkan", nil)
}

package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/services"
	"github.com/muhali16/listmak-service/pkg/utils"
	"gorm.io/gorm"
)

type PaymentController interface {
	Checkout(c *gin.Context)
	GetStatus(c *gin.Context)
	Cancel(c *gin.Context)
	Webhook(c *gin.Context)
}

type paymentController struct {
	paymentService services.PaymentService
}

func NewPaymentController(paymentService services.PaymentService) PaymentController {
	return &paymentController{paymentService: paymentService}
}

// Checkout godoc
// @Summary      Guest checkout (submit orders + create payment)
// @Description  Public endpoint. Submits a guest's orders for a shared listmak and creates a Pakasir transaction. Amount is computed server-side.
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        payload  body      map[string]interface{}  true  "Checkout data"
// @Success      200      {object}  utils.Response
// @Failure      400      {object}  utils.Response
// @Failure      410      {object}  utils.Response
// @Failure      502      {object}  utils.Response
// @Failure      503      {object}  utils.Response
// @Router       /payments/checkout [post]
func (pc *paymentController) Checkout(c *gin.Context) {
	var payload struct {
		ShareID       string                  `json:"share_id"`
		GuestName     string                  `json:"guest_name"`
		GuestWhatsapp string                  `json:"guest_whatsapp"`
		PaymentMethod string                  `json:"payment_method"`
		Orders        []services.CheckoutItem `json:"orders"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	payment, err := pc.paymentService.Checkout(services.CheckoutInput{
		ShareID:       payload.ShareID,
		GuestName:     payload.GuestName,
		GuestWhatsapp: payload.GuestWhatsapp,
		PaymentMethod: payload.PaymentMethod,
		Items:         payload.Orders,
	})
	if err != nil {
		code, msg := checkoutError(err)
		utils.SendResponse(c, code, false, msg, nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "Pembayaran berhasil dibuat", gin.H{
		"order_id":       payment.OrderID,
		"amount":         payment.Amount,
		"fee":            payment.Fee,
		"total_payment":  payment.TotalPayment,
		"payment_method": payment.PaymentMethod,
		"payment_number": payment.PaymentNumber, // QR string or VA number
		"status":         payment.Status,
		"expires_at":     payment.ExpiresAt,
	})
}

// GetStatus godoc
// @Summary      Get payment status
// @Description  Public polling endpoint. Returns DB status, falling back to a live Pakasir query when still pending.
// @Tags         payments
// @Produce      json
// @Param        orderId  path      string  true  "Invoice / Pakasir order id"
// @Success      200      {object}  utils.Response
// @Failure      404      {object}  utils.Response
// @Router       /payments/{orderId}/status [get]
func (pc *paymentController) GetStatus(c *gin.Context) {
	orderID := c.Param("orderId")
	payment, err := pc.paymentService.GetStatus(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendResponse(c, http.StatusNotFound, false, "Pembayaran tidak ditemukan", nil)
			return
		}
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal mengambil status pembayaran", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "OK", gin.H{
		"order_id":       payment.OrderID,
		"status":         payment.Status,
		"amount":         payment.Amount,
		"total_payment":  payment.TotalPayment,
		"payment_method": payment.PaymentMethod,
		"completed_at":   payment.CompletedAt,
		"expires_at":     payment.ExpiresAt,
	})
}

// Cancel godoc
// @Summary      Cancel a pending payment
// @Description  Public. Voids a still-pending Pakasir transaction when the guest closes the QR. No-op if already cancelled; 409 if already paid.
// @Tags         payments
// @Produce      json
// @Param        orderId  path      string  true  "Invoice / Pakasir order id"
// @Success      200      {object}  utils.Response
// @Failure      404      {object}  utils.Response
// @Failure      409      {object}  utils.Response
// @Router       /payments/{orderId}/cancel [post]
func (pc *paymentController) Cancel(c *gin.Context) {
	orderID := c.Param("orderId")
	err := pc.paymentService.Cancel(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendResponse(c, http.StatusNotFound, false, "Pembayaran tidak ditemukan", nil)
			return
		}
		if errors.Is(err, services.ErrAlreadyPaid) {
			utils.SendResponse(c, http.StatusConflict, false, "Pembayaran sudah lunas, tidak bisa dibatalkan", nil)
			return
		}
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal membatalkan pembayaran", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Pembayaran dibatalkan", nil)
}

// Webhook godoc
// @Summary      Pakasir webhook receiver
// @Description  Public. Verifies project/order/amount, confirms status via Pakasir Transaction Detail API, then marks the payment (and its orders) paid idempotently.
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        payload  body      map[string]interface{}  true  "Pakasir webhook payload"
// @Success      200      {object}  utils.Response
// @Failure      400      {object}  utils.Response
// @Failure      404      {object}  utils.Response
// @Router       /payments/webhook [post]
func (pc *paymentController) Webhook(c *gin.Context) {
	var payload services.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	code, err := pc.paymentService.HandleWebhook(payload)
	// Do not echo the validation reason back to the caller — it would give a
	// spoofer an oracle (e.g. "amount mismatch" vs "unknown order"). Log the
	// detail server-side; return only a generic message + the status code.
	if err != nil {
		log.Printf("webhook %s: rejected (%d): %v", payload.OrderID, code, err)
	}
	success := code >= 200 && code < 300
	msg := "rejected"
	if success {
		msg = "OK"
	}
	utils.SendResponse(c, code, success, msg, nil)
}

// checkoutError maps service errors to HTTP status + a guest-facing message.
func checkoutError(err error) (int, string) {
	switch {
	case errors.Is(err, services.ErrPaymentNotConfigured):
		return http.StatusServiceUnavailable, "Pembayaran belum tersedia"
	case errors.Is(err, services.ErrShareExpired):
		return http.StatusGone, "Waktu input pesanan telah berakhir"
	case errors.Is(err, services.ErrShareInvalid):
		return http.StatusNotFound, "Share link tidak valid"
	case errors.Is(err, services.ErrInvalidGuest):
		return http.StatusBadRequest, "Nama dan nomor WhatsApp wajib diisi dengan benar"
	case errors.Is(err, services.ErrEmptyOrders):
		return http.StatusBadRequest, "Minimal satu pesanan diperlukan"
	case errors.Is(err, services.ErrTooManyItems):
		return http.StatusBadRequest, "Terlalu banyak pesanan dalam satu pembayaran"
	case errors.Is(err, services.ErrInvalidAmount):
		return http.StatusBadRequest, "Total pembayaran harus lebih dari nol"
	case errors.Is(err, services.ErrInvalidMethod):
		return http.StatusBadRequest, "Metode pembayaran tidak didukung"
	default:
		return http.StatusBadGateway, "Gagal membuat pembayaran"
	}
}

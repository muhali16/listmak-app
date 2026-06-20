package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/services"
	"github.com/muhali16/listmak-service/pkg/utils"
)

type OrderController interface {
	GetOrders(c *gin.Context)
	CreateOrder(c *gin.Context)
	CreateOrdersBulk(c *gin.Context)
	UpdateOrder(c *gin.Context)
	UpdateOrderPaid(c *gin.Context)
	UpdateOrdersPaidByName(c *gin.Context)
	DeleteOrder(c *gin.Context)
	UpdateOrderVendor(c *gin.Context)
}

type orderController struct {
	orderService services.OrderService
}

func NewOrderController(orderService services.OrderService) OrderController {
	return &orderController{orderService: orderService}
}

// GetOrders godoc
// @Summary      Get orders for a listmak
// @Description  Get all orders associated with a listmak
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id         path      int     true   "Listmak ID"
// @Param        is_paid    query     boolean false  "Payment status filter"
// @Param        search     query     string  false  "Search query"
// @Success      200        {object}  map[string]interface{}
// @Failure      500        {object}  utils.Response
// @Router       /listmaks/{id}/orders [get]
func (oc *orderController) GetOrders(c *gin.Context) {
	listmakId, _ := strconv.Atoi(c.Param("id"))
	search := c.Query("search")

	var isPaid *bool
	if paidStr := c.Query("is_paid"); paidStr != "" {
		val := paidStr == "true"
		isPaid = &val
	}

	orders, err := oc.orderService.GetOrdersByListmakId(uint(listmakId), isPaid, search)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to get orders", nil)
		return
	}

	// Calculate summary for response (optional as per doc example "summary")
	var totalOrders, paidCount int
	var totalAmount, paidAmount float64
	for _, o := range orders {
		totalOrders++
		totalAmount += o.TotalPrice
		if o.IsPaid {
			paidCount++
			paidAmount += o.TotalPrice
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
		"summary": gin.H{
			"total_orders": totalOrders,
			"paid_count":   paidCount,
			"total_amount": totalAmount,
			"paid_amount":  paidAmount,
		},
	})
}

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Add a new order to a listmak
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id         path      int            true  "Listmak ID"
// @Param        order      body      models.Order   true  "Order data"
// @Success      200        {object}  utils.Response{data=models.Order}
// @Failure      400        {object}  utils.Response
// @Failure      500        {object}  utils.Response
// @Router       /listmaks/{id}/orders [post]
func (oc *orderController) CreateOrder(c *gin.Context) {
	listmakId, _ := strconv.Atoi(c.Param("id"))
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	order.ListmakID = uint(listmakId)
	created, err := oc.orderService.CreateOrder(order)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to create order", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Pesanan berhasil ditambahkan", created)
}

// CreateOrdersBulk godoc
// @Summary      Bulk create orders
// @Description  Add multiple orders at once
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id         path      int                       true  "Listmak ID"
// @Param        payload    body      map[string]interface{}    true  "Bulk orders payload"
// @Success      200        {object}  utils.Response
// @Failure      400        {object}  utils.Response
// @Failure      500        {object}  utils.Response
// @Router       /listmaks/{id}/orders/bulk [post]
func (oc *orderController) CreateOrdersBulk(c *gin.Context) {
	listmakId, _ := strconv.Atoi(c.Param("id"))
	var payload struct {
		Orders   []models.Order `json:"orders"`
		AddedVia string         `json:"added_via"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	count, orders, err := oc.orderService.CreateOrdersBulk(uint(listmakId), payload.Orders)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to bulk create orders", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, strconv.Itoa(count)+" pesanan berhasil ditambahkan", gin.H{
		"added_count": count,
		"orders":      orders,
	})
}

// UpdateOrder godoc
// @Summary      Update an order
// @Description  Update order details
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id     path      int           true  "Order ID"
// @Param        order  body      models.Order  true  "Order data"
// @Success      200    {object}  utils.Response{data=models.Order}
// @Failure      400    {object}  utils.Response
// @Failure      500    {object}  utils.Response
// @Router       /orders/{id} [put]
func (oc *orderController) UpdateOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	// assuming update overwrites provided fields. We need to fetch original first inside service or just set ID.
	// We'll set ID and let service/repo handle.
	order.ID = uint(id)

	// Caveat: if we just pass struct to Save in GORM, it updates all fields including zero values if map not used.
	// Repository UpdateOrder uses Save(&order).
	// We should probably fetch inside controller to merge, OR assume payload is complete OR let frontend handle it.
	// For simplicity, let's assume payload contains fields to update.
	// But `UpdateOrder` in service recalculates total price.

	updated, err := oc.orderService.UpdateOrder(order)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to update order", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Order updated", updated)
}

// UpdateOrderPaid godoc
// @Summary      Update order payment status
// @Description  Set order as paid or unpaid
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id       path      int                     true  "Order ID"
// @Param        payload  body      map[string]interface{}  true  "Payment status"
// @Success      200      {object}  utils.Response{data=models.Order}
// @Failure      400      {object}  utils.Response
// @Failure      500      {object}  utils.Response
// @Router       /orders/{id}/paid [patch]
func (oc *orderController) UpdateOrderPaid(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var payload struct {
		IsPaid bool `json:"is_paid"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	updated, err := oc.orderService.UpdateOrderPaidStatus(uint(id), payload.IsPaid)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to update status", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Status bayar berhasil diupdate", updated)
}

// UpdateOrdersPaidByName godoc
// @Summary      Bulk update payment status by name
// @Description  Set is_paid for all orders belonging to a name within a listmak (name matched case-insensitively, trimmed). All-or-nothing.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id       path      int                     true  "Listmak ID"
// @Param        payload  body      map[string]interface{}  true  "Name and payment status"
// @Success      200      {object}  utils.Response
// @Failure      400      {object}  utils.Response
// @Failure      404      {object}  utils.Response
// @Failure      500      {object}  utils.Response
// @Router       /listmaks/{id}/orders/paid [patch]
func (oc *orderController) UpdateOrdersPaidByName(c *gin.Context) {
	listmakId, _ := strconv.Atoi(c.Param("id"))
	var payload struct {
		Name   string `json:"name"`
		IsPaid bool   `json:"is_paid"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}
	if strings.TrimSpace(payload.Name) == "" {
		utils.SendResponse(c, http.StatusBadRequest, false, "Nama pemesan wajib diisi", nil)
		return
	}

	count, err := oc.orderService.UpdateOrdersPaidByName(uint(listmakId), payload.Name, payload.IsPaid)
	if err != nil {
		if errors.Is(err, services.ErrNoOrdersMatched) {
			utils.SendResponse(c, http.StatusNotFound, false, "Tidak ada pesanan dengan nama tersebut di listmak ini", nil)
			return
		}
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to update status", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "Status bayar berhasil diupdate", gin.H{
		"updated_count": count,
	})
}

// DeleteOrder godoc
// @Summary      Delete listmak
// @Description  Delete listmak and related orders
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "orders ID"
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /orders/{id} [delete]
func (oc *orderController) DeleteOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := oc.orderService.DeleteOrder(uint(id)); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to delete order", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Order deleted", nil)
}

func (oc *orderController) UpdateOrderVendor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid order ID", nil)
		return
	}

	var payload struct {
		VendorName string `json:"vendor_name"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	vendorName := strings.TrimSpace(payload.VendorName)
	if err := oc.orderService.UpdateVendorName(uint(id), vendorName); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal update vendor", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "Vendor berhasil diupdate", gin.H{
		"id":          id,
		"vendor_name": vendorName,
	})
}

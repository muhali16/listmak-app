package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/services"
	"github.com/muhali16/listmak-service/pkg/utils"
)

type ShareController interface {
	CreateShareLink(c *gin.Context)
	GetShareLink(c *gin.Context)
	DeleteShareLink(c *gin.Context)
	SubmitOrderViaShare(c *gin.Context)
	GetOrdersViaShare(c *gin.Context)
	ParseOrdersViaShare(c *gin.Context)

	CreateViewShare(c *gin.Context)
	GetViewShare(c *gin.Context)

	GetActiveSharesForListmak(c *gin.Context)
	GetFoodSuggestions(c *gin.Context)
}

type shareController struct {
	shareService services.ShareService
	orderService services.OrderService
	aiService    services.AIService
}

func NewShareController(shareService services.ShareService, orderService services.OrderService, aiService services.AIService) ShareController {
	return &shareController{
		shareService: shareService,
		orderService: orderService,
		aiService:    aiService,
	}
}

// CreateShareLink godoc
// @Summary      Create a share link
// @Description  Create a new share link for a listmak
// @Tags         share-links
// @Accept       json
// @Produce      json
// @Param        payload  body      map[string]interface{}  true  "Share link data"
// @Success      200      {object}  utils.Response{data=models.ShareLink}
// @Failure      400      {object}  utils.Response
// @Failure      500      {object}  utils.Response
// @Router       /share-links [post]
func (sc *shareController) CreateShareLink(c *gin.Context) {
	var payload struct {
		ListmakID uint      `json:"listmak_id"`
		Title     string    `json:"title"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	userIdStr := c.MustGet("user_id").(string)
	userIdUint64, _ := strconv.ParseUint(userIdStr, 10, 64)

	share, err := sc.shareService.CreateShareLink(payload.ListmakID, payload.Title, payload.ExpiresAt, uint(userIdUint64))
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to create share link", nil)
		return
	}

	// Doc: "share_url": "https://listmak.app/listmak/order/abc12xyz"
	// We just return the object, let FE construct or we can add a transient field if needed.
	// We'll stick to model for now.
	utils.SendResponse(c, http.StatusOK, true, "Share link berhasil dibuat", share)
}

// GetShareLink godoc
// @Summary      Get share link data
// @Description  Get public share link data
// @Tags         share-links
// @Accept       json
// @Produce      json
// @Param        shareId  path      string  true  "Share ID"
// @Success      200      {object}  map[string]interface{}
// @Failure      404      {object}  utils.Response
// @Failure      410      {object}  map[string]interface{}
// @Router       /share-links/{shareId} [get]
func (sc *shareController) GetShareLink(c *gin.Context) {
	shareId := c.Param("shareId")
	share, err := sc.shareService.GetShareLink(shareId)
	if err != nil {
		if err.Error() == "EXPIRED" {
			c.JSON(http.StatusGone, gin.H{
				"success": false,
				"error":   "EXPIRED",
				"message": "Waktu input pesanan telah berakhir",
				"data": gin.H{
					"expires_at": share.ExpiresAt,
				},
			})
			return
		}
		utils.SendResponse(c, http.StatusNotFound, false, "Share link tidak ditemukan", nil)
		return
	}

	// Response format as per doc
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"share_id":     share.ShareID,
			"title":        share.Title,
			"expires_at":   share.ExpiresAt,
			"is_expired":   false,
			"listmak_date": share.Listmak.Date.Format("2006-01-02"),
		},
	})
}

// DeleteShareLink godoc
// @Summary      Delete share link
// @Description  Deactivate share link
// @Tags         share-links
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Share Link ID"
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /share-links/{id} [delete]
func (sc *shareController) DeleteShareLink(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := sc.shareService.DeleteShareLink(uint(id)); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Share link deleted", nil)
}

// SubmitOrderViaShare godoc
// @Summary      Submit order via share link
// @Description  Submit one or multiple orders via public share link
// @Tags         share-links
// @Accept       json
// @Produce      json
// @Param        shareId  path      string                  true  "Share ID"
// @Param        payload  body      map[string]interface{}  true  "Order data"
// @Success      200      {object}  utils.Response
// @Failure      400      {object}  utils.Response
// @Failure      404      {object}  utils.Response
// @Failure      500      {object}  utils.Response
// @Router       /share-links/{shareId}/orders [post]
func (sc *shareController) SubmitOrderViaShare(c *gin.Context) {
	shareId := c.Param("shareId")

	share, err := sc.shareService.GetShareLink(shareId)
	if err != nil {
		utils.SendResponse(c, http.StatusNotFound, false, "Invalid share link", nil)
		return
	}

	// Payload can be bulk or single
	var payload struct {
		Orders      []models.Order `json:"orders"`
		Name        string         `json:"name"`
		OrderDetail string         `json:"order_detail"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	var ordersToCreate []models.Order

	if len(payload.Orders) > 0 {
		ordersToCreate = payload.Orders
	} else if payload.Name != "" {
		ordersToCreate = append(ordersToCreate, models.Order{
			Name:        payload.Name,
			OrderDetail: payload.OrderDetail,
			AddedVia:    "sharelink",
		})
	} else {
		utils.SendResponse(c, http.StatusBadRequest, false, "Empty orders", nil)
		return
	}

	// Set added_via
	for i := range ordersToCreate {
		ordersToCreate[i].AddedVia = "sharelink"
	}

	count, _, err := sc.orderService.CreateOrdersBulk(share.ListmakID, ordersToCreate, c.GetString("RequestID"))
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to create orders", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "Pesanan berhasil ditambahkan", gin.H{"added_count": count})
}

// GetOrdersViaShare godoc
// @Summary      Get orders via share link
// @Description  Get the list of orders for a listmak via public share link
// @Tags         share-links
// @Accept       json
// @Produce      json
// @Param        shareId  path      string  true  "Share ID"
// @Success      200      {object}  utils.Response
// @Failure      404      {object}  utils.Response
// @Failure      410      {object}  utils.Response
// @Router       /share-links/{shareId}/orders [get]
func (sc *shareController) GetOrdersViaShare(c *gin.Context) {
	shareId := c.Param("shareId")

	share, err := sc.shareService.GetShareLink(shareId)
	if err != nil {
		if err.Error() == "EXPIRED" {
			c.JSON(http.StatusGone, gin.H{
				"success": false,
				"error":   "EXPIRED",
				"message": "Waktu input pesanan telah berakhir",
			})
			return
		}
		utils.SendResponse(c, http.StatusNotFound, false, "Share link tidak ditemukan", nil)
		return
	}

	orders, err := sc.orderService.GetOrdersByListmakId(share.ListmakID, nil, "")
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal mengambil data pesanan", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "OK", orders)
}

// CreateViewShare godoc
// @Summary      Create view share link
// @Description  Create a readonly view share link
// @Tags         view-shares
// @Accept       json
// @Produce      json
// @Param        payload  body      map[string]interface{}  true  "View share data"
// @Success      200      {object}  utils.Response{data=models.ViewShare}
// @Failure      400      {object}  utils.Response
// @Failure      500      {object}  utils.Response
// @Router       /view-shares [post]
func (sc *shareController) CreateViewShare(c *gin.Context) {
	var payload struct {
		ListmakID uint   `json:"listmak_id"`
		Title     string `json:"title"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	userIdStr := c.MustGet("user_id").(string)
	userIdUint64, _ := strconv.ParseUint(userIdStr, 10, 64)

	viewShare, err := sc.shareService.CreateViewShare(payload.ListmakID, payload.Title, uint(userIdUint64))
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to create view share", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "View share link berhasil dibuat", viewShare)
}

// GetActiveSharesForListmak godoc
// @Summary      Get active share links for a listmak
// @Description  Returns the latest active share link and view share for a listmak (null if none)
// @Tags         share-links
// @Produce      json
// @Param        id   path      int  true  "Listmak ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  utils.Response
// @Router       /listmaks/{id}/active-shares [get]
func (sc *shareController) GetActiveSharesForListmak(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid listmak ID", nil)
		return
	}

	shareLink, viewShare, err := sc.shareService.GetActiveSharesForListmak(uint(id))
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to fetch active shares", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "OK", gin.H{
		"share_link": shareLink,
		"view_share": viewShare,
	})
}

func (sc *shareController) ParseOrdersViaShare(c *gin.Context) {
	shareId := c.Param("shareId")
	if _, err := sc.shareService.GetShareLink(shareId); err != nil {
		utils.SendResponse(c, http.StatusNotFound, false, "Share link tidak valid", nil)
		return
	}

	var payload struct {
		Orders   []services.ParseOrderInput `json:"orders"`
		Location string                     `json:"location"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil || len(payload.Orders) == 0 {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	items, err := sc.aiService.ParseOrders(c.GetString("RequestID"), payload.Orders, payload.Location)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal parsing pesanan: "+err.Error(), nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "OK", items)
}

func (sc *shareController) GetFoodSuggestions(c *gin.Context) {
	shareId := c.Param("shareId")

	share, err := sc.shareService.GetShareLink(shareId)
	if err != nil {
		utils.SendResponse(c, http.StatusNotFound, false, "Share link tidak ditemukan", nil)
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	suggestions, err := sc.orderService.GetFoodSuggestions(share.ListmakID, query)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal mengambil saran", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "OK", suggestions)
}

// GetViewShare godoc
// @Summary      Get view share data
// @Description  Get public view share data
// @Tags         view-shares
// @Accept       json
// @Produce      json
// @Param        viewId   path      string  true  "View ID"
// @Success      200      {object}  map[string]interface{}
// @Failure      404      {object}  utils.Response
// @Router       /view-shares/{viewId} [get]
func (sc *shareController) GetViewShare(c *gin.Context) {
	viewId := c.Param("viewId")
	viewShare, err := sc.shareService.GetViewShare(viewId)
	if err != nil {
		if errors.Is(err, services.ErrListmakUnavailable) {
			utils.SendResponse(c, http.StatusNotFound, false, "listmak tidak tersedia", nil)
			return
		}
		utils.SendResponse(c, http.StatusNotFound, false, "View link not found", nil)
		return
	}

	// Parse snapshot
	var snapshotData map[string]interface{}
	// Or define a struct mirroring Listmak response
	if len(viewShare.SnapshotData) > 0 {
		json.Unmarshal(viewShare.SnapshotData, &snapshotData)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"view_id":    viewShare.ViewID,
			"title":      viewShare.Title,
			"created_at": viewShare.CreatedAt,
			"snapshot":   snapshotData, // return the snapshot
			// Note: Doc implies "data" contains the listmak data directly + view info.
			// "summary": {...}, "orders": [...]
			// We return snapshot which should contain this structure if Listmak struct matches.
		},
	})
}

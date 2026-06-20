package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/services"
	"github.com/muhali16/listmak-service/pkg/utils"
)

type ListmakController interface {
	GetListmaks(c *gin.Context)
	GetListmakById(c *gin.Context)
	GetListmakByDate(c *gin.Context)
	CreateListmak(c *gin.Context)
	UpdateListmak(c *gin.Context)
	DeleteListmak(c *gin.Context)
}

type listmakController struct {
	listmakService services.ListmakService
}

func NewListmakController(listmakService services.ListmakService) ListmakController {
	return &listmakController{listmakService: listmakService}
}

// GetListmaks godoc
// @Summary      Get all listmaks
// @Description  Get list of listmaks with filtering options
// @Tags         listmaks
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "Page number"
// @Param        limit      query     int     false  "Items per page"
// @Param        status     query     string  false  "Status filter"
// @Param        start_date query     string  false  "Start date (YYYY-MM-DD)"
// @Param        end_date   query     string  false  "End date (YYYY-MM-DD)"
// @Success      200  {object}  utils.Response{data=[]models.Listmak}
// @Failure      500  {object}  utils.Response
// @Router       /listmaks [get]
func (lc *listmakController) GetListmaks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	var startDate, endDate *time.Time
	if sDate := c.Query("start_date"); sDate != "" {
		t, err := time.Parse("2006-01-02", sDate)
		if err == nil {
			startDate = &t
		}
	}
	if eDate := c.Query("end_date"); eDate != "" {
		t, err := time.Parse("2006-01-02", eDate)
		if err == nil {
			endDate = &t
		}
	}

	userIdStr := c.MustGet("user_id").(string)
	userIdUint64, _ := strconv.ParseUint(userIdStr, 10, 64)

	data, total, err := lc.listmakService.GetAllListmaks(page, limit, status, startDate, endDate, uint(userIdUint64))
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

// GetListmakById godoc
// @Summary      Get listmak by ID
// @Description  Get detailed listmak information
// @Tags         listmaks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Listmak ID"
// @Success      200  {object}  utils.Response{data=models.Listmak}
// @Failure      404  {object}  utils.Response
// @Router       /listmaks/{id} [get]
func (lc *listmakController) GetListmakById(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := lc.listmakService.GetListmakById(uint(id))
	if err != nil {
		utils.SendResponse(c, http.StatusNotFound, false, "Listmak not found", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Success get listmak", data)
}

// GetListmakByDate godoc
// @Summary      Get listmak by Date
// @Description  Get ALL listmaks based on date
// @Tags         listmaks
// @Accept       json
// @Produce      json
// @Param        date path      string  true  "Date (YYYY-MM-DD)"
// @Success      200  {object}  utils.Response{data=[]models.Listmak}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /listmaks/date/{date} [get]
func (lc *listmakController) GetListmakByDate(c *gin.Context) {
	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid date format", nil)
		return
	}

	userIdStr := c.MustGet("user_id").(string)
	userIdUint64, _ := strconv.ParseUint(userIdStr, 10, 64)

	data, err := lc.listmakService.GetListmakByDate(date, uint(userIdUint64))
	if err != nil {
		utils.SendResponse(c, http.StatusNotFound, false, "Listmak not found", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Success get listmaks", data)
}

// CreateListmak godoc
// @Summary      Create listmak
// @Description  Create a new listmak
// @Tags         listmaks
// @Accept       json
// @Produce      json
// @Param        listmak  body      models.Listmak  true  "Listmak data"
// @Success      200      {object}  utils.Response{data=models.Listmak}
// @Failure      400      {object}  utils.Response
// @Failure      500      {object}  utils.Response
// @Router       /listmaks [post]
func (lc *listmakController) CreateListmak(c *gin.Context) {
	var listmak models.Listmak
	if err := c.ShouldBindJSON(&listmak); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid request body", nil)
		return
	}

	userIdStr := c.MustGet("user_id").(string)
	userIdUint64, _ := strconv.ParseUint(userIdStr, 10, 64)
	userId := uint(userIdUint64)
	listmak.CreatedBy = &userId

	created, err := lc.listmakService.CreateListmak(listmak)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to create listmak", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "Listmak berhasil dibuat", created)
}

// UpdateListmak godoc
// @Summary      Update listmak
// @Description  Update listmak details
// @Tags         listmaks
// @Accept       json
// @Produce      json
// @Param        id       path      int             true  "Listmak ID"
// @Param        listmak  body      models.Listmak  true  "Update data"
// @Success      200      {object}  utils.Response{data=models.Listmak}
// @Failure      400      {object}  utils.Response
// @Failure      404      {object}  utils.Response
// @Failure      500      {object}  utils.Response
// @Router       /listmaks/{id} [put]
func (lc *listmakController) UpdateListmak(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var payload models.Listmak
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid request body", nil)
		return
	}

	listmak, err := lc.listmakService.GetListmakById(uint(id))
	if err != nil {
		utils.SendResponse(c, http.StatusNotFound, false, "Listmak not found", nil)
		return
	}

	// Update fields
	if payload.Title != "" {
		listmak.Title = payload.Title
	}
	if payload.Status != "" {
		listmak.Status = payload.Status
	}
	// Add more fields if needed

	updated, err := lc.listmakService.UpdateListmak(listmak)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to update listmak", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Listmak updated", updated)
}

// DeleteListmak godoc
// @Summary      Delete listmak
// @Description  Delete listmak and related orders
// @Tags         listmaks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Listmak ID"
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /listmaks/{id} [delete]
func (lc *listmakController) DeleteListmak(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := lc.listmakService.DeleteListmak(uint(id)); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to delete listmak", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Listmak deleted", nil)
}

package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/repository"
	"github.com/muhali16/listmak-service/pkg/utils"
)

type AdminController interface {
	GetAILogs(c *gin.Context)
	GetSystemLogs(c *gin.Context)
	UpdateUserRole(c *gin.Context)
}

type adminController struct {
	aiLogRepo     repository.AILogRepository
	systemLogRepo repository.SystemLogRepository
	userRepo      repository.UserRepository
}

func NewAdminController(aiLogRepo repository.AILogRepository, systemLogRepo repository.SystemLogRepository, userRepo repository.UserRepository) AdminController {
	return &adminController{aiLogRepo: aiLogRepo, systemLogRepo: systemLogRepo, userRepo: userRepo}
}

func (ac *adminController) GetAILogs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	logs, total, err := ac.aiLogRepo.GetAll(page, 50)
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

	// default from: 7 days ago
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

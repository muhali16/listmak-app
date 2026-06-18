package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/services"
	"github.com/muhali16/listmak-service/pkg/utils"
)

// GetAllLogs godoc
// @Summary      Get all logs
// @Tags         logs
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]models.SystemLog}
// @Failure      500  {object}  utils.Response
// @Router       /logs [get]
func GetAllLogs(c *gin.Context) {
	filters := make(map[string]interface{})
	if c.Query("request_id") != "" {
		filters["request_id"] = c.Query("request_id")
	}
	if c.Query("method") != "" {
		filters["method"] = c.Query("method")
	}
	if c.Query("path") != "" {
		filters["path"] = c.Query("path")
	}
	if c.Query("status_code") != "" {
		filters["status_code"] = c.Query("status_code")
	}
	if c.Query("client_ip") != "" {
		filters["client_ip"] = c.Query("client_ip")
	}
	logs := services.GetAllLogs(filters)
	utils.SendResponse(c, http.StatusOK, true, "Success get logs", logs)
}

// GetLogByRequestID godoc
// @Summary      Get log by request id
// @Tags         logs
// @Accept       json
// @Produce      json
// @Param        request_id  path    string  true  "Request ID"
// @Success      200  {object}  utils.Response{data=models.SystemLog}
// @Failure      500  {object}  utils.Response
// @Router       /logs/{request_id} [get]
func GetLogByRequestID(c *gin.Context) {
	requestID := c.Param("request_id")
	log := services.GetLogByRequestID(requestID)
	utils.SendResponse(c, http.StatusOK, true, "Success get log", log)
}

package utils

import (
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code      int    `json:"code"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Latency   string `json:"latency,omitempty"`
}

func SendResponse(c *gin.Context, code int, success bool, message string, data any) {
	reqID, exists := c.Get("RequestID")
	if !exists {
		reqID = ""
	}

	var latencyStr string
	if startTime, ok := c.Get("StartTime"); ok {
		if t, valid := startTime.(time.Time); valid {
			latencyStr = time.Since(t).String()
		}
	}

	c.JSON(code, Response{
		Code:      code,
		Success:   success,
		Message:   message,
		Data:      data,
		RequestID: reqID.(string),
		Latency:   latencyStr,
	})
}

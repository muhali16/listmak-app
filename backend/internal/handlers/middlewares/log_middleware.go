package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
)

func LoggerWithID(logRepo repository.SystemLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := uuid.New().String()

		c.Header("X-Request-ID", requestID)
		c.Set("RequestID", requestID)
		c.Set("StartTime", start)
		c.Next()

		logEntry := models.SystemLog{
			RequestID:  requestID,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			Latency:    time.Since(start).String(),
			ClientIP:   c.ClientIP(),
			ErrorMsg:   c.Errors.ByType(gin.ErrorTypeAny).String(),
		}

		go logRepo.Create(&logEntry)
	}
}

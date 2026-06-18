package services

import (
	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
)

func GetAllLogs(filters map[string]interface{}) []models.SystemLog {
	return repository.GetAllLogs(filters)
}

func GetLogByRequestID(requestID string) models.SystemLog {
	return repository.GetLogByRequestID(requestID)
}

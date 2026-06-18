package repository

import "github.com/muhali16/listmak-service/internal/models"

func GetAllLogs(filters map[string]interface{}) []models.SystemLog {
	var logs []models.SystemLog
	models.DBLog.Where(filters).Order("created_at desc").Limit(100).Find(&logs)
	return logs
}

func GetLogByRequestID(requestID string) models.SystemLog {
	var log models.SystemLog
	models.DBLog.Where("request_id = ?", requestID).First(&log)
	return log
}

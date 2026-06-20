package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
)

type AIService interface {
	ExtractVendor(orderDetail string, orderID *uint) (string, error)
}

type fireworksAIService struct {
	apiKey  string
	model   string
	client  *http.Client
	logRepo repository.AILogRepository
}

func NewFireworksAIService(apiKey, model string, logRepo repository.AILogRepository) AIService {
	return &fireworksAIService{
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{Timeout: 10 * time.Second},
		logRepo: logRepo,
	}
}

type noopAIService struct{}

func NewNoopAIService() AIService { return &noopAIService{} }

func (n *noopAIService) ExtractVendor(_ string, _ *uint) (string, error) { return "", nil }

func (s *fireworksAIService) ExtractVendor(orderDetail string, orderID *uint) (string, error) {
	start := time.Now()

	prompt := fmt.Sprintf(
		"Dari pesanan makanan berikut, ekstrak nama resto atau kategori tempatnya. "+
			"Jika ada nama tempat atau nama orang (contoh: 'pak donan', 'warung bu sari') gunakan itu. "+
			"Jika tidak ada nama tempat eksplisit, infer dari jenis makanannya "+
			"(contoh: 'soto daging' → 'Soto', 'dimsum' → 'Dimsum', 'ayam madura' → 'Ayam Madura'). "+
			"Jawab HANYA nama tempatnya saja, tanpa penjelasan, tanpa tanda kutip, singkat. "+
			"Pesanan: %s", orderDetail,
	)

	reqBody, _ := json.Marshal(map[string]any{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  30,
		"temperature": 0,
	})

	req, err := http.NewRequest("POST", "https://api.fireworks.ai/inference/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		s.writeLog(orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.writeLog(orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("fireworks API %d: %s", resp.StatusCode, string(b))
		s.writeLog(orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", errMsg)
		return "", fmt.Errorf("%s", errMsg)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.writeLog(orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return "", err
	}
	if len(result.Choices) == 0 {
		s.writeLog(orderID, orderDetail, "", time.Since(start).Milliseconds(), "success", "")
		return "", nil
	}

	output := strings.TrimSpace(result.Choices[0].Message.Content)
	s.writeLog(orderID, orderDetail, output, time.Since(start).Milliseconds(), "success", "")
	return output, nil
}

func (s *fireworksAIService) writeLog(orderID *uint, input, output string, latencyMs int64, status, errMsg string) {
	if s.logRepo == nil {
		return
	}
	s.logRepo.Create(&models.AILog{
		OrderID:   orderID,
		Input:     input,
		Output:    output,
		Model:     s.model,
		Provider:  "fireworks",
		LatencyMs: latencyMs,
		Status:    status,
		ErrorMsg:  errMsg,
	})
}

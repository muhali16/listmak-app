package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AIService interface {
	ExtractVendor(orderDetail string) (string, error)
}

type fireworksAIService struct {
	apiKey string
	model  string
	client *http.Client
}

func NewFireworksAIService(apiKey, model string) AIService {
	return &fireworksAIService{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type noopAIService struct{}

func NewNoopAIService() AIService { return &noopAIService{} }

func (n *noopAIService) ExtractVendor(_ string) (string, error) { return "", nil }

func (s *fireworksAIService) ExtractVendor(orderDetail string) (string, error) {
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
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("fireworks API %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", nil
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

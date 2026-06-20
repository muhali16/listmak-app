package services

import (
	"bytes"
	"context"
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
	ExtractVendorsBatch(orders []models.Order) (map[uint]string, error)
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
func (n *noopAIService) ExtractVendorsBatch(_ []models.Order) (map[uint]string, error) {
	return map[uint]string{}, nil
}

func (s *fireworksAIService) ExtractVendor(orderDetail string, orderID *uint) (string, error) {
	start := time.Now()

	prompt := fmt.Sprintf(
		"Tugas Anda adalah mengekstrak nama tempat atau kategori vendor dari pesanan makanan.\n\n"+
			"Aturan:\n"+
			"1. Jika ada nama orang/identitas unik yang digabung dengan jenis makanan (contoh: 'Bakso bakar pak donan', 'mie ayam bu nun'), ambil kombinasi makanan + nama tersebut menjadi 'Bakso Bakar Pak Donan' atau 'Mie Ayam Bu Nun'.\n"+
			"2. Jika berupa jajanan umum/kaki lima tanpa nama penjual spesifik (contoh: 'Cilok 5 tusuk', 'telur gulung'), cukup ambil nama jenis makanannya saja menjadi 'Cilok' atau 'Telur Gulung'.\n"+
			"3. Jika ada nama daerah/gaya masakan (contoh: 'ayam madura dada', 'nasi goreng madura'), ambil menjadi 'Ayam Madura' atau 'Nasi Goreng Madura'.\n\n"+
			"Format Output: Jawab HANYA nama tempat/vendor hasil ekstraksi tersebut, singkat, gunakan huruf kapital di awal kata (Title Case), tanpa tanda kutip, dan TANPA penjelasan apapun.\n\n"+
			"Pesanan: %s", orderDetail,
	)

	reqBody, _ := json.Marshal(map[string]any{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":       30,
		"temperature":      0,
		"reasoning_effort": "none",
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

func (s *fireworksAIService) ExtractVendorsBatch(orders []models.Order) (map[uint]string, error) {
	if len(orders) == 0 {
		return map[uint]string{}, nil
	}

	details := make([]string, len(orders))
	for i, o := range orders {
		details[i] = o.OrderDetail
	}
	detailsJSON, _ := json.Marshal(details)

	prompt := fmt.Sprintf(
		"Ekstrak nama vendor dari setiap pesanan makanan berikut.\n\n"+
			"Aturan:\n"+
			"1. Jika ada nama orang unik + jenis makanan (contoh: 'bakso bakar pak donan') → ambil kombinasinya: 'Bakso Bakar Pak Donan'\n"+
			"2. Jajanan umum tanpa nama penjual (contoh: 'cilok 5 tusuk') → nama jenis saja: 'Cilok'\n"+
			"3. Ada nama daerah/gaya masakan (contoh: 'ayam madura dada') → sertakan daerahnya: 'Ayam Madura'\n\n"+
			"Output: JSON array string SESUAI URUTAN input, Title Case, tanpa penjelasan, tanpa markdown.\n"+
			"Contoh input: [\"bakso bakar pak donan\",\"cilok 5 tusuk\"]\n"+
			"Contoh output: [\"Bakso Bakar Pak Donan\",\"Cilok\"]\n\n"+
			"Input: %s", string(detailsJSON),
	)

	reqBody, _ := json.Marshal(map[string]any{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":       1024,
		"temperature":      0,
		"reasoning_effort": "none",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.fireworks.ai/inference/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fireworks API %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return map[uint]string{}, nil
	}

	raw := strings.TrimSpace(result.Choices[0].Message.Content)
	// strip markdown fences if model wraps in ```json ... ```
	if idx := strings.Index(raw, "["); idx > 0 {
		raw = raw[idx:]
	}
	if idx := strings.LastIndex(raw, "]"); idx >= 0 && idx < len(raw)-1 {
		raw = raw[:idx+1]
	}

	var vendors []string
	if err := json.Unmarshal([]byte(raw), &vendors); err != nil {
		return map[uint]string{}, nil
	}

	out := make(map[uint]string, len(orders))
	for i, o := range orders {
		if i < len(vendors) {
			out[o.ID] = strings.TrimSpace(vendors[i])
		}
	}
	return out, nil
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

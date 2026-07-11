package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// PakasirClient wraps the Pakasir payment-link API (create transaction + query
// detail). See backend/docs/PAYMENT_GATEWAY_DOCS.md. Pakasir provides NO webhook
// signature, so the transaction-detail query is the authoritative status source.
type PakasirClient interface {
	Configured() bool
	Project() string
	// PayURL builds the hosted payment-link URL (fallback / QRIS render helper).
	PayURL(orderID string, amount int64) string
	CreateTransaction(method, orderID string, amount int64) (*PakasirPayment, error)
	GetTransactionDetail(orderID string, amount int64) (*PakasirTransaction, error)
	// CancelTransaction cancels a still-pending transaction so it stops occupying
	// a payment slot. Fails on the gateway side if payment already arrived.
	CancelTransaction(orderID string, amount int64) error
}

type PakasirPayment struct {
	Project       string    `json:"project"`
	OrderID       string    `json:"order_id"`
	Amount        int64     `json:"amount"`
	Fee           int64     `json:"fee"`
	TotalPayment  int64     `json:"total_payment"`
	PaymentMethod string    `json:"payment_method"`
	PaymentNumber string    `json:"payment_number"`
	ExpiredAt     time.Time `json:"expired_at"`
	IsSandbox     bool      `json:"is_sandbox"`
}

type PakasirTransaction struct {
	Amount        int64  `json:"amount"`
	OrderID       string `json:"order_id"`
	Project       string `json:"project"`
	Status        string `json:"status"`
	PaymentMethod string `json:"payment_method"`
	CompletedAt   string `json:"completed_at"`
	IsSandbox     bool   `json:"is_sandbox"`
}

// PaymentLogFunc records one gateway interaction. response is the raw Pakasir
// body (safe to store); the outbound request is never passed (carries api_key).
type PaymentLogFunc func(orderID, action string, statusCode int, response []byte, errMsg string)

type pakasirClient struct {
	baseURL string
	project string
	apiKey  string
	http    *http.Client
	logFn   PaymentLogFunc
}

func NewPakasirClient(baseURL, project, apiKey string, logFn PaymentLogFunc) PakasirClient {
	if baseURL == "" {
		baseURL = "https://app.pakasir.com"
	}
	return &pakasirClient{
		baseURL: baseURL,
		project: project,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 10 * time.Second},
		logFn:   logFn,
	}
}

func (p *pakasirClient) log(orderID, action string, code int, raw []byte, err error) {
	if p.logFn == nil {
		return
	}
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	p.logFn(orderID, action, code, raw, errMsg)
}

func (p *pakasirClient) Configured() bool { return p.project != "" && p.apiKey != "" }
func (p *pakasirClient) Project() string  { return p.project }

func (p *pakasirClient) PayURL(orderID string, amount int64) string {
	return fmt.Sprintf("%s/pay/%s/%d?order_id=%s", p.baseURL, p.project, amount, url.QueryEscape(orderID))
}

func (p *pakasirClient) CreateTransaction(method, orderID string, amount int64) (*PakasirPayment, error) {
	body, _ := json.Marshal(map[string]any{
		"project":  p.project,
		"order_id": orderID,
		"amount":   amount,
		"api_key":  p.apiKey,
	})

	endpoint := fmt.Sprintf("%s/api/transactioncreate/%s", p.baseURL, url.PathEscape(method))
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	raw, code, err := p.do(req)
	p.log(orderID, "create", code, raw, err)
	if err != nil {
		return nil, err
	}

	var out struct {
		Payment PakasirPayment `json:"payment"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("pakasir: decode create response: %w", err)
	}
	if out.Payment.OrderID == "" {
		return nil, fmt.Errorf("pakasir: empty create response: %s", string(raw))
	}
	return &out.Payment, nil
}

func (p *pakasirClient) CancelTransaction(orderID string, amount int64) error {
	body, _ := json.Marshal(map[string]any{
		"project":  p.project,
		"order_id": orderID,
		"amount":   amount,
		"api_key":  p.apiKey,
	})
	req, err := http.NewRequest(http.MethodPost, p.baseURL+"/api/transactioncancel", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	raw, code, err := p.do(req)
	p.log(orderID, "cancel", code, raw, err)
	return err
}

func (p *pakasirClient) GetTransactionDetail(orderID string, amount int64) (*PakasirTransaction, error) {
	// NOTE: the API requires `amount` too (the public docs omit it — verified live).
	endpoint := fmt.Sprintf("%s/api/transactiondetail?project=%s&order_id=%s&amount=%d&api_key=%s",
		p.baseURL, url.QueryEscape(p.project), url.QueryEscape(orderID), amount, url.QueryEscape(p.apiKey))

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	raw, code, err := p.do(req)
	p.log(orderID, "detail", code, raw, err)
	if err != nil {
		return nil, err
	}

	var out struct {
		Transaction PakasirTransaction `json:"transaction"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("pakasir: decode detail response: %w", err)
	}
	if out.Transaction.OrderID == "" {
		return nil, errors.New("pakasir: transaction not found")
	}
	return &out.Transaction, nil
}

// do returns the raw body and HTTP status. On a non-2xx it also returns an error,
// but still yields the body + code so callers can log the failure detail.
func (p *pakasirClient) do(req *http.Request) ([]byte, int, error) {
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return raw, resp.StatusCode, fmt.Errorf("pakasir: status %d: %s", resp.StatusCode, string(raw))
	}
	return raw, resp.StatusCode, nil
}

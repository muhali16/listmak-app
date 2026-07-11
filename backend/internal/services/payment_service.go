package services

import (
	"encoding/json"
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
	"github.com/muhali16/listmak-service/pkg/utils"
)

var (
	ErrPaymentNotConfigured = errors.New("payment gateway not configured")
	ErrShareInvalid         = errors.New("share link invalid")
	ErrShareExpired         = errors.New("share link expired")
	ErrInvalidGuest         = errors.New("guest name and whatsapp are required")
	ErrEmptyOrders          = errors.New("at least one order is required")
	ErrInvalidAmount        = errors.New("total amount must be greater than zero")
	ErrInvalidMethod        = errors.New("unsupported payment method")
	ErrTooManyItems         = errors.New("too many order lines")
	ErrAlreadyPaid          = errors.New("payment already completed")
)

// Input caps — bound abuse/overflow on the money path. Generous for real food
// orders, absurd values rejected.
const (
	maxCheckoutItems = 50
	maxNameLen       = 100
	maxDetailLen     = 500
	maxQty           = 1000
	maxLinePrice     = 10_000_000  // per-line unit price (IDR)
	maxTotalAmount   = 100_000_000 // total per checkout (IDR)
)

// pakasirMethods is the allowlist from the integration docs (Section C.3).
var pakasirMethods = map[string]bool{
	"qris": true, "cimb_niaga_va": true, "bni_va": true, "sampoerna_va": true,
	"bnc_va": true, "maybank_va": true, "permata_va": true, "atm_bersama_va": true,
	"artha_graha_va": true, "bri_va": true,
}

var waPattern = regexp.MustCompile(`^[0-9]{8,15}$`)

// CheckoutItem is one line a guest submits at checkout. Name is the person for
// that order line (may differ per line in bulk submissions); it falls back to
// the payer name when empty. The payer identity lives on the Payment.
type CheckoutItem struct {
	Name        string  `json:"name"`
	OrderDetail string  `json:"order_detail"`
	VendorName  string  `json:"vendor_name"`
	Price       float64 `json:"price"`
	Qty         int     `json:"qty"`
}

type CheckoutInput struct {
	ShareID       string
	GuestName     string
	GuestWhatsapp string
	PaymentMethod string
	Items         []CheckoutItem
}

// WebhookPayload mirrors Pakasir's webhook body.
type WebhookPayload struct {
	Amount        int64  `json:"amount"`
	OrderID       string `json:"order_id"`
	Project       string `json:"project"`
	Status        string `json:"status"`
	PaymentMethod string `json:"payment_method"`
	CompletedAt   string `json:"completed_at"`
}

type PaymentService interface {
	Checkout(in CheckoutInput) (models.Payment, error)
	GetStatus(orderID string) (models.Payment, error)
	// Cancel voids a still-pending payment (guest closed the QR).
	Cancel(orderID string) error
	// HandleWebhook returns the HTTP status code Pakasir should receive.
	HandleWebhook(payload WebhookPayload) (int, error)
	// ReconcileStalePending settles pending payments older than olderThan:
	// complete if actually paid, otherwise cancel. Returns how many were handled.
	ReconcileStalePending(olderThan time.Duration) (int, error)
}

type paymentService struct {
	paymentRepo  repository.PaymentRepository
	orderService OrderService
	shareService ShareService
	pakasir      PakasirClient
	logFn        PaymentLogFunc
	appConfig    AppConfig
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	orderService OrderService,
	shareService ShareService,
	pakasir PakasirClient,
	logFn PaymentLogFunc,
	appConfig AppConfig,
) PaymentService {
	return &paymentService{
		paymentRepo:  paymentRepo,
		orderService: orderService,
		shareService: shareService,
		pakasir:      pakasir,
		logFn:        logFn,
		appConfig:    appConfig,
	}
}

func (s *paymentService) log(orderID, action string, statusCode int, response []byte, errMsg string) {
	if s.logFn != nil {
		s.logFn(orderID, action, statusCode, response, errMsg)
	}
}

func (s *paymentService) Checkout(in CheckoutInput) (models.Payment, error) {
	if !s.pakasir.Configured() {
		return models.Payment{}, ErrPaymentNotConfigured
	}

	// Validate the share context (existence + expiry) — this is the trust boundary
	// for an unauthenticated guest.
	share, err := s.shareService.GetShareLink(in.ShareID)
	if err != nil {
		if err.Error() == "EXPIRED" {
			return models.Payment{}, ErrShareExpired
		}
		return models.Payment{}, ErrShareInvalid
	}

	name := strings.TrimSpace(in.GuestName)
	wa := normalizeWhatsapp(in.GuestWhatsapp)
	if name == "" || len(name) > maxNameLen || !waPattern.MatchString(wa) {
		return models.Payment{}, ErrInvalidGuest
	}

	if len(in.Items) == 0 {
		return models.Payment{}, ErrEmptyOrders
	}
	if len(in.Items) > maxCheckoutItems {
		return models.Payment{}, ErrTooManyItems
	}

	method := strings.TrimSpace(in.PaymentMethod)
	if method == "" {
		method = "qris"
	}
	if !pakasirMethods[method] {
		return models.Payment{}, ErrInvalidMethod
	}

	// Build the item snapshot and compute the amount server-side. Never trust a
	// client-supplied total. Orders are materialized only when the payment
	// completes, so an unpaid/cancelled checkout never enters the order list.
	var items []models.PaymentItem
	var amountFloat float64
	for _, it := range in.Items {
		detail := strings.TrimSpace(it.OrderDetail)
		if detail == "" {
			continue
		}
		detail = truncate(detail, maxDetailLen)
		qty := it.Qty
		if qty <= 0 {
			qty = 1
		}
		if qty > maxQty {
			return models.Payment{}, ErrInvalidAmount
		}
		if it.Price < 0 || it.Price > maxLinePrice {
			return models.Payment{}, ErrInvalidAmount
		}
		amountFloat += it.Price * float64(qty)

		orderName := strings.TrimSpace(it.Name)
		if orderName == "" {
			orderName = name // fall back to payer name
		}
		orderName = truncate(orderName, maxNameLen)
		vendor := truncate(strings.TrimSpace(it.VendorName), maxNameLen)
		items = append(items, models.PaymentItem{
			Name:        orderName,
			OrderDetail: detail,
			VendorName:  vendor,
			Price:       it.Price,
			Qty:         qty,
		})
	}
	if len(items) == 0 {
		return models.Payment{}, ErrEmptyOrders
	}

	amount := int64(math.Round(amountFloat))
	if amount <= 0 || amount > maxTotalAmount {
		return models.Payment{}, ErrInvalidAmount
	}

	snapshot, err := json.Marshal(items)
	if err != nil {
		return models.Payment{}, err
	}

	// Generate a unique invoice id (== Pakasir order_id), retry on collision.
	var invoiceID string
	for range 3 {
		suffix, err := utils.GenerateRandomString(8)
		if err != nil {
			return models.Payment{}, err
		}
		candidate := "LMK" + strconv.FormatUint(uint64(share.ListmakID), 10) + "-" + suffix
		exists, err := s.paymentRepo.ExistsByOrderID(candidate)
		if err != nil {
			return models.Payment{}, err
		}
		if !exists {
			invoiceID = candidate
			break
		}
	}
	if invoiceID == "" {
		return models.Payment{}, errors.New("failed to generate unique invoice id")
	}

	// Create the gateway transaction first — if this fails we persist nothing.
	pakResp, err := s.pakasir.CreateTransaction(method, invoiceID, amount)
	if err != nil {
		return models.Payment{}, err
	}

	var expiresAt *time.Time
	if !pakResp.ExpiredAt.IsZero() {
		expiresAt = &pakResp.ExpiredAt
	}
	payment := models.Payment{
		OrderID:       invoiceID,
		ListmakID:     share.ListmakID,
		GuestName:     name,
		GuestWhatsapp: wa,
		Amount:        amount,
		Fee:           pakResp.Fee,
		TotalPayment:  pakResp.TotalPayment,
		PaymentMethod: pakResp.PaymentMethod,
		PaymentNumber: pakResp.PaymentNumber,
		Status:        "pending",
		ExpiresAt:     expiresAt,
		ItemsSnapshot: snapshot,
		// Sandbox if the gateway says so OR the app is in testing mode.
		IsSandbox: pakResp.IsSandbox || s.appConfig.TestingMode(),
	}

	// Persist a pending payment only. Orders are created when the payment
	// completes, keeping unpaid/cancelled checkouts out of the order list.
	created, err := s.paymentRepo.Create(payment)
	if err != nil {
		return models.Payment{}, err
	}
	return created, nil
}

func (s *paymentService) GetStatus(orderID string) (models.Payment, error) {
	payment, err := s.paymentRepo.GetByOrderID(orderID)
	if err != nil {
		return models.Payment{}, err
	}
	if payment.Status == "completed" {
		return payment, nil
	}
	if !s.pakasir.Configured() {
		return payment, nil
	}

	// Poll the authoritative source; promote to completed if Pakasir confirms.
	detail, err := s.pakasir.GetTransactionDetail(orderID, payment.Amount)
	if err != nil {
		return payment, nil // gateway unreachable: return last-known status
	}
	if detail.Status == "completed" && detail.Amount == payment.Amount {
		updated, _, err := s.paymentRepo.MarkCompleted(orderID, detail.PaymentMethod, parseCompletedAt(detail.CompletedAt), detail.IsSandbox)
		if err != nil {
			return payment, nil
		}
		s.orderService.RecalcListmakTotals(updated.ListmakID)
		s.log(orderID, "complete", 200, mustJSON(updated), "")
		return updated, nil
	}
	return payment, nil
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func (s *paymentService) Cancel(orderID string) error {
	payment, err := s.paymentRepo.GetByOrderID(orderID)
	if err != nil {
		return err
	}
	switch payment.Status {
	case "cancelled":
		return nil // idempotent
	case "completed":
		return ErrAlreadyPaid
	}

	// Void on the gateway first; if it already received payment it will refuse,
	// and we must not mark it cancelled locally.
	if s.pakasir.Configured() {
		if err := s.pakasir.CancelTransaction(orderID, payment.Amount); err != nil {
			return err
		}
	}
	// Cancel the pending payment (no orders exist yet). Only affects pending rows,
	// guarding against a race with completion.
	return s.paymentRepo.MarkCancelled(orderID)
}

func (s *paymentService) ReconcileStalePending(olderThan time.Duration) (int, error) {
	if !s.pakasir.Configured() {
		return 0, nil
	}
	stale, err := s.paymentRepo.ListStalePending(time.Now().Add(-olderThan))
	if err != nil {
		return 0, err
	}
	for _, p := range stale {
		// GetStatus completes it if Pakasir confirms payment (never lose real money).
		updated, err := s.GetStatus(p.OrderID)
		if err != nil {
			continue
		}
		// Still unpaid after checking the authoritative source -> void it so the
		// QR (and any saved screenshot) can no longer be paid.
		if updated.Status == "pending" {
			_ = s.Cancel(p.OrderID)
		}
	}
	return len(stale), nil
}

func (s *paymentService) HandleWebhook(payload WebhookPayload) (code int, retErr error) {
	// Audit every webhook delivery with its raw payload + outcome.
	defer func() {
		errMsg := ""
		if retErr != nil {
			errMsg = retErr.Error()
		}
		raw, _ := json.Marshal(payload)
		s.log(payload.OrderID, "webhook", code, raw, errMsg)
	}()

	// 1. Project must match our configured project.
	if !s.pakasir.Configured() || payload.Project != s.pakasir.Project() {
		return 400, errors.New("project mismatch")
	}

	// 2. Order must exist.
	payment, err := s.paymentRepo.GetByOrderID(payload.OrderID)
	if err != nil {
		return 404, errors.New("unknown order")
	}

	// 3. Idempotent: already processed.
	if payment.Status == "completed" {
		return 200, nil
	}

	// 4. Amount tampering check against our stored record.
	if payload.Amount != payment.Amount {
		return 400, errors.New("amount mismatch")
	}

	// 5. Authoritative status via Transaction Detail API (no webhook signature exists).
	detail, err := s.pakasir.GetTransactionDetail(payload.OrderID, payment.Amount)
	if err != nil {
		return 202, err // accepted, retry later
	}
	if detail.Status != "completed" || detail.Amount != payment.Amount {
		return 202, nil
	}

	// 6. Mark completed under row lock (idempotent) + resync listmak totals.
	updated, alreadyDone, err := s.paymentRepo.MarkCompleted(payload.OrderID, detail.PaymentMethod, parseCompletedAt(detail.CompletedAt), detail.IsSandbox)
	if err != nil {
		return 500, err
	}
	if !alreadyDone {
		s.orderService.RecalcListmakTotals(updated.ListmakID)
		s.log(payload.OrderID, "complete", 200, mustJSON(updated), "")
	}
	return 200, nil
}

// normalizeWhatsapp strips spaces, dashes and a leading + or 0, prefixing 62.
// e.g. "0812-3456-7890" -> "6281234567890", "+62 812 3456" -> "62812345678".
func normalizeWhatsapp(raw string) string {
	s := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(raw))
	s = strings.TrimPrefix(s, "+")
	if strings.HasPrefix(s, "0") {
		s = "62" + strings.TrimPrefix(s, "0")
	}
	return s
}

// truncate caps s to at most n runes (never splitting a multibyte character).
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func parseCompletedAt(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Now()
}

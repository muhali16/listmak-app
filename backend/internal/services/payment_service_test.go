package services

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
)

// --- fakes ---

type fakePakasir struct {
	project string
	detail  *PakasirTransaction
	err     error
}

func (f *fakePakasir) Configured() bool { return true }
func (f *fakePakasir) Project() string  { return f.project }
func (f *fakePakasir) PayURL(string, int64) string {
	return ""
}
func (f *fakePakasir) CreateTransaction(method, orderID string, amount int64) (*PakasirPayment, error) {
	// Echo the request back as a successful QRIS transaction.
	return &PakasirPayment{
		OrderID:       orderID,
		Amount:        amount,
		TotalPayment:  amount,
		PaymentMethod: method,
		PaymentNumber: "QRIS-" + orderID,
	}, nil
}
func (f *fakePakasir) GetTransactionDetail(string, int64) (*PakasirTransaction, error) {
	return f.detail, f.err
}
func (f *fakePakasir) CancelTransaction(string, int64) error { return nil }

type fakePaymentRepo struct {
	payment     models.Payment
	completedN  int
	alreadyDone bool
}

func (r *fakePaymentRepo) Create(p models.Payment) (models.Payment, error) {
	return p, nil
}
func (r *fakePaymentRepo) GetByOrderID(orderID string) (models.Payment, error) {
	if r.payment.OrderID != orderID {
		return models.Payment{}, errors.New("not found")
	}
	return r.payment, nil
}
func (r *fakePaymentRepo) ExistsByOrderID(string) (bool, error) { return false, nil }
func (r *fakePaymentRepo) ListForAdmin(int, int, string, string) ([]models.Payment, int64, error) {
	return nil, 0, nil
}
func (r *fakePaymentRepo) Stats() (repository.PaymentStats, error) {
	return repository.PaymentStats{}, nil
}
func (r *fakePaymentRepo) MarkCancelled(string) error {
	r.payment.Status = "cancelled"
	return nil
}
func (r *fakePaymentRepo) ListStalePending(time.Time) ([]models.Payment, error) {
	return nil, nil
}
func (r *fakePaymentRepo) MarkCompleted(orderID, method string, at time.Time, isSandbox bool) (models.Payment, bool, error) {
	r.completedN++
	r.payment.Status = "completed"
	r.payment.CompletedAt = &at
	r.payment.IsSandbox = isSandbox
	return r.payment, r.alreadyDone, nil
}

// orderService is only used for RecalcListmakTotals here; a no-op stub suffices.
type fakeOrderService struct{ OrderService }

func (fakeOrderService) RecalcListmakTotals(uint) {}

func newSvc(repo *fakePaymentRepo, pak *fakePakasir) *paymentService {
	return &paymentService{
		paymentRepo:  repo,
		orderService: fakeOrderService{},
		pakasir:      pak,
	}
}

// fakeShareService returns a valid, non-expired share for any id.
type fakeShareService struct{ ShareService }

func (fakeShareService) GetShareLink(string) (models.ShareLink, error) {
	return models.ShareLink{ListmakID: 1}, nil
}

type fakeAppConfig struct{ testing bool }

func (f fakeAppConfig) TestingMode() bool       { return f.testing }
func (fakeAppConfig) SetTestingMode(bool) error { return nil }

func newCheckoutSvc(repo *fakePaymentRepo) *paymentService {
	return &paymentService{
		paymentRepo:  repo,
		orderService: fakeOrderService{},
		shareService: fakeShareService{},
		pakasir:      &fakePakasir{project: "proj"},
		appConfig:    fakeAppConfig{},
	}
}

func TestCheckoutValidation(t *testing.T) {
	base := CheckoutInput{
		ShareID: "s1", GuestName: "Budi", GuestWhatsapp: "081234567890",
		Items: []CheckoutItem{{OrderDetail: "Nasi", Price: 15000, Qty: 1}},
	}
	tooMany := make([]CheckoutItem, maxCheckoutItems+1)
	for i := range tooMany {
		tooMany[i] = CheckoutItem{OrderDetail: "x", Price: 1000, Qty: 1}
	}

	tests := []struct {
		name    string
		mutate  func(c *CheckoutInput)
		wantErr error
	}{
		{"missing whatsapp", func(c *CheckoutInput) { c.GuestWhatsapp = "" }, ErrInvalidGuest},
		{"bad whatsapp", func(c *CheckoutInput) { c.GuestWhatsapp = "abc" }, ErrInvalidGuest},
		{"empty name", func(c *CheckoutInput) { c.GuestName = "  " }, ErrInvalidGuest},
		{"no items", func(c *CheckoutInput) { c.Items = nil }, ErrEmptyOrders},
		{"too many items", func(c *CheckoutInput) { c.Items = tooMany }, ErrTooManyItems},
		{"negative price", func(c *CheckoutInput) { c.Items[0].Price = -1 }, ErrInvalidAmount},
		{"price over cap", func(c *CheckoutInput) { c.Items[0].Price = maxLinePrice + 1 }, ErrInvalidAmount},
		{"qty over cap", func(c *CheckoutInput) { c.Items[0].Qty = maxQty + 1 }, ErrInvalidAmount},
		{"zero total", func(c *CheckoutInput) { c.Items[0].Price = 0 }, ErrInvalidAmount},
		{"total over cap", func(c *CheckoutInput) { c.Items[0].Price = maxLinePrice; c.Items[0].Qty = 20 }, ErrInvalidAmount},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := base
			in.Items = append([]CheckoutItem(nil), base.Items...) // deep-copy items
			tc.mutate(&in)
			_, err := newCheckoutSvc(&fakePaymentRepo{}).Checkout(in)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestCheckoutHappyPath(t *testing.T) {
	in := CheckoutInput{
		ShareID: "s1", GuestName: "Budi", GuestWhatsapp: "081234567890",
		Items: []CheckoutItem{
			{Name: "Budi", OrderDetail: "Nasi Ayam", Price: 15000, Qty: 2},
			{Name: "Ani", OrderDetail: "Es Teh", Price: 5000, Qty: 1},
		},
	}
	p, err := newCheckoutSvc(&fakePaymentRepo{}).Checkout(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Amount != 35000 { // 15000*2 + 5000
		t.Fatalf("amount = %d, want 35000", p.Amount)
	}
	// Orders must NOT be materialized yet — only the snapshot is stored.
	var items []models.PaymentItem
	if err := json.Unmarshal(p.ItemsSnapshot, &items); err != nil {
		t.Fatalf("snapshot not valid json: %v", err)
	}
	if len(items) != 2 || items[1].Name != "Ani" {
		t.Fatalf("snapshot items wrong: %+v", items)
	}
}

func TestHandleWebhook(t *testing.T) {
	base := models.Payment{OrderID: "LMK1-abc", ListmakID: 1, Amount: 22000, Status: "pending"}

	tests := []struct {
		name     string
		payment  models.Payment
		payload  WebhookPayload
		detail   *PakasirTransaction
		wantCode int
		wantMark bool
	}{
		{
			name:     "wrong project rejected",
			payment:  base,
			payload:  WebhookPayload{Project: "evil", OrderID: "LMK1-abc", Amount: 22000},
			wantCode: 400,
		},
		{
			name:     "unknown order",
			payment:  models.Payment{OrderID: "other"},
			payload:  WebhookPayload{Project: "proj", OrderID: "LMK1-abc", Amount: 22000},
			wantCode: 404,
		},
		{
			name:     "amount tampering rejected before touching gateway",
			payment:  base,
			payload:  WebhookPayload{Project: "proj", OrderID: "LMK1-abc", Amount: 99999},
			wantCode: 400,
		},
		{
			name:     "gateway says not completed -> accepted, not marked",
			payment:  base,
			payload:  WebhookPayload{Project: "proj", OrderID: "LMK1-abc", Amount: 22000},
			detail:   &PakasirTransaction{OrderID: "LMK1-abc", Status: "pending", Amount: 22000},
			wantCode: 202,
		},
		{
			name:     "happy path marks completed",
			payment:  base,
			payload:  WebhookPayload{Project: "proj", OrderID: "LMK1-abc", Amount: 22000},
			detail:   &PakasirTransaction{OrderID: "LMK1-abc", Status: "completed", Amount: 22000},
			wantCode: 200,
			wantMark: true,
		},
		{
			name:     "already completed is idempotent",
			payment:  models.Payment{OrderID: "LMK1-abc", Amount: 22000, Status: "completed"},
			payload:  WebhookPayload{Project: "proj", OrderID: "LMK1-abc", Amount: 22000},
			wantCode: 200,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakePaymentRepo{payment: tc.payment}
			svc := newSvc(repo, &fakePakasir{project: "proj", detail: tc.detail})
			code, _ := svc.HandleWebhook(tc.payload)
			if code != tc.wantCode {
				t.Fatalf("code = %d, want %d", code, tc.wantCode)
			}
			if (repo.completedN > 0) != tc.wantMark {
				t.Fatalf("marked completed = %v, want %v", repo.completedN > 0, tc.wantMark)
			}
		})
	}
}

func TestNormalizeWhatsapp(t *testing.T) {
	cases := map[string]string{
		"0812-3456-7890":  "6281234567890",
		"+62 812 3456789": "62812" + "3456789",
		"6281234567890":   "6281234567890",
		"(0812) 34567890": "62812" + "34567890",
	}
	for in, want := range cases {
		if got := normalizeWhatsapp(in); got != want {
			t.Errorf("normalizeWhatsapp(%q) = %q, want %q", in, got, want)
		}
	}
}

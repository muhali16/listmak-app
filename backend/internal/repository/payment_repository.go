package repository

import (
	"encoding/json"
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRepository interface {
	// Create persists a pending payment (with its item snapshot). No orders are
	// written yet — they are materialized only when the payment completes, so an
	// unpaid/cancelled checkout never pollutes the order list.
	Create(payment models.Payment) (models.Payment, error)
	GetByOrderID(orderID string) (models.Payment, error)
	ExistsByOrderID(orderID string) (bool, error)
	// MarkCompleted flips the payment to completed and materializes its snapshot
	// items into paid Orders, under a row lock. Returns alreadyDone=true if it was
	// already completed, so callers stay idempotent against duplicate webhooks.
	MarkCompleted(orderID, method string, completedAt time.Time, isSandbox bool) (payment models.Payment, alreadyDone bool, err error)
	// MarkCancelled cancels a still-pending payment. Idempotent: no-op if the
	// payment is not pending. No orders exist yet, so none are touched.
	MarkCancelled(orderID string) error
	// ListStalePending returns pending payments created before cutoff (for the
	// background reconciler).
	ListStalePending(cutoff time.Time) ([]models.Payment, error)

	// Admin
	ListForAdmin(page, limit int, status, search string) ([]models.Payment, int64, error)
	Stats() (PaymentStats, error)
}

// PaymentStats is the revenue/ops summary for the admin dashboard.
type PaymentStats struct {
	TotalCollected int64 `json:"total_collected"` // sum(amount) of completed payments (base, excl. fee)
	TotalFee       int64 `json:"total_fee"`       // sum(fee) of completed payments
	CompletedCount int64 `json:"completed_count"`
	PendingCount   int64 `json:"pending_count"`
	CancelledCount int64 `json:"cancelled_count"`
	TotalCount     int64 `json:"total_count"`
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment models.Payment) (models.Payment, error) {
	if err := r.db.Create(&payment).Error; err != nil {
		return models.Payment{}, err
	}
	return payment, nil
}

func (r *paymentRepository) ListForAdmin(page, limit int, status, search string) ([]models.Payment, int64, error) {
	var payments []models.Payment
	var total int64
	offset := (page - 1) * limit

	q := r.db.Model(&models.Payment{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("order_id ILIKE ? OR guest_name ILIKE ? OR guest_whatsapp ILIKE ?", like, like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&payments).Error; err != nil {
		return nil, 0, err
	}
	return payments, total, nil
}

func (r *paymentRepository) Stats() (PaymentStats, error) {
	var s PaymentStats
	// Revenue from completed, non-sandbox payments only.
	row := r.db.Model(&models.Payment{}).
		Where("status = ? AND is_sandbox = ?", "completed", false).
		Select("COALESCE(SUM(amount),0), COALESCE(SUM(fee),0)").Row()
	if err := row.Scan(&s.TotalCollected, &s.TotalFee); err != nil {
		return s, err
	}
	// Counts per status.
	type sc struct {
		Status string
		Count  int64
	}
	var rows []sc
	if err := r.db.Model(&models.Payment{}).
		Select("status, COUNT(*) AS count").Group("status").Scan(&rows).Error; err != nil {
		return s, err
	}
	for _, x := range rows {
		s.TotalCount += x.Count
		switch x.Status {
		case "completed":
			s.CompletedCount = x.Count
		case "pending":
			s.PendingCount = x.Count
		case "cancelled":
			s.CancelledCount = x.Count
		}
	}
	return s, nil
}

func (r *paymentRepository) GetByOrderID(orderID string) (models.Payment, error) {
	var p models.Payment
	if err := r.db.Where("order_id = ?", orderID).First(&p).Error; err != nil {
		return models.Payment{}, err
	}
	return p, nil
}

func (r *paymentRepository) MarkCancelled(orderID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var p models.Payment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_id = ?", orderID).First(&p).Error; err != nil {
			return err
		}
		if p.Status != "pending" {
			return nil // idempotent: already cancelled/completed
		}
		// No orders exist yet (snapshot only), so just cancel the payment.
		return tx.Model(&models.Payment{}).Where("id = ?", p.ID).
			Update("status", "cancelled").Error
	})
}

func (r *paymentRepository) ListStalePending(cutoff time.Time) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Where("status = ? AND created_at < ?", "pending", cutoff).
		Order("created_at asc").Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) ExistsByOrderID(orderID string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Payment{}).Where("order_id = ?", orderID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *paymentRepository) MarkCompleted(orderID, method string, completedAt time.Time, isSandbox bool) (models.Payment, bool, error) {
	var p models.Payment
	var alreadyDone bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Lock the payment row so concurrent webhook deliveries serialize here.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_id = ?", orderID).First(&p).Error; err != nil {
			return err
		}

		if p.Status == "completed" {
			alreadyDone = true
			return nil
		}

		p.Status = "completed"
		if method != "" {
			p.PaymentMethod = method
		}
		p.CompletedAt = &completedAt
		if isSandbox {
			p.IsSandbox = true // authoritative from Pakasir transactiondetail
		}

		// Guard: only materialize orders while the listmak is still active. If it
		// was already closed (completed/cancelled/deleted), the money still counts
		// as received, but we create NO orders and leave Fulfilled=false so admin
		// can refund/review — prevents phantom late orders after the window closed.
		var listmak models.Listmak
		listmakOpen := tx.Where("id = ?", p.ListmakID).First(&listmak).Error == nil &&
			listmak.Status == "active"

		p.Fulfilled = listmakOpen
		if err := tx.Save(&p).Error; err != nil {
			return err
		}
		if !listmakOpen {
			return nil
		}

		// Materialize the snapshot into paid orders now that money is confirmed.
		var items []models.PaymentItem
		if len(p.ItemsSnapshot) > 0 {
			if err := json.Unmarshal(p.ItemsSnapshot, &items); err != nil {
				return err
			}
		}
		if len(items) == 0 {
			return nil
		}
		now := time.Now()
		orders := make([]models.Order, 0, len(items))
		for _, it := range items {
			orders = append(orders, models.Order{
				ListmakID:   p.ListmakID,
				PaymentID:   &p.ID,
				Name:        it.Name,
				OrderDetail: it.OrderDetail,
				VendorName:  it.VendorName,
				Price:       it.Price,
				Qty:         it.Qty,
				IsPaid:      true,
				PaidAt:      &now,
				PaidAmount:  it.Price * float64(it.Qty), // freeze what was actually paid
				AddedVia:    "sharelink",
			})
		}
		return tx.Create(&orders).Error
	})
	if err != nil {
		return models.Payment{}, false, err
	}
	return p, alreadyDone, nil
}

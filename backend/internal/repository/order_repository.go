package repository

import (
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

type OrderRepository interface {
	GetOrdersByListmakId(listmakId uint, isPaid *bool, search string) ([]models.Order, error)
	GetOrderById(id uint) (models.Order, error)
	CreateOrder(order models.Order) (models.Order, error)
	CreateOrders(orders []models.Order) ([]models.Order, error)
	UpdateOrder(order models.Order) (models.Order, error)
	UpdateOrdersPaidByName(listmakId uint, name string, isPaid bool) (int64, error)
	DeleteOrder(id uint) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) GetOrdersByListmakId(listmakId uint, isPaid *bool, search string) ([]models.Order, error) {
	var orders []models.Order
	query := r.db.Where("listmak_id = ?", listmakId)

	if isPaid != nil {
		query = query.Where("is_paid = ?", *isPaid)
	}
	if search != "" {
		likePattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR order_detail LIKE ?", likePattern, likePattern)
	}

	if err := query.Order("id asc").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) GetOrderById(id uint) (models.Order, error) {
	var order models.Order
	if err := r.db.First(&order, id).Error; err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (r *orderRepository) CreateOrder(order models.Order) (models.Order, error) {
	if err := r.db.Create(&order).Error; err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (r *orderRepository) CreateOrders(orders []models.Order) ([]models.Order, error) {
	if err := r.db.Create(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) UpdateOrder(order models.Order) (models.Order, error) {
	if err := r.db.Save(&order).Error; err != nil {
		return models.Order{}, err
	}
	return order, nil
}

// UpdateOrdersPaidByName flips is_paid for every order in a listmak whose name
// matches the given name (exact, case-insensitive, whitespace-trimmed). The
// update runs inside a single transaction so it is all-or-nothing: either every
// matching row is updated or none is. Returns the number of rows affected.
func (r *orderRepository) UpdateOrdersPaidByName(listmakId uint, name string, isPaid bool) (int64, error) {
	updates := map[string]interface{}{
		"is_paid": isPaid,
	}
	// Replicate UpdateOrderPaidStatus: set paid_at when marking paid, clear it otherwise.
	if isPaid {
		updates["paid_at"] = time.Now()
	} else {
		updates["paid_at"] = nil
	}

	var affected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Order{}).
			Where("listmak_id = ? AND LOWER(TRIM(name)) = LOWER(TRIM(?))", listmakId, name).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		affected = result.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (r *orderRepository) DeleteOrder(id uint) error {
	return r.db.Delete(&models.Order{}, id).Error
}

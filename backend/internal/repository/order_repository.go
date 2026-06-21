package repository

import (
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

type OrderPriceUpdate struct {
	VendorName  string
	OrderDetail string
	Price       int
}

type OrderRepository interface {
	GetOrdersByListmakId(listmakId uint, isPaid *bool, search string) ([]models.Order, error)
	GetOrderById(id uint) (models.Order, error)
	CreateOrder(order models.Order) (models.Order, error)
	CreateOrders(orders []models.Order) ([]models.Order, error)
	UpdateOrder(order models.Order) (models.Order, error)
	UpdateOrdersPaidByName(listmakId uint, name string, isPaid bool) (int64, error)
	BulkUpdatePriceByVendorAndDetail(listmakID uint, updates []OrderPriceUpdate) error
	DeleteOrder(id uint) error
	UpdateVendorName(id uint, vendor string) error
	GetFoodSuggestions(listmakID uint, query string) ([]string, error)
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

func (r *orderRepository) BulkUpdatePriceByVendorAndDetail(listmakID uint, updates []OrderPriceUpdate) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, u := range updates {
			if err := tx.Model(&models.Order{}).
				Where("listmak_id = ? AND LOWER(TRIM(vendor_name)) = LOWER(TRIM(?)) AND LOWER(TRIM(order_detail)) = LOWER(TRIM(?))",
					listmakID, u.VendorName, u.OrderDetail).
				Update("price", u.Price).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *orderRepository) DeleteOrder(id uint) error {
	return r.db.Delete(&models.Order{}, id).Error
}

func (r *orderRepository) UpdateVendorName(id uint, vendor string) error {
	return r.db.Model(&models.Order{}).Where("id = ?", id).Update("vendor_name", vendor).Error
}

func (r *orderRepository) GetFoodSuggestions(listmakID uint, query string) ([]string, error) {
	var results []string

	if query == "" {
		err := r.db.Model(&models.Order{}).
			Select("order_detail").
			Where("listmak_id = ? AND deleted_at IS NULL", listmakID).
			Group("order_detail").
			Order("COUNT(*) DESC").
			Limit(8).
			Pluck("order_detail", &results).Error
		return results, err
	}

	likePattern := "%" + query + "%"

	var sameListmak []string
	if err := r.db.Model(&models.Order{}).
		Where("listmak_id = ? AND order_detail ILIKE ? AND deleted_at IS NULL", listmakID, likePattern).
		Distinct("order_detail").
		Order("order_detail").
		Limit(5).
		Pluck("order_detail", &sameListmak).Error; err != nil {
		return nil, err
	}
	results = append(results, sameListmak...)

	remaining := 8 - len(results)
	if remaining > 0 {
		q := r.db.Model(&models.Order{}).
			Where("order_detail ILIKE ? AND deleted_at IS NULL", likePattern).
			Distinct("order_detail").
			Order("order_detail").
			Limit(remaining)
		if len(sameListmak) > 0 {
			q = q.Where("order_detail NOT IN ?", sameListmak)
		}
		var global []string
		if err := q.Pluck("order_detail", &global).Error; err != nil {
			return nil, err
		}
		results = append(results, global...)
	}

	return results, nil
}

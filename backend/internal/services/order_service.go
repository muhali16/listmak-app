package services

import (
	"errors"
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
)

// ErrNoOrdersMatched is returned when a bulk paid-status update finds no order
// matching the given name in the listmak. Callers should map this to HTTP 404.
var ErrNoOrdersMatched = errors.New("no orders matched the given name")

type OrderService interface {
	GetOrdersByListmakId(listmakId uint, isPaid *bool, search string) ([]models.Order, error)
	CreateOrder(order models.Order) (models.Order, error)
	CreateOrdersBulk(listmakId uint, orders []models.Order) (int, []models.Order, error)
	UpdateOrder(order models.Order) (models.Order, error)
	UpdateOrderPaidStatus(id uint, isPaid bool) (models.Order, error)
	UpdateOrdersPaidByName(listmakId uint, name string, isPaid bool) (int64, error)
	DeleteOrder(id uint) error
}

type orderService struct {
	orderRepo   repository.OrderRepository
	listmakRepo repository.ListmakRepository // Injected to update totals manually if no trigger
}

func NewOrderService(orderRepo repository.OrderRepository, listmakRepo repository.ListmakRepository) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		listmakRepo: listmakRepo,
	}
}

func (s *orderService) GetOrdersByListmakId(listmakId uint, isPaid *bool, search string) ([]models.Order, error) {
	return s.orderRepo.GetOrdersByListmakId(listmakId, isPaid, search)
}

func (s *orderService) CreateOrder(order models.Order) (models.Order, error) {
	// Calculate TotalPrice manually just in case
	order.TotalPrice = order.Price * float64(order.Qty)

	newOrder, err := s.orderRepo.CreateOrder(order)
	if err != nil {
		return models.Order{}, err
	}
	s.updateListmakTotals(order.ListmakID)
	return newOrder, nil
}

func (s *orderService) CreateOrdersBulk(listmakId uint, orders []models.Order) (int, []models.Order, error) {
	// Prep data
	for i := range orders {
		orders[i].ListmakID = listmakId
		orders[i].TotalPrice = orders[i].Price * float64(orders[i].Qty)
		if orders[i].AddedVia == "" {
			orders[i].AddedVia = "parse"
		}
	}

	createdOrders, err := s.orderRepo.CreateOrders(orders)
	if err != nil {
		return 0, nil, err
	}

	s.updateListmakTotals(listmakId)
	return len(createdOrders), createdOrders, nil
}

func (s *orderService) UpdateOrder(order models.Order) (models.Order, error) {
	existing, err := s.orderRepo.GetOrderById(order.ID)
	if err != nil {
		return models.Order{}, err
	}

	existing.Name = order.Name
	existing.OrderDetail = order.OrderDetail
	existing.Price = order.Price
	existing.Qty = order.Qty

	updatedOrder, err := s.orderRepo.UpdateOrder(existing)
	if err != nil {
		return models.Order{}, err
	}

	s.updateListmakTotals(existing.ListmakID)
	return updatedOrder, nil
}

func (s *orderService) UpdateOrderPaidStatus(id uint, isPaid bool) (models.Order, error) {
	order, err := s.orderRepo.GetOrderById(id)
	if err != nil {
		return models.Order{}, err
	}

	order.IsPaid = isPaid
	if isPaid {
		now := time.Now()
		order.PaidAt = &now
	} else {
		order.PaidAt = nil
	}

	order, err = s.orderRepo.UpdateOrder(order)
	if err != nil {
		return models.Order{}, err
	}

	s.updateListmakTotals(order.ListmakID)
	return order, nil
}

// UpdateOrdersPaidByName toggles is_paid for all orders belonging to a name in
// a listmak (used for the per-person "Lunas" toggle in the UI). The row updates
// happen all-or-nothing in the repository's transaction; the listmak totals are
// recalculated once after the bulk update succeeds. Returns ErrNoOrdersMatched
// if no order matches the name.
func (s *orderService) UpdateOrdersPaidByName(listmakId uint, name string, isPaid bool) (int64, error) {
	count, err := s.orderRepo.UpdateOrdersPaidByName(listmakId, name, isPaid)
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, ErrNoOrdersMatched
	}

	s.updateListmakTotals(listmakId)
	return count, nil
}

func (s *orderService) DeleteOrder(id uint) error {
	order, err := s.orderRepo.GetOrderById(id)
	if err != nil {
		return err // Or ignore if not found
	}
	listmakId := order.ListmakID

	if err := s.orderRepo.DeleteOrder(id); err != nil {
		return err
	}

	s.updateListmakTotals(listmakId)
	return nil
}

// Helper to recalculate totals
func (s *orderService) updateListmakTotals(listmakId uint) {
	// Get all orders
	orders, err := s.orderRepo.GetOrdersByListmakId(listmakId, nil, "")
	if err != nil {
		return
	}

	var totalOrders int
	var totalAmount float64
	var paidAmount float64

	for _, o := range orders {
		totalOrders++
		totalAmount += o.TotalPrice
		if o.IsPaid {
			paidAmount += o.TotalPrice
		}
	}

	listmak, err := s.listmakRepo.GetListmakById(listmakId)
	if err == nil {
		listmak.TotalOrders = totalOrders
		listmak.TotalAmount = totalAmount
		listmak.PaidAmount = paidAmount
		// Avoid recursive loop if listmak update triggers something? No, it's fine.
		// Also ListmakRepo.UpdateListmak saves all fields.
		s.listmakRepo.UpdateListmak(listmak)
	}
}

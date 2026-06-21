package services

import (
	"errors"
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
	"gorm.io/gorm"
)

type SummaryService interface {
	GetOrGenerateSummary(requestID string, listmakID uint, location string) (*models.ListmakSummary, error)
	ConfirmPrices(listmakID uint, items []ConfirmItem) (*models.ListmakSummary, error)
	EstimatePrice(requestID string, itemDetail string, location string) (int, bool, error)
}

type ConfirmItem struct {
	VendorName      string `json:"vendor_name"`
	ItemName        string `json:"item_name"`
	UnitPriceActual int    `json:"unit_price_actual"`
}

type summaryService struct {
	summaryRepo repository.SummaryRepository
	catalogRepo repository.PriceCatalogRepository
	orderRepo   repository.OrderRepository
	aiService   AIService
}

func NewSummaryService(
	summaryRepo repository.SummaryRepository,
	catalogRepo repository.PriceCatalogRepository,
	orderRepo repository.OrderRepository,
	aiService AIService,
) SummaryService {
	return &summaryService{
		summaryRepo: summaryRepo,
		catalogRepo: catalogRepo,
		orderRepo:   orderRepo,
		aiService:   aiService,
	}
}

func (s *summaryService) GetOrGenerateSummary(requestID string, listmakID uint, location string) (*models.ListmakSummary, error) {
	orders, err := s.orderRepo.GetOrdersByListmakId(listmakID, nil, "")
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, errors.New("tidak ada order di listmak ini")
	}

	var maxOrderID uint
	for _, o := range orders {
		if o.ID > maxOrderID {
			maxOrderID = o.ID
		}
	}

	existing, err := s.summaryRepo.GetByListmakID(listmakID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if existing != nil && existing.OrderWatermark >= maxOrderID {
		return existing, nil
	}

	var ordersToProcess []models.Order
	var existingSummary *models.SummaryData
	if existing != nil {
		existingSummary = existing.Summary
		for _, o := range orders {
			if o.ID > existing.OrderWatermark {
				ordersToProcess = append(ordersToProcess, o)
			}
		}
	} else {
		ordersToProcess = orders
	}

	catalog, _ := s.catalogRepo.GetAll()

	summaryData, err := s.aiService.SummarizeOrders(requestID, ordersToProcess, catalog, existingSummary, location)
	if err != nil {
		return nil, err
	}

	result := &models.ListmakSummary{
		ListmakID:      listmakID,
		Summary:        summaryData,
		OrderWatermark: maxOrderID,
		GeneratedAt:    time.Now(),
	}
	if existing != nil {
		result.ID = existing.ID
		result.ConfirmedAt = existing.ConfirmedAt
		result.TotalActual = existing.TotalActual
	}

	if err := s.summaryRepo.Upsert(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *summaryService) ConfirmPrices(listmakID uint, items []ConfirmItem) (*models.ListmakSummary, error) {
	existing, err := s.summaryRepo.GetByListmakID(listmakID)
	if err != nil {
		return nil, err
	}

	for i, vendor := range existing.Summary.Vendors {
		for j, item := range vendor.Items {
			for _, confirm := range items {
				if confirm.VendorName == vendor.Name && confirm.ItemName == item.Name {
					price := confirm.UnitPriceActual
					existing.Summary.Vendors[i].Items[j].UnitPriceActual = &price
				}
			}
		}
	}

	totalActual := 0
	allConfirmed := true
	for _, vendor := range existing.Summary.Vendors {
		for _, item := range vendor.Items {
			if item.UnitPriceActual != nil {
				totalActual += *item.UnitPriceActual * item.Qty
			} else {
				allConfirmed = false
			}
		}
	}

	existing.TotalActual = &totalActual
	if allConfirmed {
		now := time.Now()
		existing.ConfirmedAt = &now
	}

	if err := s.summaryRepo.Upsert(existing); err != nil {
		return nil, err
	}

	entries := make([]models.PriceCatalog, 0, len(items))
	priceUpdates := make([]repository.OrderPriceUpdate, 0, len(items))
	for _, item := range items {
		entries = append(entries, models.PriceCatalog{
			VendorName: item.VendorName,
			ItemName:   item.ItemName,
			Price:      item.UnitPriceActual,
		})
		if item.UnitPriceActual > 0 {
			priceUpdates = append(priceUpdates, repository.OrderPriceUpdate{
				VendorName:  item.VendorName,
				OrderDetail: item.ItemName,
				Price:       item.UnitPriceActual,
			})
		}
	}
	s.catalogRepo.UpsertBatch(entries)
	if len(priceUpdates) > 0 {
		_ = s.orderRepo.BulkUpdatePriceByVendorAndDetail(listmakID, priceUpdates)
	}

	return existing, nil
}

func (s *summaryService) EstimatePrice(requestID string, itemDetail string, location string) (int, bool, error) {
	entry, err := s.catalogRepo.FuzzyMatch(itemDetail)
	if err == nil {
		return entry.Price, false, nil
	}

	price, err := s.aiService.EstimatePrice(requestID, itemDetail, location)
	if err != nil {
		return 0, true, err
	}
	return price, true, nil
}

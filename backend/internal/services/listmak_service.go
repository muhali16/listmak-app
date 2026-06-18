package services

import (
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
)

type ListmakService interface {
	GetAllListmaks(page, limit int, status string, startDate, endDate *time.Time) ([]models.Listmak, int64, error)
	GetListmakById(id uint) (models.Listmak, error)
	GetListmakByDate(date time.Time) ([]models.Listmak, error)
	CreateListmak(listmak models.Listmak) (models.Listmak, error)
	UpdateListmak(listmak models.Listmak) (models.Listmak, error)
	DeleteListmak(id uint) error
}

type listmakService struct {
	listmakRepo repository.ListmakRepository
}

func NewListmakService(listmakRepo repository.ListmakRepository) ListmakService {
	return &listmakService{
		listmakRepo: listmakRepo,
	}
}

func (s *listmakService) GetAllListmaks(page, limit int, status string, startDate, endDate *time.Time) ([]models.Listmak, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return s.listmakRepo.GetAllListmaks(page, limit, status, startDate, endDate)
}

func (s *listmakService) GetListmakById(id uint) (models.Listmak, error) {
	return s.listmakRepo.GetListmakById(id)
}

func (s *listmakService) GetListmakByDate(date time.Time) ([]models.Listmak, error) {
	return s.listmakRepo.GetListmakByDate(date)
}

func (s *listmakService) CreateListmak(listmak models.Listmak) (models.Listmak, error) {
	// Business logic: check if listmak for date exists?
	// Unique constraint on date handled by DB or repo
	return s.listmakRepo.CreateListmak(listmak)
}

func (s *listmakService) UpdateListmak(listmak models.Listmak) (models.Listmak, error) {
	return s.listmakRepo.UpdateListmak(listmak)
}

func (s *listmakService) DeleteListmak(id uint) error {
	return s.listmakRepo.DeleteListmak(id)
}

package services

import (
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
)

type ListmakService interface {
	GetAllListmaks(page, limit int, status string, startDate, endDate *time.Time, userId uint) ([]models.Listmak, int64, error)
	GetListmakById(id uint) (models.Listmak, error)
	GetListmakByDate(date time.Time, userId uint) ([]models.Listmak, error)
	CreateListmak(listmak models.Listmak) (models.Listmak, error)
	UpdateListmak(listmak models.Listmak) (models.Listmak, error)
	DeleteListmak(id uint) error
}

type listmakService struct {
	listmakRepo repository.ListmakRepository
	appConfig   AppConfig
}

func NewListmakService(listmakRepo repository.ListmakRepository, appConfig AppConfig) ListmakService {
	return &listmakService{
		listmakRepo: listmakRepo,
		appConfig:   appConfig,
	}
}

func (s *listmakService) GetAllListmaks(page, limit int, status string, startDate, endDate *time.Time, userId uint) ([]models.Listmak, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	// Mode-scoped for regular users: only listmaks matching the current mode.
	mode := s.appConfig.TestingMode()
	return s.listmakRepo.GetAllListmaks(page, limit, status, startDate, endDate, userId, &mode)
}

func (s *listmakService) GetListmakById(id uint) (models.Listmak, error) {
	return s.listmakRepo.GetListmakById(id)
}

func (s *listmakService) GetListmakByDate(date time.Time, userId uint) ([]models.Listmak, error) {
	mode := s.appConfig.TestingMode()
	return s.listmakRepo.GetListmakByDate(date, userId, &mode)
}

func (s *listmakService) CreateListmak(listmak models.Listmak) (models.Listmak, error) {
	// Tag data created while in testing mode so it stays out of production views.
	listmak.IsSandbox = s.appConfig.TestingMode()
	return s.listmakRepo.CreateListmak(listmak)
}

func (s *listmakService) UpdateListmak(listmak models.Listmak) (models.Listmak, error) {
	return s.listmakRepo.UpdateListmak(listmak)
}

func (s *listmakService) DeleteListmak(id uint) error {
	return s.listmakRepo.DeleteListmak(id)
}

package services

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
	"github.com/muhali16/listmak-service/pkg/utils"
)

type ShareService interface {
	CreateShareLink(listmakId uint, title string, expiresAt time.Time, userId uint) (models.ShareLink, error)
	GetShareLink(shareId string) (models.ShareLink, error)
	DeleteShareLink(id uint) error

	CreateViewShare(listmakId uint, title string, userId uint) (models.ViewShare, error)
	GetViewShare(viewId string) (models.ViewShare, error)
}

type shareService struct {
	shareRepo     repository.ShareLinkRepository
	viewShareRepo repository.ViewShareRepository
	listmakRepo   repository.ListmakRepository
}

func NewShareService(
	shareRepo repository.ShareLinkRepository,
	viewShareRepo repository.ViewShareRepository,
	listmakRepo repository.ListmakRepository,
) ShareService {
	return &shareService{
		shareRepo:     shareRepo,
		viewShareRepo: viewShareRepo,
		listmakRepo:   listmakRepo,
	}
}

func (s *shareService) CreateShareLink(listmakId uint, title string, expiresAt time.Time, userId uint) (models.ShareLink, error) {
	// Verify listmak exists
	_, err := s.listmakRepo.GetListmakById(listmakId)
	if err != nil {
		return models.ShareLink{}, errors.New("Listmak not found")
	}

	shareId, _ := utils.GenerateRandomString(8)
	// Make sure unique... but for simplicity assume random is enough or handle db error

	shareLink := models.ShareLink{
		ShareID:   shareId,
		ListmakID: listmakId,
		Title:     title,
		ExpiresAt: expiresAt,
		CreatedBy: &userId,
	}

	return s.shareRepo.CreateShareLink(shareLink)
}

func (s *shareService) GetShareLink(shareId string) (models.ShareLink, error) {
	shareLink, err := s.shareRepo.GetShareLinkByShareId(shareId)
	if err != nil {
		return models.ShareLink{}, err
	}

	// Check expiry
	if time.Now().After(shareLink.ExpiresAt) {
		return shareLink, errors.New("EXPIRED")
	}

	return shareLink, nil
}

func (s *shareService) DeleteShareLink(id uint) error {
	return s.shareRepo.DeleteShareLink(id)
}

func (s *shareService) CreateViewShare(listmakId uint, title string, userId uint) (models.ViewShare, error) {
	// Verify and get listmak data for snapshot
	listmak, err := s.listmakRepo.GetListmakById(listmakId)
	if err != nil {
		return models.ViewShare{}, errors.New("Listmak not found")
	}

	snapshot, _ := json.Marshal(listmak)

	viewId, _ := utils.GenerateRandomString(8)

	viewShare := models.ViewShare{
		ViewID:       viewId,
		ListmakID:    listmakId,
		Title:        title,
		SnapshotData: snapshot,
		CreatedBy:    &userId,
	}

	return s.viewShareRepo.CreateViewShare(viewShare)
}

func (s *shareService) GetViewShare(viewId string) (models.ViewShare, error) {
	// Retrieve view share
	// Note: Request says "Get data listmak for view (public)".
	// If we use snapshot, we return snapshot. If we return real-time, we preload.
	// API doc says: "Ambil data listmak untuk view".
	// DB schema says "SnapshotData JSONB".
	// The implementation choice depends on requirements. Usually "view share" implies read-only view of current state OR snapshot.
	// Given "SnapshotData" exists in schema, we probably return that or merge.
	// But `GetViewShare` just returns model. Controller will parse.
	return s.viewShareRepo.GetViewShareByViewId(viewId)
}

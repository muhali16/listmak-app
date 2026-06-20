package services

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
	"github.com/muhali16/listmak-service/pkg/utils"
)

// ErrListmakUnavailable is returned when a live view-share points at a listmak
// that no longer exists (e.g. soft-deleted). Callers should map this to HTTP 404
// with a clear message instead of leaking the raw GORM error as a 500.
var ErrListmakUnavailable = errors.New("listmak unavailable")

type ShareService interface {
	CreateShareLink(listmakId uint, title string, expiresAt time.Time, userId uint) (models.ShareLink, error)
	GetShareLink(shareId string) (models.ShareLink, error)
	DeleteShareLink(id uint) error

	CreateViewShare(listmakId uint, title string, userId uint) (models.ViewShare, error)
	GetViewShare(viewId string) (models.ViewShare, error)

	GetActiveSharesForListmak(listmakId uint) (shareLink *models.ShareLink, viewShare *models.ViewShare, err error)
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

func (s *shareService) GetActiveSharesForListmak(listmakId uint) (shareLink *models.ShareLink, viewShare *models.ViewShare, err error) {
	sl, err := s.shareRepo.GetActiveShareLinkByListmakId(listmakId)
	if err != nil {
		return nil, nil, err
	}

	vs, err := s.viewShareRepo.GetViewShareByListmakId(listmakId)
	if err != nil {
		return nil, nil, err
	}

	return sl, vs, nil
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
		// New links serve live data. The snapshot above is still stored as a
		// frozen fallback, but GetViewShare overwrites it with fresh data on read.
		IsLive:    true,
		CreatedBy: &userId,
	}

	return s.viewShareRepo.CreateViewShare(viewShare)
}

func (s *shareService) GetViewShare(viewId string) (models.ViewShare, error) {
	viewShare, err := s.viewShareRepo.GetViewShareByViewId(viewId)
	if err != nil {
		return models.ViewShare{}, err
	}

	// Legacy links (is_live == false, the AutoMigrate default for every
	// pre-existing row) keep serving their frozen snapshot exactly as before.
	if !viewShare.IsLive {
		return viewShare, nil
	}

	// Live links re-fetch the current listmak state and overwrite SnapshotData
	// in-memory, so the controller and frontend see an identical response shape
	// to the snapshot path (no changes needed downstream).
	listmak, err := s.listmakRepo.GetListmakById(viewShare.ListmakID)
	if err != nil {
		// Listmak gone (e.g. soft-deleted): surface a clear 404, not a 500.
		return models.ViewShare{}, ErrListmakUnavailable
	}

	liveSnapshot, err := json.Marshal(listmak)
	if err != nil {
		return models.ViewShare{}, err
	}
	viewShare.SnapshotData = liveSnapshot

	return viewShare, nil
}

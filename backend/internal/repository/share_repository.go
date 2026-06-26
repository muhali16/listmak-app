package repository

import (
	"errors"

	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

type ShareLinkRepository interface {
	CreateShareLink(shareLink models.ShareLink) (models.ShareLink, error)
	GetShareLinkByShareId(shareId string) (models.ShareLink, error)
	DeleteShareLink(id uint) error
	GetActiveShareLinkByListmakId(listmakId uint) (*models.ShareLink, error)
}

type shareLinkRepository struct {
	db *gorm.DB
}

func NewShareLinkRepository(db *gorm.DB) ShareLinkRepository {
	return &shareLinkRepository{db: db}
}

func (r *shareLinkRepository) CreateShareLink(shareLink models.ShareLink) (models.ShareLink, error) {
	if err := r.db.Create(&shareLink).Error; err != nil {
		return models.ShareLink{}, err
	}
	return shareLink, nil
}

func (r *shareLinkRepository) GetShareLinkByShareId(shareId string) (models.ShareLink, error) {
	var shareLink models.ShareLink
	// Preload Listmak to get date etc (needed for validation/display)
	err := r.db.Preload("Listmak").Where("share_id = ?", shareId).First(&shareLink).Error
	if err != nil {
		return models.ShareLink{}, err
	}

	// Check if active and not expired (logic can be here or service, let's keep repo simple: just fetch)
	// But api logic says "if expired return error EXPIRED".
	// Fetching raw data is repository responsibility.

	return shareLink, nil
}

func (r *shareLinkRepository) DeleteShareLink(id uint) error {
	return r.db.Delete(&models.ShareLink{}, id).Error
}

func (r *shareLinkRepository) GetActiveShareLinkByListmakId(listmakId uint) (*models.ShareLink, error) {
	var shareLink models.ShareLink
	err := r.db.
		Where("listmak_id = ? AND is_active = true AND expires_at > NOW()", listmakId).
		Order("created_at DESC").
		Limit(1).
		First(&shareLink).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &shareLink, nil
}

// ------

type ViewShareRepository interface {
	CreateViewShare(viewShare models.ViewShare) (models.ViewShare, error)
	GetViewShareByViewId(viewId string) (models.ViewShare, error)
	GetViewShareByListmakId(listmakId uint) (*models.ViewShare, error)
	Delete(id uint) error
}

type viewShareRepository struct {
	db *gorm.DB
}

func NewViewShareRepository(db *gorm.DB) ViewShareRepository {
	return &viewShareRepository{db: db}
}

func (r *viewShareRepository) CreateViewShare(viewShare models.ViewShare) (models.ViewShare, error) {
	if err := r.db.Create(&viewShare).Error; err != nil {
		return models.ViewShare{}, err
	}
	return viewShare, nil
}

func (r *viewShareRepository) GetViewShareByViewId(viewId string) (models.ViewShare, error) {
	var viewShare models.ViewShare
	// Preload Listmak logic might be in snapshot data but reference is good
	err := r.db.Preload("Listmak").Where("view_id = ?", viewId).First(&viewShare).Error
	if err != nil {
		return models.ViewShare{}, err
	}
	return viewShare, nil
}

func (r *viewShareRepository) Delete(id uint) error {
	return r.db.Delete(&models.ViewShare{}, id).Error
}

func (r *viewShareRepository) GetViewShareByListmakId(listmakId uint) (*models.ViewShare, error) {
	var viewShare models.ViewShare
	err := r.db.
		Where("listmak_id = ?", listmakId).
		Order("created_at DESC").
		Limit(1).
		First(&viewShare).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &viewShare, nil
}

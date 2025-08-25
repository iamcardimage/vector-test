package repository

import (
	"context"
	"time"
	"vector/internal/models"

	"gorm.io/gorm"
)

type clientGormRepo struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) ClientRepository {
	return &clientGormRepo{db: db}
}

func (r *clientGormRepo) GetCurrentVersion(ctx context.Context, clientID int) (*models.ClientVersion, error) {
	var version models.ClientVersion
	err := r.db.WithContext(ctx).
		Where("client_id = ? AND is_current = true", clientID).
		Take(&version).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &version, err
}

func (r *clientGormRepo) CreateVersion(ctx context.Context, version *models.ClientVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

func (r *clientGormRepo) UpdateCurrentVersionStatus(ctx context.Context, clientID int, isCurrent bool, validTo *time.Time) error {
	updates := map[string]any{"is_current": isCurrent}
	if validTo != nil {
		updates["valid_to"] = *validTo
	}

	return r.db.WithContext(ctx).Model(&models.ClientVersion{}).
		Where("client_id = ? AND is_current = true", clientID).
		Updates(updates).Error
}

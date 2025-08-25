package repository

import (
	"context"
	"vector/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type stagingGormRepo struct {
	db *gorm.DB
}

func NewStagingRepository(db *gorm.DB) StagingRepository {
	return &stagingGormRepo{db: db}
}

func (r *stagingGormRepo) UpsertUsers(ctx context.Context, users []models.StagingExternalUser) error {
	if len(users) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"raw", "synced_at"}),
	}).Create(&users).Error
}

package repository

import (
	"context"
	"vector/internal/db/sync"
	"vector/internal/models"

	"gorm.io/gorm"
)

type syncStagingRepository struct {
	database *gorm.DB
}

func NewSyncStagingRepository(database *gorm.DB) SyncStagingRepository {
	return &syncStagingRepository{database: database}
}

func (r *syncStagingRepository) UpsertUsers(ctx context.Context, users []models.StagingExternalUser) error {
	return sync.UpsertStagingExternalUsers(r.database, users)
}

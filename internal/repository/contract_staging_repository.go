package repository

import (
	"context"
	"vector/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type contractStagingRepository struct {
	database *gorm.DB
}

func NewContractStagingRepository(database *gorm.DB) ContractStagingRepository {
	return &contractStagingRepository{database: database}
}

func (r *contractStagingRepository) UpsertContracts(ctx context.Context, contracts []models.StagingExternalContract) error {
	if len(contracts) == 0 {
		return nil
	}

	return r.database.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"raw", "synced_at"}),
		}).
		Create(&contracts).Error
}

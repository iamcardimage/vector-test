package sync

import (
	"vector/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func UpsertStagingExternalUsers(gdb *gorm.DB, items []models.StagingExternalUser) error {
	if len(items) == 0 {
		return nil
	}
	return gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"raw", "synced_at"}),
	}).Create(&items).Error
}

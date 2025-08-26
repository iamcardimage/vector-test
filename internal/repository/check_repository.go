package repository

import (
	appdb "vector/internal/db/app"
	"vector/internal/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type checkRepository struct {
	database *gorm.DB
}

func NewCheckRepository(database *gorm.DB) CheckRepository {
	return &checkRepository{database: database}
}

func (r *checkRepository) CreateSecondPartCheck(clientID, spVersion int, kind string, payload *datatypes.JSON, runBy *int) (models.SecondPartCheck, error) {
	return appdb.CreateSecondPartCheck(r.database, clientID, spVersion, kind, payload, runBy) // ИСПРАВЛЕНО
}

func (r *checkRepository) UpdateResult(checkID uint, status string, result *datatypes.JSON) (models.SecondPartCheck, error) {
	return appdb.UpdateSecondPartCheckResult(r.database, checkID, status, result) // ИСПРАВЛЕНО
}

func (r *checkRepository) ListByClient(clientID int, spVersion *int) ([]models.SecondPartCheck, error) {
	return appdb.ListSecondPartChecks(r.database, clientID, spVersion) // ИСПРАВЛЕНО
}

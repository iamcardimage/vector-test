package repository

import (
	appdb "vector/internal/db/app"

	"gorm.io/gorm"
)

type recalcRepository struct {
	database *gorm.DB
}

func NewRecalcRepository(database *gorm.DB) RecalcRepository {
	return &recalcRepository{database: database}
}

func (r *recalcRepository) RecalcNeedsSecondPart() (int64, error) {
	return appdb.RecalcNeedsSecondPart(r.database)
}

func (r *recalcRepository) RecalcPassportExpiry() (int64, error) {
	return appdb.RecalcPassportExpiry(r.database)
}

func (r *recalcRepository) RecalcAll() error {
	return appdb.RecalcAll(r.database)
}

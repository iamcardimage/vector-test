package repository

import (
	"context"
	syncdb "vector/internal/db/sync"
	"vector/internal/models"

	"gorm.io/gorm"
)

type syncContractRepository struct {
	database *gorm.DB
}

func NewSyncContractRepository(database *gorm.DB) SyncContractRepository {
	return &syncContractRepository{database: database}
}

func (r *syncContractRepository) GetCurrentContract(ctx context.Context, contractID int) (*models.Contract, error) {
	return syncdb.GetCurrentContract(r.database, contractID)
}

func (r *syncContractRepository) ListContracts(page, perPage int, userID *int, status *string) ([]models.Contract, int64, error) {
	return syncdb.ListContracts(r.database, page, perPage, userID, status)
}

func (r *syncContractRepository) ApplyContractsBatch(ctx context.Context, contracts []ApplyContractData) (ApplyStats, error) {
	// Конвертируем repository.ApplyContractData в syncdb.ApplyContractData
	syncData := make([]syncdb.ApplyContractData, len(contracts))
	for i, c := range contracts {
		syncData[i] = syncdb.ApplyContractData{
			ContractID: c.ContractID,
			RawData:    c.RawData,
			Hash:       c.Hash,
		}
	}

	// Вызываем DB слой
	syncStats, err := syncdb.ApplyContractsBatch(r.database, ctx, syncData)
	if err != nil {
		return ApplyStats{}, err
	}

	// Конвертируем обратно в repository.ApplyStats
	return ApplyStats{
		Created:   syncStats.Created,
		Updated:   syncStats.Updated,
		Unchanged: syncStats.Unchanged,
	}, nil
}

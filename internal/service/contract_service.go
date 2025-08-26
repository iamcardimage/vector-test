package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
	"vector/internal/models"
	"vector/internal/pkg/utils"
	"vector/internal/repository"

	"gorm.io/datatypes"
)

type ContractService struct {
	stagingRepo  repository.ContractStagingRepository
	contractRepo repository.SyncContractRepository
	externalAPI  repository.ExternalAPIClient
}

func NewContractService(
	stagingRepo repository.ContractStagingRepository,
	contractRepo repository.SyncContractRepository,
	externalAPI repository.ExternalAPIClient,
) *ContractService {
	return &ContractService{
		stagingRepo:  stagingRepo,
		contractRepo: contractRepo,
		externalAPI:  externalAPI,
	}
}

type SyncContractsRequest struct {
	Page    int
	PerPage int
}

type SyncContractsResponse struct {
	Success    bool `json:"success"`
	Applied    int  `json:"applied"`
	Created    int  `json:"created"`
	Updated    int  `json:"updated"`
	Unchanged  int  `json:"unchanged"`
	Page       int  `json:"page"`
	TotalPages int  `json:"total_pages"`
	TotalCount int  `json:"total_count"`
	PerPage    int  `json:"per_page"`
}

func (s *ContractService) SyncContracts(ctx context.Context, req SyncContractsRequest) (*SyncContractsResponse, error) {
	// 1. Получаем данные из внешнего API
	resp, err := s.externalAPI.GetContractsRaw(ctx, req.Page, req.PerPage)
	if err != nil {
		return nil, err
	}

	// 2. Сохраняем в staging
	stagingBatch := s.prepareStagingBatch(resp.Contracts)
	if err := s.stagingRepo.UpsertContracts(ctx, stagingBatch); err != nil {
		return nil, err
	}

	// 3. Применяем изменения
	applyBatch := s.prepareApplyBatch(resp.Contracts)
	stats, err := s.contractRepo.ApplyContractsBatch(ctx, applyBatch)
	if err != nil {
		return nil, err
	}

	return &SyncContractsResponse{
		Success:    true,
		Applied:    len(stagingBatch),
		Created:    stats.Created,
		Updated:    stats.Updated,
		Unchanged:  stats.Unchanged,
		Page:       resp.CurrentPage,
		TotalPages: resp.TotalPages,
		TotalCount: resp.TotalCount,
		PerPage:    resp.PerPage,
	}, nil
}

func (s *ContractService) prepareStagingBatch(rawContracts []json.RawMessage) []models.StagingExternalContract {
	now := time.Now().UTC()
	batch := make([]models.StagingExternalContract, 0, len(rawContracts))

	for _, r := range rawContracts {
		contractID, err := utils.ExtractContractID(r)
		if err != nil {
			continue
		}

		batch = append(batch, models.StagingExternalContract{
			ID:       contractID,
			Raw:      datatypes.JSON(r),
			SyncedAt: now,
		})
	}

	return batch
}

func (s *ContractService) prepareApplyBatch(rawContracts []json.RawMessage) []repository.ApplyContractData {
	batch := make([]repository.ApplyContractData, 0, len(rawContracts))

	for _, r := range rawContracts {
		contractID, err := utils.ExtractContractID(r)
		if err != nil {
			continue
		}

		// Простой хэш для отслеживания изменений (без триггеров)
		hash := sha256.Sum256(r)
		hashStr := hex.EncodeToString(hash[:])

		batch = append(batch, repository.ApplyContractData{
			ContractID: contractID,
			RawData:    r,
			Hash:       hashStr,
		})
	}

	return batch
}

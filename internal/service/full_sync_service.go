package service

import (
	"context"
	"log"
	"vector/internal/repository"
)

type FullSyncService struct {
	applyService    *ApplyService
	contractService *ContractService
	externalAPI     repository.ExternalAPIClient
}

func NewFullSyncService(
	applyService *ApplyService,
	contractService *ContractService,
	externalAPI repository.ExternalAPIClient,
) *FullSyncService {
	return &FullSyncService{
		applyService:    applyService,
		contractService: contractService,
		externalAPI:     externalAPI,
	}
}

type FullSyncRequest struct {
	PerPage       int
	SyncContracts bool
}

type FullSyncResponse struct {
	Success bool `json:"success"`

	// Статистика по пользователям
	UserPages     int `json:"user_pages"`
	UserSaved     int `json:"user_saved"`
	UserApplied   int `json:"user_applied"`
	UserCreated   int `json:"user_created"`
	UserUpdated   int `json:"user_updated"`
	UserUnchanged int `json:"user_unchanged"`

	// Статистика по договорам
	ContractPages     int `json:"contract_pages"`
	ContractSaved     int `json:"contract_saved"`
	ContractApplied   int `json:"contract_applied"`
	ContractCreated   int `json:"contract_created"`
	ContractUpdated   int `json:"contract_updated"`
	ContractUnchanged int `json:"contract_unchanged"`

	// Общая статистика (для обратной совместимости)
	Pages     int `json:"pages"`
	Saved     int `json:"saved"`
	Applied   int `json:"applied"`
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Unchanged int `json:"unchanged"`
}

func (s *FullSyncService) SyncFull(ctx context.Context, req FullSyncRequest) (*FullSyncResponse, error) {
	if req.PerPage <= 0 {
		req.PerPage = 100
	}

	stats := &FullSyncResponse{Success: true}

	// 1. Синхронизация пользователей (как было)
	log.Println("[full-sync] Starting users synchronization...")
	if err := s.syncUsers(ctx, req.PerPage, stats); err != nil {
		return nil, err
	}

	// 2. Синхронизация договоров (НОВОЕ)
	if req.SyncContracts {
		log.Println("[full-sync] Starting contracts synchronization...")
		if err := s.syncContracts(ctx, req.PerPage, stats); err != nil {
			return nil, err
		}
	}

	// Общая статистика
	stats.Pages = stats.UserPages + stats.ContractPages
	stats.Saved = stats.UserSaved + stats.ContractSaved
	stats.Applied = stats.UserApplied + stats.ContractApplied
	stats.Created = stats.UserCreated + stats.ContractCreated
	stats.Updated = stats.UserUpdated + stats.ContractUpdated
	stats.Unchanged = stats.UserUnchanged + stats.ContractUnchanged

	log.Printf("[full-sync] Completed. Users: %d pages, %d applied. Contracts: %d pages, %d applied",
		stats.UserPages, stats.UserApplied, stats.ContractPages, stats.ContractApplied)

	return stats, nil
}

func (s *FullSyncService) syncUsers(ctx context.Context, perPage int, stats *FullSyncResponse) error {
	first, err := s.externalAPI.GetUsersRaw(ctx, 1, perPage)
	if err != nil {
		return err
	}

	totalPages := first.TotalPages
	if totalPages <= 0 {
		totalPages = 1
	}

	for page := 1; page <= totalPages; page++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		applyResp, err := s.applyService.SyncApply(ctx, SyncApplyRequest{
			Page:    page,
			PerPage: perPage,
		})
		if err != nil {
			return err
		}

		stats.UserPages++
		stats.UserSaved += applyResp.Applied
		stats.UserApplied += applyResp.Applied
		stats.UserCreated += applyResp.Created
		stats.UserUpdated += applyResp.Updated
		stats.UserUnchanged += applyResp.Unchanged
	}

	return nil
}

func (s *FullSyncService) syncContracts(ctx context.Context, perPage int, stats *FullSyncResponse) error {
	first, err := s.externalAPI.GetContractsRaw(ctx, 1, perPage)
	if err != nil {
		return err
	}

	totalPages := first.TotalPages
	if totalPages <= 0 {
		totalPages = 1
	}

	for page := 1; page <= totalPages; page++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		contractResp, err := s.contractService.SyncContracts(ctx, SyncContractsRequest{
			Page:    page,
			PerPage: perPage,
		})
		if err != nil {
			return err
		}

		stats.ContractPages++
		stats.ContractSaved += contractResp.Applied
		stats.ContractApplied += contractResp.Applied
		stats.ContractCreated += contractResp.Created
		stats.ContractUpdated += contractResp.Updated
		stats.ContractUnchanged += contractResp.Unchanged
	}

	return nil
}

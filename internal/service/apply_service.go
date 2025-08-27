package service

import (
	"context"
	"encoding/json"
	"time"
	"vector/internal/models"
	"vector/internal/pkg/utils"
	"vector/internal/repository"

	"gorm.io/datatypes"
)

type ApplyService struct {
	stagingRepo    repository.SyncStagingRepository
	clientRepo     repository.SyncClientRepository
	externalAPI    repository.ExternalAPIClient
	triggerService *TriggerService
}

func NewApplyService(
	stagingRepo repository.SyncStagingRepository,
	clientRepo repository.SyncClientRepository,
	externalAPI repository.ExternalAPIClient,
	triggerService *TriggerService,
) *ApplyService {
	return &ApplyService{
		stagingRepo:    stagingRepo,
		clientRepo:     clientRepo,
		externalAPI:    externalAPI,
		triggerService: triggerService,
	}
}

type SyncApplyRequest struct {
	Page    int
	PerPage int
}

type SyncApplyResponse struct {
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

func (s *ApplyService) SyncApply(ctx context.Context, req SyncApplyRequest) (*SyncApplyResponse, error) {

	resp, err := s.externalAPI.GetUsersRaw(ctx, req.Page, req.PerPage)
	if err != nil {
		return nil, err
	}

	stagingBatch := s.prepareStagingBatch(resp.Users)
	if err := s.stagingRepo.UpsertUsers(ctx, stagingBatch); err != nil {
		return nil, err
	}

	applyBatch := s.prepareApplyBatch(resp.Users)
	stats, err := s.clientRepo.ApplyUsersBatch(ctx, applyBatch)
	if err != nil {
		return nil, err
	}

	return &SyncApplyResponse{
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

func (s *ApplyService) prepareStagingBatch(rawUsers []json.RawMessage) []models.StagingExternalUser {
	now := time.Now().UTC()
	batch := make([]models.StagingExternalUser, 0, len(rawUsers))

	for _, r := range rawUsers {
		userID, err := utils.ExtractUserID(r)
		if err != nil {
			continue
		}

		batch = append(batch, models.StagingExternalUser{
			ID:       userID,
			Raw:      datatypes.JSON(r),
			SyncedAt: now,
		})
	}

	return batch
}

func (s *ApplyService) prepareApplyBatch(rawUsers []json.RawMessage) []repository.ApplyUserData {
	batch := make([]repository.ApplyUserData, 0, len(rawUsers))

	for _, r := range rawUsers {
		userID, err := utils.ExtractUserID(r)
		if err != nil {
			continue
		}

		triggerHash, err := s.triggerService.ComputeSecondPartTriggerHash(r)
		if err != nil {
			continue
		}

		batch = append(batch, repository.ApplyUserData{
			UserID:      userID,
			RawData:     r,
			TriggerHash: triggerHash,
		})
	}

	return batch
}

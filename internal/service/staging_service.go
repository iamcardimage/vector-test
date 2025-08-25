package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"vector/internal/models"
	"vector/internal/repository"

	"gorm.io/datatypes"
)

type StagingService struct {
	stagingRepo repository.StagingRepository
	externalAPI repository.ExternalAPIClient
}

func NewStagingService(stagingRepo repository.StagingRepository, externalAPI repository.ExternalAPIClient) *StagingService {
	return &StagingService{
		stagingRepo: stagingRepo,
		externalAPI: externalAPI,
	}
}

type SyncStagingRequest struct {
	Page    int
	PerPage int
}

type SyncStagingResponse struct {
	Success    bool `json:"success"`
	Saved      int  `json:"saved"`
	Page       int  `json:"page"`
	TotalPages int  `json:"total_pages"`
	TotalCount int  `json:"total_count"`
	PerPage    int  `json:"per_page"`
}

func (s *StagingService) SyncStaging(ctx context.Context, req SyncStagingRequest) (*SyncStagingResponse, error) {
	resp, err := s.externalAPI.GetUsersRaw(ctx, req.Page, req.PerPage)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	batch := make([]models.StagingExternalUser, 0, len(resp.Users))

	for _, r := range resp.Users {
		id, err := extractUserID(r)
		if err != nil {
			continue
		}

		batch = append(batch, models.StagingExternalUser{
			ID:       id,
			Raw:      datatypes.JSON(r),
			SyncedAt: now,
		})
	}

	if err := s.stagingRepo.UpsertUsers(ctx, batch); err != nil {
		return nil, err
	}

	return &SyncStagingResponse{
		Success:    true,
		Saved:      len(batch),
		Page:       resp.CurrentPage,
		TotalPages: resp.TotalPages,
		TotalCount: resp.TotalCount,
		PerPage:    resp.PerPage,
	}, nil
}

func extractUserID(raw json.RawMessage) (int, error) {
	var tmp map[string]any
	if err := json.Unmarshal(raw, &tmp); err != nil {
		return 0, err
	}
	idVal, ok := tmp["id"]
	if !ok {
		return 0, errors.New("id field not found")
	}
	switch v := idVal.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, errors.New("invalid id type")
	}

}

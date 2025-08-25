package service

import (
	"context"
	"vector/internal/repository"
)

type FullSyncService struct {
	applyService *ApplyService
	externalAPI  repository.ExternalAPIClient
}

func NewFullSyncService(applyService *ApplyService, externalAPI repository.ExternalAPIClient) *FullSyncService {
	return &FullSyncService{
		applyService: applyService,
		externalAPI:  externalAPI,
	}
}

type FullSyncRequest struct {
	PerPage int
}

type FullSyncResponse struct {
	Success   bool `json:"success"`
	Pages     int  `json:"pages"`
	Saved     int  `json:"saved"`
	Applied   int  `json:"applied"`
	Created   int  `json:"created"`
	Updated   int  `json:"updated"`
	Unchanged int  `json:"unchanged"`
}

func (s *FullSyncService) SyncFull(ctx context.Context, req FullSyncRequest) (*FullSyncResponse, error) {
	if req.PerPage <= 0 {
		req.PerPage = 100
	}

	stats := &FullSyncResponse{Success: true}

	first, err := s.externalAPI.GetUsersRaw(ctx, 1, req.PerPage)
	if err != nil {
		return nil, err
	}

	totalPages := first.TotalPages
	if totalPages <= 0 {
		totalPages = 1
	}

	for page := 1; page <= totalPages; page++ {

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		applyResp, err := s.applyService.SyncApply(ctx, SyncApplyRequest{
			Page:    page,
			PerPage: req.PerPage,
		})
		if err != nil {
			return nil, err
		}

		stats.Pages++
		stats.Saved += applyResp.Applied
		stats.Applied += applyResp.Applied
		stats.Created += applyResp.Created
		stats.Updated += applyResp.Updated
		stats.Unchanged += applyResp.Unchanged
	}

	return stats, nil
}

package repository

import (
	"context"
	"encoding/json"
	"time"
	"vector/internal/models"
)

type StagingRepository interface {
	UpsertUsers(ctx context.Context, users []models.StagingExternalUser) error
}

type ExternalAPIClient interface {
	GetUsersRaw(ctx context.Context, page, perPage int) (*ExternalUsersResponse, error)
}

type ExternalUsersResponse struct {
	Success     bool              `json:"success"`
	TotalCount  int               `json:"total_count"`
	PerPage     int               `json:"per_page"`
	CurrentPage int               `json:"current_page"`
	TotalPages  int               `json:"total_pages"`
	Users       []json.RawMessage `json:"users"`
}

type ClientRepository interface {
	GetCurrentVersion(ctx context.Context, clientID int) (*models.ClientVersion, error)
	CreateVersion(ctx context.Context, version *models.ClientVersion) error
	UpdateCurrentVersionStatus(ctx context.Context, clientID int, isCurrent bool, validTo *time.Time) error
}

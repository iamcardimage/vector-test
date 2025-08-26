package repository

import (
	"context"
	"encoding/json"
	"time"
	"vector/internal/models"
)

type SyncStagingRepository interface {
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

type SyncClientRepository interface {
	GetCurrentVersion(ctx context.Context, clientID int) (*models.ClientVersion, error)
	CreateVersion(ctx context.Context, version *models.ClientVersion) error
	UpdateCurrentVersionStatus(ctx context.Context, clientID int, isCurrent bool, validTo *time.Time) error
	ListCurrentClients(page, perPage int, needsSecondPart *bool) ([]models.ClientListItem, int64, error)
	ApplyUsersBatch(ctx context.Context, users []ApplyUserData) (ApplyStats, error)
}

type ApplyUserData struct {
	UserID      int             `json:"user_id"`
	RawData     json.RawMessage `json:"raw_data"`
	TriggerHash string          `json:"trigger_hash"`
}

type ApplyStats struct {
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Unchanged int `json:"unchanged"`
}

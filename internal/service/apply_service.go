package service

import (
	"context"
	"encoding/json"
	"time"
	"vector/internal/models"
	"vector/internal/repository"
	"vector/internal/syncer"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ApplyService struct {
	stagingRepo repository.StagingRepository
	clientRepo  repository.ClientRepository
	externalAPI repository.ExternalAPIClient
	db          *gorm.DB
}

func NewApplyService(
	stagingRepo repository.StagingRepository,
	clientRepo repository.ClientRepository,
	externalAPI repository.ExternalAPIClient,
	db *gorm.DB,
) *ApplyService {
	return &ApplyService{
		stagingRepo: stagingRepo,
		clientRepo:  clientRepo,
		externalAPI: externalAPI,
		db:          db,
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

type ApplyStats struct {
	Created   int
	Updated   int
	Unchanged int
}

func (s *ApplyService) SyncApply(ctx context.Context, req SyncApplyRequest) (*SyncApplyResponse, error) {

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

	stats, err := s.applyUsersBatch(ctx, resp.Users)
	if err != nil {
		return nil, err
	}

	return &SyncApplyResponse{
		Success:    true,
		Applied:    len(batch),
		Created:    stats.Created,
		Updated:    stats.Updated,
		Unchanged:  stats.Unchanged,
		Page:       resp.CurrentPage,
		TotalPages: resp.TotalPages,
		TotalCount: resp.TotalCount,
		PerPage:    resp.PerPage,
	}, nil
}

func (s *ApplyService) applyUsersBatch(ctx context.Context, users []json.RawMessage) (*ApplyStats, error) {
	stats := &ApplyStats{}

	return stats, s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, raw := range users {
			if err := s.applyUser(ctx, raw, stats); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *ApplyService) applyUser(ctx context.Context, raw json.RawMessage, stats *ApplyStats) error {

	userID, err := extractUserID(raw)
	if err != nil {
		return nil
	}

	triggerHash, err := syncer.ComputeSecondPartTriggerHash(raw)
	if err != nil {
		return nil
	}

	externalRisk := syncer.ExtractExternalRiskLevel(raw)

	current, err := s.clientRepo.GetCurrentVersion(ctx, userID)
	if err != nil {
		return err
	}

	userData, err := extractUserData(raw)
	if err != nil {
		return nil
	}

	now := time.Now().UTC()

	if current == nil {

		newVersion := &models.ClientVersion{
			ClientID:              userID,
			Version:               1,
			Surname:               userData.Surname,
			Name:                  userData.Name,
			Patronymic:            userData.Patronymic,
			Birthday:              userData.Birthday,
			BirthPlace:            userData.BirthPlace,
			ContactEmail:          userData.ContactEmail,
			Inn:                   userData.Inn,
			Snils:                 userData.Snils,
			CreatedLKAt:           userData.CreatedLKAt,
			UpdatedLKAt:           userData.UpdatedLKAt,
			PassIssuerCode:        userData.PassIssuerCode,
			PassSeries:            userData.PassSeries,
			PassNumber:            userData.PassNumber,
			PassIssueDate:         userData.PassIssueDate,
			PassIssuer:            userData.PassIssuer,
			MainPhone:             userData.MainPhone,
			ExternalRiskLevel:     externalRisk,
			SecondPartTriggerHash: triggerHash,
			NeedsSecondPart:       true,
			SecondPartCreated:     false,
			Hash:                  triggerHash,
			Raw:                   datatypes.JSON(raw),
			SyncedAt:              now,
			ValidFrom:             now,
			IsCurrent:             true,
			Status:                "changed",
		}

		if err := s.clientRepo.CreateVersion(ctx, newVersion); err != nil {
			return err
		}
		stats.Created++
		return nil
	}

	if current.SecondPartTriggerHash == triggerHash {
		stats.Unchanged++
		return nil
	}

	if err := s.clientRepo.UpdateCurrentVersionStatus(ctx, userID, false, &now); err != nil {
		return err
	}

	newVersion := &models.ClientVersion{
		ClientID:              userID,
		Version:               current.Version + 1,
		Surname:               userData.Surname,
		Name:                  userData.Name,
		Patronymic:            userData.Patronymic,
		Birthday:              userData.Birthday,
		BirthPlace:            userData.BirthPlace,
		ContactEmail:          userData.ContactEmail,
		Inn:                   userData.Inn,
		Snils:                 userData.Snils,
		CreatedLKAt:           userData.CreatedLKAt,
		UpdatedLKAt:           userData.UpdatedLKAt,
		PassIssuerCode:        userData.PassIssuerCode,
		PassSeries:            userData.PassSeries,
		PassNumber:            userData.PassNumber,
		PassIssueDate:         userData.PassIssueDate,
		PassIssuer:            userData.PassIssuer,
		MainPhone:             userData.MainPhone,
		ExternalRiskLevel:     externalRisk,
		SecondPartTriggerHash: triggerHash,
		NeedsSecondPart:       true,
		SecondPartCreated:     current.SecondPartCreated,
		Hash:                  triggerHash,
		Raw:                   datatypes.JSON(raw),
		SyncedAt:              now,
		ValidFrom:             now,
		IsCurrent:             true,
		Status:                "changed",
	}

	if err := s.clientRepo.CreateVersion(ctx, newVersion); err != nil {
		return err
	}
	stats.Updated++
	return nil
}

type UserData struct {
	Surname        string
	Name           string
	Patronymic     string
	Birthday       string
	BirthPlace     string
	ContactEmail   string
	Inn            string
	Snils          string
	CreatedLKAt    string
	UpdatedLKAt    string
	PassIssuerCode string
	PassSeries     string
	PassNumber     string
	PassIssueDate  string
	PassIssuer     string
	MainPhone      string
}

func extractUserData(raw json.RawMessage) (*UserData, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}

	extractString := func(key string) string {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}

		if pi, ok := m["person_info"].(map[string]any); ok {
			if v, ok := pi[key]; ok {
				if s, ok := v.(string); ok {
					return s
				}
			}
		}
		return ""
	}

	return &UserData{
		Surname:        extractString("surname"),
		Name:           extractString("name"),
		Patronymic:     extractString("patronymic"),
		Birthday:       extractString("birthday"),
		BirthPlace:     extractString("birth_place"),
		ContactEmail:   extractString("contact_email"),
		Inn:            extractString("inn"),
		Snils:          extractString("snils"),
		CreatedLKAt:    extractString("created_lk_at"),
		UpdatedLKAt:    extractString("updated_lk_at"),
		PassIssuerCode: extractString("pass_issuer_code"),
		PassSeries:     extractString("pass_series"),
		PassNumber:     extractString("pass_number"),
		PassIssueDate:  extractString("pass_issue_date"),
		PassIssuer:     extractString("pass_issuer"),
		MainPhone:      extractString("main_phone"),
	}, nil
}

package service

import (
	"context"
	"encoding/json"
	"time"
	"vector/internal/models"
	"vector/internal/pkg/utils"
	"vector/internal/repository"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ApplyService struct {
	stagingRepo    repository.StagingRepository
	clientRepo     repository.ClientRepository
	externalAPI    repository.ExternalAPIClient
	triggerService *TriggerService
	db             *gorm.DB
}

func NewApplyService(
	stagingRepo repository.StagingRepository,
	clientRepo repository.ClientRepository,
	externalAPI repository.ExternalAPIClient,
	triggerService *TriggerService,
	db *gorm.DB,
) *ApplyService {
	return &ApplyService{
		stagingRepo:    stagingRepo,
		clientRepo:     clientRepo,
		externalAPI:    externalAPI,
		triggerService: triggerService,
		db:             db,
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

type rawUser struct {
	ID int `json:"id"`
}

func (s *ApplyService) SyncApply(ctx context.Context, req SyncApplyRequest) (*SyncApplyResponse, error) {
	// Получаем данные из внешнего API
	resp, err := s.externalAPI.GetUsersRaw(ctx, req.Page, req.PerPage)
	if err != nil {
		return nil, err
	}

	// Сначала сохраняем в staging
	stagingBatch := make([]models.StagingExternalUser, 0, len(resp.Users))
	now := time.Now().UTC()

	for _, r := range resp.Users {
		userID, err := utils.ExtractUserID(r) // ИСПОЛЬЗУЕМ ОБЩУЮ ФУНКЦИЮ
		if err != nil {
			continue
		}

		stagingBatch = append(stagingBatch, models.StagingExternalUser{
			ID:       userID,
			Raw:      datatypes.JSON(r),
			SyncedAt: now,
		})
	}

	if err := s.stagingRepo.UpsertUsers(ctx, stagingBatch); err != nil {
		return nil, err
	}

	// Теперь применяем изменения в core
	stats, err := s.applyUsersBatch(ctx, resp.Users)
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

func (s *ApplyService) applyUsersBatch(ctx context.Context, raws []json.RawMessage) (ApplyStats, error) {
	stats := ApplyStats{}
	now := time.Now().UTC()

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, r := range raws {
			var u rawUser
			if err := json.Unmarshal(r, &u); err != nil || u.ID == 0 {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(r, &m); err != nil {
				continue
			}

			triggerHash, err := s.triggerService.ComputeSecondPartTriggerHash(r)
			if err != nil {
				continue
			}
			externalRisk := s.triggerService.ExtractExternalRiskLevel(r)

			// Получаем текущую версию клиента
			var cur models.ClientVersion
			err = tx.Where("client_id = ? AND is_current = true", u.ID).Take(&cur).Error

			// Извлекаем поля ИСПОЛЬЗУЯ ОБЩУЮ ФУНКЦИЮ
			name := utils.ExtractString(m, "name")
			surname := utils.ExtractString(m, "surname")
			patronymic := utils.ExtractString(m, "patronymic")
			birthday := utils.ExtractString(m, "birthday")
			birthPlace := utils.ExtractString(m, "birth_place")
			contactEmail := utils.ExtractString(m, "contact_email")
			inn := utils.ExtractString(m, "inn")
			snils := utils.ExtractString(m, "snils")
			createdLKAt := utils.ExtractString(m, "created_lk_at")
			updatedLKAt := utils.ExtractString(m, "updated_lk_at")
			passIssuerCode := utils.ExtractString(m, "pass_issuer_code")
			passSeries := utils.ExtractString(m, "pass_series")
			passNumber := utils.ExtractString(m, "pass_number")
			passIssueDate := utils.ExtractString(m, "pass_issue_date")
			passIssuer := utils.ExtractString(m, "pass_issuer")
			mainPhone := utils.ExtractString(m, "main_phone")

			if err == gorm.ErrRecordNotFound {
				// Создаем новый клиент
				nv := models.ClientVersion{
					ClientID:              u.ID,
					Version:               1,
					Surname:               surname,
					Name:                  name,
					Patronymic:            patronymic,
					Birthday:              birthday,
					BirthPlace:            birthPlace,
					ContactEmail:          contactEmail,
					Inn:                   inn,
					Snils:                 snils,
					CreatedLKAt:           createdLKAt,
					UpdatedLKAt:           updatedLKAt,
					PassIssuerCode:        passIssuerCode,
					PassSeries:            passSeries,
					PassNumber:            passNumber,
					PassIssueDate:         passIssueDate,
					PassIssuer:            passIssuer,
					MainPhone:             mainPhone,
					ExternalRiskLevel:     externalRisk,
					SecondPartTriggerHash: triggerHash,
					NeedsSecondPart:       true,
					SecondPartCreated:     false,
					Hash:                  triggerHash,
					Raw:                   datatypes.JSON(r),
					SyncedAt:              now,
					ValidFrom:             now,
					ValidTo:               nil,
					IsCurrent:             true,
					Status:                "changed",
				}
				if err := tx.Create(&nv).Error; err != nil {
					return err
				}
				stats.Created++
				continue
			}
			if err != nil {
				return err
			}

			// Ничего не изменилось по триггер-полям
			if cur.SecondPartTriggerHash == triggerHash {
				stats.Unchanged++
				continue
			}

			// Изменения — новая версия (SCD2)
			if err := tx.Model(&models.ClientVersion{}).
				Where("client_id = ? AND is_current = true", u.ID).
				Updates(map[string]any{
					"is_current": false,
					"valid_to":   now,
				}).Error; err != nil {
				return err
			}

			nv := models.ClientVersion{
				ClientID:              u.ID,
				Version:               cur.Version + 1,
				Surname:               surname,
				Name:                  name,
				Patronymic:            patronymic,
				Birthday:              birthday,
				BirthPlace:            birthPlace,
				ContactEmail:          contactEmail,
				Inn:                   inn,
				Snils:                 snils,
				CreatedLKAt:           createdLKAt,
				UpdatedLKAt:           updatedLKAt,
				PassIssuerCode:        passIssuerCode,
				PassSeries:            passSeries,
				ExternalRiskLevel:     externalRisk,
				PassNumber:            passNumber,
				PassIssueDate:         passIssueDate,
				PassIssuer:            passIssuer,
				MainPhone:             mainPhone,
				SecondPartTriggerHash: triggerHash,
				NeedsSecondPart:       true,
				SecondPartCreated:     cur.SecondPartCreated,
				Hash:                  triggerHash,
				Raw:                   datatypes.JSON(r),
				SyncedAt:              now,
				ValidFrom:             now,
				ValidTo:               nil,
				IsCurrent:             true,
				Status:                "changed",
			}
			if err := tx.Create(&nv).Error; err != nil {
				return err
			}
			stats.Updated++
		}
		return nil
	})
	return stats, err
}

// УБРАЛИ ДУБЛИРОВАННЫЕ ФУНКЦИИ - используем utils

package repository

import (
	"context"
	"encoding/json"
	"time"
	syncdb "vector/internal/db/sync"
	"vector/internal/models"
	"vector/internal/pkg/utils"

	"gorm.io/gorm"
)

type syncClientRepository struct {
	database *gorm.DB
}

func NewSyncClientRepository(database *gorm.DB) SyncClientRepository {
	return &syncClientRepository{database: database}
}

func (r *syncClientRepository) GetCurrentVersion(ctx context.Context, clientID int) (*models.ClientVersion, error) {
	var version models.ClientVersion
	err := r.database.WithContext(ctx).
		Where("client_id = ? AND is_current = true", clientID).
		Take(&version).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &version, err
}

func (r *syncClientRepository) CreateVersion(ctx context.Context, version *models.ClientVersion) error {
	return r.database.WithContext(ctx).Create(version).Error
}

func (r *syncClientRepository) UpdateCurrentVersionStatus(ctx context.Context, clientID int, isCurrent bool, validTo *time.Time) error {
	updates := map[string]any{"is_current": isCurrent}
	if validTo != nil {
		updates["valid_to"] = *validTo
	}

	return r.database.WithContext(ctx).Model(&models.ClientVersion{}).
		Where("client_id = ? AND is_current = true", clientID).
		Updates(updates).Error
}

func (r *syncClientRepository) ListCurrentClients(page, perPage int, needsSecondPart *bool) ([]models.ClientListItem, int64, error) {
	// ОБНОВЛЕНО: используем новый пакет
	dbItems, total, err := syncdb.ListCurrentClients(r.database, page, perPage, needsSecondPart)
	if err != nil {
		return nil, 0, err
	}

	// Конвертируем syncdb.ClientListItem в models.ClientListItem
	items := make([]models.ClientListItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = models.ClientListItem{
			ClientID:          dbItem.ClientID,
			Surname:           dbItem.Surname,
			Name:              dbItem.Name,
			Patronymic:        dbItem.Patronymic,
			Birthday:          dbItem.Birthday,
			BirthPlace:        dbItem.BirthPlace,
			ContactEmail:      dbItem.ContactEmail,
			Inn:               dbItem.Inn,
			Snils:             dbItem.Snils,
			CreatedLKAt:       dbItem.CreatedLKAt,
			UpdatedLKAt:       dbItem.UpdatedLKAt,
			PassIssuerCode:    dbItem.PassIssuerCode,
			PassSeries:        dbItem.PassSeries,
			PassNumber:        dbItem.PassNumber,
			PassIssueDate:     dbItem.PassIssueDate,
			PassIssuer:        dbItem.PassIssuer,
			MainPhone:         dbItem.MainPhone,
			ExternalRiskLevel: dbItem.ExternalRiskLevel,
			NeedsSecondPart:   dbItem.NeedsSecondPart,
			SecondPartCreated: dbItem.SecondPartCreated,
		}
	}

	return items, total, nil
}

func (r *syncClientRepository) ApplyUsersBatch(ctx context.Context, users []ApplyUserData) (ApplyStats, error) {
	stats := ApplyStats{}
	now := time.Now().UTC()

	err := r.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, userData := range users {
			// Парсим raw data
			var m map[string]any
			if err := json.Unmarshal(userData.RawData, &m); err != nil {
				continue
			}

			// Получаем текущую версию клиента
			var cur models.ClientVersion
			err := tx.Where("client_id = ? AND is_current = true", userData.UserID).Take(&cur).Error

			if err == gorm.ErrRecordNotFound {
				// Создаем новый клиент
				newVersion := r.buildClientVersion(userData, m, 1, now)
				if err := tx.Create(&newVersion).Error; err != nil {
					return err
				}
				stats.Created++
				continue
			}
			if err != nil {
				return err
			}

			// Проверяем изменения по trigger hash
			if cur.SecondPartTriggerHash == userData.TriggerHash {
				stats.Unchanged++
				continue
			}

			// Есть изменения - создаем новую версию (SCD2)
			// 1. Закрываем текущую версию
			if err := tx.Model(&models.ClientVersion{}).
				Where("client_id = ? AND is_current = true", userData.UserID).
				Updates(map[string]any{
					"is_current": false,
					"valid_to":   now,
				}).Error; err != nil {
				return err
			}

			// 2. Создаем новую версию
			newVersion := r.buildClientVersion(userData, m, cur.Version+1, now)
			newVersion.SecondPartCreated = cur.SecondPartCreated // Сохраняем статус второй части

			if err := tx.Create(&newVersion).Error; err != nil {
				return err
			}
			stats.Updated++
		}
		return nil
	})

	return stats, err
}

// Вспомогательный метод для построения ClientVersion
func (r *syncClientRepository) buildClientVersion(userData ApplyUserData, m map[string]any, version int, now time.Time) models.ClientVersion {
	// Используем утилиту для полного парсинга
	client := utils.ParseClientVersion(userData.RawData)

	// Устанавливаем версионные поля
	client.ClientID = userData.UserID
	client.Version = version
	client.SecondPartTriggerHash = userData.TriggerHash
	client.Hash = userData.TriggerHash
	client.NeedsSecondPart = true
	client.Status = "changed"
	client.SyncedAt = now
	client.ValidFrom = now
	client.IsCurrent = true

	return client
}

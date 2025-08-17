package syncer

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"vector/internal/models"
)

type rawUser struct {
	ID int `json:"id"`
}

func extractString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	// дубли в person_info
	if pi, ok := m["person_info"].(map[string]any); ok {
		if v, ok := pi[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

type ApplyStats struct {
	Created   int
	Updated   int
	Unchanged int
}

// ApplyUsersBatch: для каждого raw создаёт новую версию, если изменился SecondPartTriggerHash.
// Не архивируем отсутствующих — только изменения.
func ApplyUsersBatch(gdb *gorm.DB, raws []json.RawMessage) (ApplyStats, error) {
	stats := ApplyStats{}
	now := time.Now().UTC()

	err := gdb.Transaction(func(tx *gorm.DB) error {
		for _, r := range raws {
			var u rawUser
			if err := json.Unmarshal(r, &u); err != nil || u.ID == 0 {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(r, &m); err != nil {
				continue
			}

			triggerHash, err := ComputeSecondPartTriggerHash(r)
			if err != nil {
				continue
			}
			externalRisk := ExtractExternalRiskLevel(r)

			var cur models.ClientVersion
			err = tx.Where("client_id = ? AND is_current = true", u.ID).Take(&cur).Error
			name := extractString(m, "name")
			surname := extractString(m, "surname")
			patronymic := extractString(m, "patronymic")

			if err == gorm.ErrRecordNotFound {
				nv := models.ClientVersion{
					ClientID:              u.ID,
					Version:               1,
					Surname:               surname,
					Name:                  name,
					Patronymic:            patronymic,
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

			// Изменения — новая версия
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
				ExternalRiskLevel:     externalRisk,
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

package syncer

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"vector/internal/db"
	"vector/internal/external"
	"vector/internal/models"
)

type FullSyncStats struct {
	Pages     int
	Saved     int
	Applied   int
	Created   int
	Updated   int
	Unchanged int
}

func FullSync(ctx context.Context, gdb *gorm.DB, client *external.Client, perPage int, pageTimeout time.Duration) (FullSyncStats, error) {
	if perPage <= 0 {
		perPage = 100
	}
	stats := FullSyncStats{}

	// первая страница — чтобы узнать total_pages
	firstCtx, cancel := context.WithTimeout(ctx, pageTimeout)
	defer cancel()
	first, err := client.GetUsersRaw(firstCtx, 1, perPage)
	if err != nil {
		return stats, err
	}
	totalPages := first.TotalPages
	if totalPages <= 0 {
		totalPages = 1
	}

	for page := 1; page <= totalPages; page++ {
		pctx, pcancel := context.WithTimeout(ctx, pageTimeout)
		resp, err := client.GetUsersRaw(pctx, page, perPage)
		pcancel()
		if err != nil {
			return stats, err
		}

		now := time.Now().UTC()
		batch := make([]models.StagingExternalUser, 0, len(resp.Users))
		rawUsers := make([]json.RawMessage, 0, len(resp.Users))

		for _, u := range resp.Users {
			var tmp map[string]any
			if err := json.Unmarshal(u, &tmp); err != nil {
				continue
			}
			idVal, ok := tmp["id"]
			if !ok {
				continue
			}

			// поддержим возможные типы числа
			var id int
			switch t := idVal.(type) {
			case float64:
				id = int(t)
			case int:
				id = t
			default:
				continue
			}

			batch = append(batch, models.StagingExternalUser{
				ID:       id,
				Raw:      datatypes.JSON(u),
				SyncedAt: now,
			})
			rawUsers = append(rawUsers, u)
		}

		if err := db.UpsertStagingExternalUsers(gdb, batch); err != nil {
			return stats, err
		}
		st, err := ApplyUsersBatch(gdb, rawUsers)
		if err != nil {
			return stats, err
		}
		stats.Applied += len(rawUsers)
		stats.Created += st.Created
		stats.Updated += st.Updated
		stats.Unchanged += st.Unchanged

		stats.Pages++
		stats.Saved += len(batch)
		stats.Applied += len(batch)
	}

	return stats, nil
}

package cron

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"

	"vector/internal/db"
	"vector/internal/external"
	"vector/internal/repository"
	"vector/internal/service"
)

func StartCron() (*cron.Cron, error) {
	spec := os.Getenv("SYNC_CRON")
	if spec == "" {
		spec = "0 3 * * *" // ежедневно в 03:00
	}

	perPage := 100
	if v := os.Getenv("SYNC_PER_PAGE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			perPage = n
		}
	}

	c := cron.New()

	_, err := c.AddFunc(spec, func() {
		start := time.Now()
		log.Printf("[cron] full sync start (per_page=%d)", perPage)

		// Инициализируем сервисы
		gdb, err := db.Connect()
		if err != nil {
			log.Printf("[cron] db connect error: %v", err)
			return
		}

		stagingRepo := repository.NewStagingRepository(gdb)
		clientRepo := repository.NewClientRepository(gdb)
		externalClient := external.NewClient()
		externalAPI := repository.NewExternalAPIClient(externalClient)
		applyService := service.NewApplyService(stagingRepo, clientRepo, externalAPI, gdb)
		fullSyncService := service.NewFullSyncService(applyService, externalAPI)

		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()

		resp, err := fullSyncService.SyncFull(ctx, service.FullSyncRequest{
			PerPage: perPage,
		})
		if err != nil {
			log.Printf("[cron] full sync error: %v", err)
			return
		}

		log.Printf("[cron] full sync done in %s pages=%d saved=%d applied=%d created=%d updated=%d unchanged=%d",
			time.Since(start), resp.Pages, resp.Saved, resp.Applied, resp.Created, resp.Updated, resp.Unchanged)
	})
	if err != nil {
		return nil, err
	}

	c.Start()
	log.Printf("[cron] started with spec=%q", spec)
	return c, nil
}

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

		gdb, err := db.Connect()
		if err != nil {
			log.Printf("[cron] db connect error: %v", err)
			return
		}

		stagingRepo := repository.NewSyncStagingRepository(gdb)
		clientRepo := repository.NewSyncClientRepository(gdb)

		contractStagingRepo := repository.NewContractStagingRepository(gdb)
		contractRepo := repository.NewSyncContractRepository(gdb)

		externalClient := external.NewClient()
		externalAPI := repository.NewSyncExternalAPIClient(externalClient)

		triggerService := service.NewTriggerService()
		applyService := service.NewApplyService(stagingRepo, clientRepo, externalAPI, triggerService)

		contractService := service.NewContractService(contractStagingRepo, contractRepo, externalAPI)

		fullSyncService := service.NewFullSyncService(applyService, contractService, externalAPI)

		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()

		resp, err := fullSyncService.SyncFull(ctx, service.FullSyncRequest{
			PerPage:       perPage,
			SyncContracts: true,
		})
		if err != nil {
			log.Printf("[cron] full sync error: %v", err)
			return
		}

		log.Printf("[cron] full sync done in %s users: pages=%d applied=%d created=%d updated=%d contracts: pages=%d applied=%d created=%d updated=%d",
			time.Since(start),
			resp.UserPages, resp.UserApplied, resp.UserCreated, resp.UserUpdated,
			resp.ContractPages, resp.ContractApplied, resp.ContractCreated, resp.ContractUpdated)
	})
	if err != nil {
		return nil, err
	}

	c.Start()
	log.Printf("[cron] started with spec=%q", spec)
	return c, nil
}

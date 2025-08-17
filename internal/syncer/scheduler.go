package syncer

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"

	"vector/internal/db"
	"vector/internal/external"
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

		client := external.NewClient()
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()

		stats, err := FullSync(ctx, gdb, client, perPage, 30*time.Second)
		if err != nil {
			log.Printf("[cron] full sync error: %v", err)
			return
		}
		log.Printf("[cron] full sync done in %s pages=%d saved=%d applied=%d created=%d updated=%d unchanged=%d",
			time.Since(start), stats.Pages, stats.Saved, stats.Applied, stats.Created, stats.Updated, stats.Unchanged)
	})
	if err != nil {
		return nil, err
	}

	c.Start()
	log.Printf("[cron] started with spec=%q", spec)
	return c, nil
}

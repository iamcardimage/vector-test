package appsvc

import (
	"context"
	"log"
	"os"
	"time"

	"vector/internal/db"
	"vector/internal/repository"

	"github.com/robfig/cron/v3"
)

func StartCron() (*cron.Cron, error) {
	spec := os.Getenv("APP_CRON")
	if spec == "" {
		spec = "0 4 * * *"
	}
	c := cron.New()
	_, err := c.AddFunc(spec, func() {
		log.Printf("[app-cron] recalc start")
		gdb, err := db.Connect()
		if err != nil {
			log.Printf("[app-cron] db connect error: %v", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		_ = ctx

		recalcRepo := repository.NewRecalcRepository(gdb)

		n, err := recalcRepo.RecalcNeedsSecondPart()
		if err != nil {
			log.Printf("[app-cron] recalc error: %v", err)
		} else {
			log.Printf("[app-cron] recalc done, updated=%d", n)
		}

		n2, err2 := recalcRepo.RecalcPassportExpiry()
		if err2 != nil {
			log.Printf("[app-cron] passport recalc error: %v", err2)
		} else {
			log.Printf("[app-cron] passport recalc updated=%d", n2)
		}
	})
	if err != nil {
		return nil, err
	}
	c.Start()
	log.Printf("[app-cron] started spec=%q", spec)
	return c, nil
}

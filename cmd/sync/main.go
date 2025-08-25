package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"vector/internal/cron"
	"vector/internal/db"
	"vector/internal/external"
	"vector/internal/repository"
	"vector/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()

	_, _ = cron.StartCron()

	gdb, err := db.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	stagingRepo := repository.NewStagingRepository(gdb)
	clientRepo := repository.NewClientRepository(gdb)
	externalClient := external.NewClient()
	externalApi := repository.NewExternalAPIClient(externalClient)
	stagingService := service.NewStagingService(stagingRepo, externalApi)
	applyService := service.NewApplyService(stagingRepo, clientRepo, externalApi, gdb)
	fullSyncService := service.NewFullSyncService(applyService, externalApi)

	app := fiber.New()

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Get("/dbping", func(c *fiber.Ctx) error {
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).SendString("db connect error: " + err.Error())
		}
		sqlDB, err := gdb.DB()
		if err != nil {
			return c.Status(500).SendString("db handle error: " + err.Error())
		}
		if err := sqlDB.Ping(); err != nil {
			return c.Status(500).SendString("db ping error: " + err.Error())
		}
		return c.SendString("db ok")
	})

	app.Post("/sync/staging", func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		if page <= 0 {
			page = 1
		}
		perPage, _ := strconv.Atoi(c.Query("per_page", "100"))
		if perPage <= 0 {
			perPage = 100
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := stagingService.SyncStaging(ctx, service.SyncStagingRequest{
			Page:    page,
			PerPage: perPage,
		})
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(resp)

	})

	app.Post("/sync/apply", func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		if page <= 0 {
			page = 1
		}
		perPage, _ := strconv.Atoi(c.Query("per_page", "100"))
		if perPage <= 0 {
			perPage = 100
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := applyService.SyncApply(ctx, service.SyncApplyRequest{
			Page:    page,
			PerPage: perPage,
		})
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(resp)
	})

	app.Post("/sync/full", func(c *fiber.Ctx) error {
		perPage, _ := strconv.Atoi(c.Query("per_page", "100"))

		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()

		resp, err := fullSyncService.SyncFull(ctx, service.FullSyncRequest{
			PerPage: perPage,
		})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(resp)
	})

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}

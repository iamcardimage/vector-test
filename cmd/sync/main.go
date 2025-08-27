// @title Vector Sync API
// @version 1.0
// @description API for sync endpoints
// @BasePath /
package main

import (
	"log"

	"vector/internal/cron"
	"vector/internal/db"
	"vector/internal/external"
	"vector/internal/handlers"
	"vector/internal/repository"
	"vector/internal/routes"
	"vector/internal/service"

	"github.com/gofiber/contrib/swagger"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	_, _ = cron.StartCron()

	gdb, err := db.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	deps := initDependencies(gdb)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Unhandled error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal server error",
				"success": false,
			})
		},
	})

	// Swagger UI (Sync)
	cfg := swagger.Config{
		BasePath: "/",
		FilePath: "./docs/sync/swagger.json",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}

	app.Use(swagger.New(cfg))

	routes.SetupSyncRoutes(app, deps.syncHandlers, deps.healthHandlers)

	if err := app.Listen(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

type dependencies struct {
	syncHandlers   *handlers.SyncHandlers
	healthHandlers *handlers.HealthHandlers
}

func initDependencies(gdb *gorm.DB) *dependencies {

	stagingRepo := repository.NewSyncStagingRepository(gdb)
	clientRepo := repository.NewSyncClientRepository(gdb)

	contractStagingRepo := repository.NewContractStagingRepository(gdb)
	contractRepo := repository.NewSyncContractRepository(gdb)

	externalClient := external.NewClient()
	externalAPI := repository.NewSyncExternalAPIClient(externalClient)

	triggerService := service.NewTriggerService()
	stagingService := service.NewStagingService(stagingRepo, externalAPI)
	applyService := service.NewApplyService(stagingRepo, clientRepo, externalAPI, triggerService)

	contractService := service.NewContractService(contractStagingRepo, contractRepo, externalAPI)

	fullSyncService := service.NewFullSyncService(applyService, contractService, externalAPI)

	syncHandlers := handlers.NewSyncHandlers(stagingService, applyService, fullSyncService)
	healthHandlers := handlers.NewHealthHandlers()

	return &dependencies{
		syncHandlers:   syncHandlers,
		healthHandlers: healthHandlers,
	}
}

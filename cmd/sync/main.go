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

	routes.SetupSyncRoutes(app, deps.syncHandlers, deps.healthHandlers)

	// Start server
	if err := app.Listen(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

type dependencies struct {
	syncHandlers   *handlers.SyncHandlers
	healthHandlers *handlers.HealthHandlers
}

func initDependencies(gdb *gorm.DB) *dependencies {
	// Repositories
	stagingRepo := repository.NewSyncStagingRepository(gdb)
	clientRepo := repository.NewSyncClientRepository(gdb)

	contractStagingRepo := repository.NewContractStagingRepository(gdb)
	contractRepo := repository.NewSyncContractRepository(gdb)

	// External client
	externalClient := external.NewClient()
	externalAPI := repository.NewSyncExternalAPIClient(externalClient)

	// Services
	triggerService := service.NewTriggerService()
	stagingService := service.NewStagingService(stagingRepo, externalAPI)
	applyService := service.NewApplyService(stagingRepo, clientRepo, externalAPI, triggerService)

	// Contract Service
	contractService := service.NewContractService(contractStagingRepo, contractRepo, externalAPI)

	// Full Sync Service с Contract Service
	fullSyncService := service.NewFullSyncService(applyService, contractService, externalAPI)

	// Handlers
	syncHandlers := handlers.NewSyncHandlers(stagingService, applyService, fullSyncService)
	healthHandlers := handlers.NewHealthHandlers()

	return &dependencies{
		syncHandlers:   syncHandlers,
		healthHandlers: healthHandlers,
	}
}

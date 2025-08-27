package main

import (
	"log"
	"os"

	"vector/internal/db"
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

	routes.SetupAppRoutes(app, deps.appHandlers, deps.healthHandlers, nil)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}

type dependencies struct {
	appHandlers    *handlers.AppHandlers
	healthHandlers *handlers.HealthHandlers
}

func initDependencies(gdb *gorm.DB) *dependencies {

	clientRepo := repository.NewAppClientRepository(gdb)
	userRepo := repository.NewUserRepository(gdb)
	checkRepo := repository.NewCheckRepository(gdb)
	recalcRepo := repository.NewRecalcRepository(gdb)

	appService := service.NewAppService(clientRepo, userRepo, checkRepo, recalcRepo)

	appHandlers := handlers.NewAppHandlers(appService)
	healthHandlers := handlers.NewHealthHandlers()

	return &dependencies{
		appHandlers:    appHandlers,
		healthHandlers: healthHandlers,
	}
}

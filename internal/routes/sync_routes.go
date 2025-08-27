package routes

import (
	"time"
	"vector/internal/handlers"
	"vector/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupSyncRoutes(app *fiber.App, syncHandlers *handlers.SyncHandlers, healthHandlers *handlers.HealthHandlers) {

	app.Use(middleware.ErrorHandler())

	app.Get("/healthz", healthHandlers.Health)
	app.Get("/dbping", healthHandlers.DBPing)

	syncGroup := app.Group("/sync")
	syncGroup.Use(
		middleware.RequestTimeout(30*time.Second),
		middleware.ValidatePagination(),
	)

	syncGroup.Post("/staging", syncHandlers.SyncStaging)
	syncGroup.Post("/apply", syncHandlers.SyncApply)

	app.Post("/sync/full",
		middleware.RequestTimeout(time.Hour),
		middleware.ValidatePagination(),
		syncHandlers.SyncFull,
	)
}

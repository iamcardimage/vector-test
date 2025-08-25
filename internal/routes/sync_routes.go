// internal/routes/sync_routes.go
package routes

import (
	"time"
	"vector/internal/handlers"
	"vector/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupSyncRoutes(app *fiber.App, syncHandlers *handlers.SyncHandlers, healthHandlers *handlers.HealthHandlers) {
	// Middleware для всех routes
	app.Use(middleware.ErrorHandler())

	// Health endpoints (без дополнительных middleware)
	app.Get("/healthz", healthHandlers.Health)
	app.Get("/dbping", healthHandlers.DBPing)

	// Sync endpoints group с middleware
	syncGroup := app.Group("/sync")
	syncGroup.Use(
		middleware.RequestTimeout(30*time.Second),
		middleware.ValidatePagination(),
	)

	syncGroup.Post("/staging", syncHandlers.SyncStaging)
	syncGroup.Post("/apply", syncHandlers.SyncApply)

	// Full sync с увеличенным timeout
	app.Post("/sync/full",
		middleware.RequestTimeout(time.Hour),
		middleware.ValidatePagination(),
		syncHandlers.SyncFull,
	)
}

package routes

import (
	"vector/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupAppRoutes настраивает основные роуты приложения
func SetupAppRoutes(
	app *fiber.App,
	healthHandlers *handlers.HealthHandlers,
) {
	// Health check роуты (публичные)
	app.Get("/healthz", healthHandlers.Health)
	app.Get("/dbping", healthHandlers.DBPing)
}

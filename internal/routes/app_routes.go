package routes

import (
	"vector/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupAppRoutes(
	app *fiber.App,
	appHandlers *handlers.AppHandlers,
	healthHandlers *handlers.HealthHandlers,
	authMiddleware fiber.Handler,
) {

	app.Get("/healthz", healthHandlers.Health)
	app.Get("/dbping", healthHandlers.DBPing)

	app.Get("/clients/:id", appHandlers.GetClient)
	app.Get("/clients/:id/second-part/history", appHandlers.GetSecondPartHistory)
	app.Post("/clients/:id/second-part/draft", appHandlers.CreateSecondPartDraft)

	app.Get("/contracts", appHandlers.ListContracts)
	app.Get("/contracts/:id", appHandlers.GetContract)

	app.Post("/auth/register", appHandlers.CreateUser)
	app.Get("/auth/users", appHandlers.ListUsers)
	app.Patch("/auth/users/:id/role", appHandlers.UpdateUserRole)
	app.Post("/auth/users/:id/rotate-token", appHandlers.RotateUserToken)
	app.Delete("/auth/users/:id", appHandlers.DeleteUser)
}

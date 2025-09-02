package routes

import (
	"vector/internal/handlers"
	"vector/internal/middleware"
	"vector/internal/service"

	"github.com/gofiber/fiber/v2"
)

// SetupAuthRoutes настраивает роуты для JWT аутентификации
func SetupAuthRoutes(
	app *fiber.App,
	authHandlers *handlers.AuthHandlers,
	authService *service.AuthService,
) {
	// JWT middleware
	jwtMiddleware := middleware.JWTMiddleware(authService)

	// Публичные роуты (без аутентификации)
	authGroup := app.Group("/auth")
	{
		// Вход в систему
		authGroup.Post("/login", authHandlers.Login)

		// Получение списка ролей (для UI)
		authGroup.Get("/roles", authHandlers.GetRoles)
	}

	// Защищенные роуты (требуют аутентификации)
	protectedAuth := authGroup.Group("", jwtMiddleware)
	{
		// Профиль текущего пользователя
		protectedAuth.Get("/profile", authHandlers.GetProfile)
	}

	// Роуты только для администраторов
	adminAuth := authGroup.Group("", jwtMiddleware, middleware.RequireAdmin())
	{
		// Управление пользователями
		// adminAuth.Get("/users", authHandlers.ListUsers)
		adminAuth.Post("/users", authHandlers.CreateUser)
		// adminAuth.Patch("/users/:id/role", authHandlers.UpdateUserRole)
		// adminAuth.Post("/users/:id/deactivate", authHandlers.DeactivateUser)
	}
}

// SetupProtectedRoutes настраивает защищенные роуты с проверкой ролей
func SetupProtectedRoutes(
	app *fiber.App,
	appHandlers *handlers.AppHandlers,
	authService *service.AuthService,
) {
	// JWT middleware
	jwtMiddleware := middleware.JWTMiddleware(authService)

	// Клиенты - доступ для всех аутентифицированных пользователей
	clientsGroup := app.Group("/clients", jwtMiddleware, middleware.RequireAnyRole())
	{
		clientsGroup.Get("/", appHandlers.ListClients)
		clientsGroup.Get("/:id", appHandlers.GetClient)
		clientsGroup.Get("/:id/history", appHandlers.GetClientHistory)
		clientsGroup.Get("/:id/history/:version", appHandlers.GetClientVersion)
		clientsGroup.Get("/:id/second-part/current", appHandlers.GetSecondPartCurrent)
		clientsGroup.Get("/:id/second-part/history", appHandlers.GetSecondPartHistory)
	}

	// Операции с Second Part - только для администраторов и отдела ПОДФТ
	secondPartGroup := clientsGroup.Group("/:id/second-part", middleware.RequireAdminOrPodft())
	{
		secondPartGroup.Post("/draft", appHandlers.CreateSecondPartDraft)
	}

	// Контракты - доступ для всех аутентифицированных пользователей
	contractsGroup := app.Group("/contracts", jwtMiddleware, middleware.RequireAnyRole())
	{
		contractsGroup.Get("/", appHandlers.ListContracts)
		contractsGroup.Get("/:id", appHandlers.GetContract)
	}
}

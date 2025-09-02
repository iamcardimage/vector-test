// @title Vector App API
// @version 1.0
// @description API for app endpoints with JWT authentication
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"log"
	"os"

	"vector/internal/config"
	"vector/internal/db"
	"vector/internal/handlers"
	"vector/internal/repository"
	"vector/internal/routes"
	"vector/internal/service"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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

	// Middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // В продакшене указать конкретные домены
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
	}))

	// Swagger documentation
	cfg := swagger.Config{
		BasePath: "/",
		FilePath: "./docs/app/swagger.json",
		Path:     "swagger",
		Title:    "Vector API Documentation",
	}
	app.Use(swagger.New(cfg))

	// Настройка роутов
	setupRoutes(app, deps)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📚 API Documentation: http://localhost:%s/swagger", port)
	log.Printf("🔐 JWT Authentication enabled")

	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}

type dependencies struct {
	appHandlers    *handlers.AppHandlers
	authHandlers   *handlers.AuthHandlers
	healthHandlers *handlers.HealthHandlers
	authService    *service.AuthService
}

func initDependencies(gdb *gorm.DB) *dependencies {
	// Repositories
	clientRepo := repository.NewAppClientRepository(gdb)
	userRepo := repository.NewUserRepository(gdb)
	checkRepo := repository.NewCheckRepository(gdb)
	recalcRepo := repository.NewRecalcRepository(gdb)
	syncContractRepo := repository.NewSyncContractRepository(gdb)

	// JWT Configuration
	jwtConfig := config.GetJWTConfig()

	// Services
	appService := service.NewAppService(clientRepo, userRepo, checkRepo, recalcRepo, syncContractRepo)
	authService := service.NewAuthService(userRepo, jwtConfig)

	// Handlers
	appHandlers := handlers.NewAppHandlers(appService)
	authHandlers := handlers.NewAuthHandlers(authService)
	healthHandlers := handlers.NewHealthHandlers()

	if err := userRepo.Seed(); err != nil {
		log.Printf("Warning: Failed to seed admin user: %v", err)
	} else {
		log.Println("✅ Default admin user seeded (admin@vector.com / admin123)")
	}

	return &dependencies{
		appHandlers:    appHandlers,
		authHandlers:   authHandlers,
		healthHandlers: healthHandlers,
		authService:    authService,
	}
}

func setupRoutes(app *fiber.App, deps *dependencies) {
	// Основная информация об API
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":        "Vector API",
			"version":     "1.0.0",
			"description": "API для работы с клиентами и контрактами",
			"auth":        "JWT Bearer Token",
			"docs":        "/swagger",
			"endpoints": fiber.Map{
				"auth": fiber.Map{
					"login":   "POST /auth/login",
					"profile": "GET /auth/profile",
					"users":   "GET /auth/users (admin only)",
				},
				"clients": fiber.Map{
					"list": "GET /clients",
					"get":  "GET /clients/:id",
				},
				"contracts": fiber.Map{
					"list": "GET /contracts",
					"get":  "GET /contracts/:id",
				},
			},
		})
	})

	// Health check роуты (публичные)
	routes.SetupAppRoutes(app, deps.healthHandlers)

	// JWT Authentication роуты
	routes.SetupAuthRoutes(app, deps.authHandlers, deps.authService)

	// Защищенные роуты с проверкой ролей
	routes.SetupProtectedRoutes(app, deps.appHandlers, deps.authService)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"error":   "Endpoint not found",
			"success": false,
			"path":    c.Path(),
			"method":  c.Method(),
		})
	})
}

package middleware

import (
	"strings"
	"vector/internal/models"
	"vector/internal/service"

	"github.com/gofiber/fiber/v2"
)

// JWTMiddleware middleware для проверки JWT токенов
func JWTMiddleware(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{
				"error":   "missing bearer token",
				"success": false,
			})
		}

		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))

		user, err := authService.ValidateToken(token)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error":   err.Error(),
				"success": false,
			})
		}

		// Сохраняем пользователя в контексте запроса
		c.Locals("user", *user)
		return c.Next()
	}
}

// RequireRole middleware для проверки роли пользователя
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(models.AppUser)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error":   "пользователь не аутентифицирован",
				"success": false,
			})
		}

		// Проверяем, есть ли роль пользователя среди разрешенных
		for _, allowedRole := range allowedRoles {
			if user.Role == allowedRole {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"error":   "недостаточно прав доступа",
			"success": false,
		})
	}
}

// RequireAdmin middleware для проверки прав администратора
func RequireAdmin() fiber.Handler {
	return RequireRole(models.RoleAdministrator)
}

// RequireAdminOrPodft middleware для администратора или отдела ПОДФТ
func RequireAdminOrPodft() fiber.Handler {
	return RequireRole(models.RoleAdministrator, models.RolePodft)
}

// RequireAnyRole middleware для любой роли (только аутентифицированные пользователи)
func RequireAnyRole() fiber.Handler {
	return RequireRole(models.RoleAdministrator, models.RolePodft, models.RoleClientManagement)
}

// GetCurrentUser helper функция для получения текущего пользователя из контекста
func GetCurrentUser(c *fiber.Ctx) (*models.AppUser, error) {
	user, ok := c.Locals("user").(models.AppUser)
	if !ok {
		return nil, fiber.NewError(401, "пользователь не аутентифицирован")
	}
	return &user, nil
}

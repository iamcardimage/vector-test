package middleware

import (
	"strings"
	"vector/internal/service"

	"github.com/gofiber/fiber/v2"
)

func NewAuthMiddleware(appService *service.AppService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{"error": "missing bearer token"})
		}

		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))

		u, err := appService.GetUserByToken(token)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
		}

		c.Locals("user", u)
		return c.Next()
	}
}

package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestTimeout(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), timeout)
		defer cancel()

		c.SetUserContext(ctx)
		return c.Next()
	}
}

func ValidatePagination() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		if page <= 0 {
			page = 1
		}

		perPage, _ := strconv.Atoi(c.Query("per_page", "100"))
		if perPage <= 0 || perPage > 1000 {
			perPage = 100
		}

		c.Locals("page", page)
		c.Locals("per_page", perPage)

		return c.Next()
	}
}

func ErrorHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {

			// log.Printf("Request error: %v", err)

			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"error":   err.Error(),
				"success": false,
			})
		}
		return nil
	}
}

package handlers

import (
	"vector/internal/db"

	"github.com/gofiber/fiber/v2"
)

type HealthHandlers struct{}

func NewHealthHandlers() *HealthHandlers {
	return &HealthHandlers{}
}

// Health godoc
// @Summary Health check
// @Tags health
// @Success 200 {string} string "ok"
// @Router /healthz [get]
func (h *HealthHandlers) Health(c *fiber.Ctx) error {
	return c.SendString("ok")
}

// DBPing godoc
// @Summary Database ping
// @Tags health
// @Success 200 {string} string "db ok"
// @Failure 500 {string} string
// @Router /dbping [get]
func (h *HealthHandlers) DBPing(c *fiber.Ctx) error {
	gdb, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db connect error: "+err.Error())
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db handle error: "+err.Error())
	}

	if err := sqlDB.Ping(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db ping error: "+err.Error())
	}

	return c.SendString("db ok")
}

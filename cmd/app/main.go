package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"vector/internal/db"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/datatypes"
)

func main() {
	_ = godotenv.Load()

	app := fiber.New()

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Get("/dbping", func(c *fiber.Ctx) error {
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).SendString("db connect error: " + err.Error())
		}
		sqlDB, err := gdb.DB()
		if err != nil {
			return c.Status(500).SendString("db handle error: " + err.Error())
		}
		if err := sqlDB.Ping(); err != nil {
			return c.Status(500).SendString("db ping error: " + err.Error())
		}
		return c.SendString("db ok")
	})

	app.Post("/migrate/app/second-part", func(c *fiber.Ctx) error {
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).SendString("db connect error: " + err.Error())
		}
		if err := db.MigrateCoreSecondPart(gdb); err != nil {
			return c.Status(500).SendString("migrate second_part error: " + err.Error())
		}
		return c.SendString("migrated second_part")
	})

	app.Get("/clients/:id", func(c *fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		cur, err := db.GetClientCurrent(gdb, id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "client not found"})
		}
		var sp fiber.Map
		if curSP, err := db.GetSecondPartCurrent(gdb, id); err == nil {
			sp = fiber.Map{
				"client_version": curSP.ClientVersion,
				"version":        curSP.Version,
				"status":         curSP.Status,
				"risk_level":     curSP.RiskLevel,
				"due_at":         curSP.DueAt,
				"is_current":     curSP.IsCurrent,
			}
			if curSP.ClientVersion != cur.Version {
				sp["stale"] = true
			}
		}
		return c.JSON(fiber.Map{
			"id":                  cur.ClientID,
			"version":             cur.Version,
			"surname":             cur.Surname,
			"name":                cur.Name,
			"patronymic":          cur.Patronymic,
			"needs_second_part":   cur.NeedsSecondPart,
			"second_part_created": cur.SecondPartCreated,
			"second_part":         sp,
		})
	})

	app.Get("/clients/:id/second-part/history", func(c *fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		vs, err := db.ListSecondPartHistory(gdb, id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "history: " + err.Error()})
		}
		out := make([]fiber.Map, 0, len(vs))
		for _, v := range vs {
			out = append(out, fiber.Map{
				"client_version": v.ClientVersion,
				"version":        v.Version,
				"status":         v.Status,
				"risk_level":     v.RiskLevel,
				"due_at":         v.DueAt,
				"valid_from":     v.ValidFrom,
				"valid_to":       v.ValidTo,
				"is_current":     v.IsCurrent,
			})
		}
		return c.JSON(fiber.Map{"success": true, "versions": out})
	})

	app.Post("/clients/:id/second-part", func(c *fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))

		// простая схема ввода: { "risk_level":"low|high", "data": { ... } }
		var in struct {
			RiskLevel string         `json:"risk_level"`
			Data      map[string]any `json:"data"`
			CreatedBy *int           `json:"created_by_user_id"`
		}
		if err := c.BodyParser(&in); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
		}

		var dataOverride *datatypes.JSON
		if in.Data != nil {
			b, _ := json.Marshal(in.Data)
			dj := datatypes.JSON(b)
			dataOverride = &dj
		}

		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}

		sp, err := db.CreateSecondPartDraft(gdb, id, &in.RiskLevel, in.CreatedBy, dataOverride)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "create draft: " + err.Error()})
		}

		return c.JSON(fiber.Map{
			"success":        true,
			"client_id":      sp.ClientID,
			"client_version": sp.ClientVersion,
			"version":        sp.Version,
			"status":         sp.Status,
			"risk_level":     sp.RiskLevel,
			"due_at":         sp.DueAt,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}

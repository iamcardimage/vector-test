package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"vector/internal/db"
	"vector/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/datatypes"
)

func asIntPtr(u uint) *int {
	i := int(u)
	return &i
}

func main() {
	_ = godotenv.Load()

	app := fiber.New()

	// Auth middleware (Bearer <token>)
	authmw := func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{"error": "missing bearer token"})
		}
		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		u, err := db.GetUserByToken(gdb, token)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
		}
		c.Locals("user", u)
		return c.Next()
	}

	// Публичные маршруты (health, миграции)
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
	app.Post("/migrate/app/users", func(c *fiber.Ctx) error {
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).SendString("db connect error: " + err.Error())
		}
		if err := db.MigrateCoreUsers(gdb); err != nil {
			return c.Status(500).SendString("migrate users error: " + err.Error())
		}
		return c.SendString("migrated users")
	})
	app.Post("/migrate/app/seed-users", func(c *fiber.Ctx) error {
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).SendString("db connect error: " + err.Error())
		}
		if err := db.SeedAppUsers(gdb); err != nil {
			return c.Status(500).SendString("seed error: " + err.Error())
		}
		return c.SendString("seeded users")
	})

	// Защищённая группа (Bearer)
	api := app.Group("/", authmw)

	// GET текущий клиент + текущая 2-я часть
	api.Get("/clients/:id", func(c *fiber.Ctx) error {
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

	// GET история 2-й части
	api.Get("/clients/:id/second-part/history", func(c *fiber.Ctx) error {
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

	// Helper: запретить мутации для viewer
	canMutate := func(c *fiber.Ctx) (models.AppUser, bool) {
		u := c.Locals("user").(models.AppUser)
		if u.Role == "viewer" {
			c.Status(403).JSON(fiber.Map{"error": "forbidden"})
			return u, false
		}
		return u, true
	}

	// POST создать draft 2-й части (prefill + опциональные data/risk)
	api.Post("/clients/:id/second-part", func(c *fiber.Ctx) error {
		u, ok := canMutate(c)
		if !ok {
			return nil
		}
		id, _ := strconv.Atoi(c.Params("id"))
		var in struct {
			RiskLevel string         `json:"risk_level"` // low|high
			Data      map[string]any `json:"data"`
		}
		if err := c.BodyParser(&in); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
		}
		var dataOverride *datatypes.JSON
		if in.Data != nil {
			b, _ := json.Marshal(in.Data)
			d := datatypes.JSON(b)
			dataOverride = &d
		}
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		sp, err := db.CreateSecondPartDraft(gdb, id, &in.RiskLevel, asIntPtr(u.ID), dataOverride)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "create draft: " + err.Error()})
		}
		return c.JSON(sp)
	})

	// POST submit 2-й части
	api.Post("/clients/:id/second-part/submit", func(c *fiber.Ctx) error {
		u, ok := canMutate(c)
		if !ok {
			return nil
		}
		id, _ := strconv.Atoi(c.Params("id"))
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		sp, err := db.SubmitSecondPart(gdb, id, asIntPtr(u.ID))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "submit: " + err.Error()})
		}
		return c.JSON(sp)
	})

	// POST approve 2-й части
	api.Post("/clients/:id/second-part/approve", func(c *fiber.Ctx) error {
		u, ok := canMutate(c)
		if !ok {
			return nil
		}
		id, _ := strconv.Atoi(c.Params("id"))
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		sp, err := db.ApproveSecondPart(gdb, id, asIntPtr(u.ID))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "approve: " + err.Error()})
		}
		return c.JSON(sp)
	})

	// POST reject 2-й части
	api.Post("/clients/:id/second-part/reject", func(c *fiber.Ctx) error {
		u, ok := canMutate(c)
		if !ok {
			return nil
		}
		id, _ := strconv.Atoi(c.Params("id"))
		var in struct {
			Reason string `json:"reason"`
		}
		if err := c.BodyParser(&in); err != nil || strings.TrimSpace(in.Reason) == "" {
			return c.Status(400).JSON(fiber.Map{"error": "reason required"})
		}
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		sp, err := db.RejectSecondPart(gdb, id, asIntPtr(u.ID), in.Reason)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "reject: " + err.Error()})
		}
		return c.JSON(sp)
	})

	// POST запрос документов
	api.Post("/clients/:id/second-part/request-docs", func(c *fiber.Ctx) error {
		u, ok := canMutate(c)
		if !ok {
			return nil
		}
		id, _ := strconv.Atoi(c.Params("id"))
		var in struct {
			Reason string `json:"reason"`
		}
		if err := c.BodyParser(&in); err != nil || strings.TrimSpace(in.Reason) == "" {
			return c.Status(400).JSON(fiber.Map{"error": "reason required"})
		}
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		sp, err := db.RequestDocsSecondPart(gdb, id, asIntPtr(u.ID), in.Reason)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "request-docs: " + err.Error()})
		}
		return c.JSON(sp)
	})

	api.Get("/clients", func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		perPage, _ := strconv.Atoi(c.Query("per_page", "100"))

		var filterPtr *bool
		if v := c.Query("needs_second_part"); v != "" {
			b := v == "true" || v == "1"
			filterPtr = &b
		}

		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}

		items, total, err := db.ListCurrentClients(gdb, page, perPage, filterPtr)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "list: " + err.Error()})
		}

		return c.JSON(fiber.Map{
			"success":      true,
			"total_count":  total,
			"per_page":     perPage,
			"current_page": page,
			"users":        items,
		})
	})

	api.Get("/clients/search", func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		perPage, _ := strconv.Atoi(c.Query("per_page", "100"))

		var needsPtr *bool
		if v := c.Query("needs_second_part"); v != "" {
			b := v == "true" || v == "1"
			needsPtr = &b
		}

		var statusPtr *string
		if v := c.Query("sp_status"); v != "" {
			statusPtr = &v // draft|submitted|approved|rejected|doc_requested
		}

		var duePtr *time.Time
		if v := c.Query("due_before"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				duePtr = &t
			} else {
				return c.Status(400).JSON(fiber.Map{"error": "invalid due_before (RFC3339)"})
			}
		}

		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		items, total, err := db.ListClientsWithSP(gdb, page, perPage, needsPtr, statusPtr, duePtr)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "list: " + err.Error()})
		}

		// пометить устаревшую 2-ю часть, если client_version != sp_client_version
		type Row struct {
			db.ClientWithSP
			Stale *bool `json:"stale,omitempty"`
		}
		out := make([]Row, 0, len(items))
		for _, it := range items {
			row := Row{ClientWithSP: it}
			if it.SpClientVersion != nil {
				stale := *it.SpClientVersion != it.ClientVersion
				if stale {
					row.Stale = &stale
				}
			}
			out = append(out, row)
		}

		return c.JSON(fiber.Map{
			"success":      true,
			"total_count":  total,
			"per_page":     perPage,
			"current_page": page,
			"users":        out,
		})
	})

	api.Post("/recalc/needs-second-part", func(c *fiber.Ctx) error {
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		n, err := db.RecalcNeedsSecondPart(gdb)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "recalc: " + err.Error()})
		}
		return c.JSON(fiber.Map{"success": true, "updated": n})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"vector/internal/db"
	"vector/internal/external"
	"vector/internal/models"
	"vector/internal/syncer"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/datatypes"
)

func main() {

	_ = godotenv.Load()

	_, _ = syncer.StartCron()

	app := fiber.New()

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// DB ping
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

	// External users (id, surname, name, patronymic)
	app.Get("/external/users", func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		if page <= 0 {
			page = 1
		}
		perPage, _ := strconv.Atoi(c.Query("per_page", "100"))
		if perPage <= 0 {
			perPage = 100
		}

		client := external.NewClient()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		resp, err := client.GetUsers(ctx, page, perPage)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"error": err.Error()})
		}

		type userDTO struct {
			ID         int    `json:"id"`
			Surname    string `json:"surname"`
			Name       string `json:"name"`
			Patronymic string `json:"patronymic"`
		}
		users := make([]userDTO, len(resp.Users))
		for i, u := range resp.Users {
			users[i] = userDTO{
				ID:         u.ID,
				Surname:    u.Surname,
				Name:       u.Name,
				Patronymic: u.Patronymic,
			}
		}

		return c.JSON(fiber.Map{
			"success":      true,
			"total_count":  resp.TotalCount,
			"per_page":     resp.PerPage,
			"current_page": resp.CurrentPage,
			"total_pages":  resp.TotalPages,
			"users":        users,
		})
	})
	// Запуск миграции таблицы external_users
	app.Post("/migrate", func(c *fiber.Ctx) error {
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).SendString("db connect error: " + err.Error())
		}
		if err := db.MigrateExternal(gdb); err != nil {
			return c.Status(500).SendString("migrate error: " + err.Error())
		}
		return c.SendString("migrated")
	})

	// Синхронизация внешних пользователей в локальную БД (апсерт)
	app.Post("/external/users/sync", func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		if page <= 0 {
			page = 1
		}
		perPage, _ := strconv.Atoi(c.Query("per_page", "100"))
		if perPage <= 0 {
			perPage = 100
		}

		client := external.NewClient()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := client.GetUsers(ctx, page, perPage)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"error": err.Error()})
		}

		// map external → models
		batch := make([]models.ExternalUser, 0, len(resp.Users))
		for _, u := range resp.Users {
			batch = append(batch, models.ExternalUser{
				ID:         u.ID,
				Surname:    u.Surname,
				Name:       u.Name,
				Patronymic: u.Patronymic,
			})
		}

		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		if err := db.UpsertExternalUsers(gdb, batch); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "upsert: " + err.Error()})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"saved":   len(batch),
			"page":    resp.CurrentPage,
			"total":   resp.TotalCount,
		})
	})

	app.Post("/migrate/staging", func(c *fiber.Ctx) error {
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).SendString("db connect error: " + err.Error())
		}
		if err := db.MigrateStaging(gdb); err != nil {
			return c.Status(500).SendString("migrate staging error: " + err.Error())
		}
		return c.SendString("migrated staging")
	})

	app.Post("/migrate/core", func(c *fiber.Ctx) error {
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).SendString("db connect error: " + err.Error())
		}
		if err := db.MigrateCoreClients(gdb); err != nil {
			return c.Status(500).SendString("migrate core error: " + err.Error())
		}
		return c.SendString("migrated core")
	})

	// Сохранение сырого snapshot в staging из внешнего API
	app.Post("/sync/staging", func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		if page <= 0 {
			page = 1
		}
		perPage, _ := strconv.Atoi(c.Query("per_page", "100"))
		if perPage <= 0 {
			perPage = 100
		}

		client := external.NewClient()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		rawResp, err := client.GetUsersRaw(ctx, page, perPage)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"error": err.Error()})
		}

		now := time.Now().UTC()
		batch := make([]models.StagingExternalUser, 0, len(rawResp.Users))
		for _, r := range rawResp.Users {
			// вытянуть id из raw
			var tmp map[string]any
			if err := json.Unmarshal(r, &tmp); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "decode raw user: " + err.Error()})
			}
			idVal, ok := tmp["id"]
			if !ok {
				continue
			}
			id, ok := idVal.(float64)
			if !ok {
				continue
			}

			batch = append(batch, models.StagingExternalUser{
				ID:       int(id),
				Raw:      datatypes.JSON(r),
				SyncedAt: now,
			})
		}

		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		if err := db.UpsertStagingExternalUsers(gdb, batch); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "staging upsert: " + err.Error()})
		}

		return c.JSON(fiber.Map{
			"success":     true,
			"saved":       len(batch),
			"page":        rawResp.CurrentPage,
			"total_pages": rawResp.TotalPages,
			"total_count": rawResp.TotalCount,
			"per_page":    rawResp.PerPage,
		})
	})

	// Применение синка в core: тянем raw → staging upsert → SCD2 апдейт клиентов
	app.Post("/sync/apply", func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		if page <= 0 {
			page = 1
		}
		perPage, _ := strconv.Atoi(c.Query("per_page", "100"))
		if perPage <= 0 {
			perPage = 100
		}

		client := external.NewClient()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		rawResp, err := client.GetUsersRaw(ctx, page, perPage)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"error": err.Error()})
		}

		now := time.Now().UTC()
		batch := make([]models.StagingExternalUser, 0, len(rawResp.Users))
		for _, r := range rawResp.Users {
			var tmp map[string]any
			if err := json.Unmarshal(r, &tmp); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "decode raw user: " + err.Error()})
			}
			idVal, ok := tmp["id"]
			if !ok {
				continue
			}
			idFloat, ok := idVal.(float64)
			if !ok {
				continue
			}
			batch = append(batch, models.StagingExternalUser{
				ID:       int(idFloat),
				Raw:      datatypes.JSON(r),
				SyncedAt: now,
			})
		}

		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}
		if err := db.UpsertStagingExternalUsers(gdb, batch); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "staging upsert: " + err.Error()})
		}

		st, err := syncer.ApplyUsersBatch(gdb, rawResp.Users)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "apply batch: " + err.Error()})
		}
		return c.JSON(fiber.Map{
			"success":     true,
			"applied":     len(batch),
			"created":     st.Created,
			"updated":     st.Updated,
			"unchanged":   st.Unchanged,
			"page":        rawResp.CurrentPage,
			"total_pages": rawResp.TotalPages,
			"total_count": rawResp.TotalCount,
			"per_page":    rawResp.PerPage,
		})
	})

	app.Get("/clients", func(c *fiber.Ctx) error {
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

	app.Post("/sync/full", func(c *fiber.Ctx) error {
		perPage, _ := strconv.Atoi(c.Query("per_page", "100"))

		client := external.NewClient()
		gdb, err := db.Connect()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db connect: " + err.Error()})
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()

		stats, err := syncer.FullSync(ctx, gdb, client, perPage, 30*time.Second)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "full sync: " + err.Error()})
		}

		return c.JSON(fiber.Map{
			"success":   true,
			"pages":     stats.Pages,
			"saved":     stats.Saved,
			"applied":   stats.Applied,
			"created":   stats.Created,
			"updated":   stats.Updated,
			"unchanged": stats.Unchanged,
		})
	})

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}

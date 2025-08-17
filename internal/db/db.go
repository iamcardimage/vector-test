package db

import (
	"fmt"
	"log"
	"os"
	"time"
	"vector/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func Connect() (*gorm.DB, error) {
	_ = godotenv.Load()

	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	user := getenv("POSTGRES_USER", "app")
	pass := getenv("POSTGRES_PASSWORD", "app")
	name := getenv("POSTGRES_DB", "vector")
	ssl := getenv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		host, port, user, pass, name, ssl,
	)

	lg := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Warn, // тише лог
		IgnoreRecordNotFoundError: true,        // скрыть "record not found"
		Colorful:                  true,
	})

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: lg,
	})
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func MigrateExternal(gdb *gorm.DB) error {
	return gdb.AutoMigrate(&models.ExternalUser{})
}

func UpsertExternalUsers(gdb *gorm.DB, items []models.ExternalUser) error {
	if len(items) == 0 {
		return nil
	}
	return gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"surname", "name", "patronymic", "updated_at"}),
	}).Create(&items).Error
}

func MigrateStaging(gdb *gorm.DB) error {
	if err := gdb.Exec("CREATE SCHEMA IF NOT EXISTS staging").Error; err != nil {
		return err
	}
	return gdb.AutoMigrate(&models.StagingExternalUser{})
}

func MigrateCoreClients(gdb *gorm.DB) error {
	if err := gdb.Exec("CREATE SCHEMA IF NOT EXISTS core").Error; err != nil {
		return err
	}
	if err := gdb.AutoMigrate(&models.ClientVersion{}); err != nil {
		return err
	}
	// Один текущий срез на клиента (частичный уникальный индекс)
	return gdb.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_current
		ON core.clients_versions (client_id)
		WHERE is_current = true
	`).Error
}

func UpsertStagingExternalUsers(gdb *gorm.DB, items []models.StagingExternalUser) error {
	if len(items) == 0 {
		return nil
	}
	return gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"raw", "synced_at"}),
	}).Create(&items).Error
}

type ClientListItem struct {
	ClientID          int    `gorm:"column:client_id" json:"id"`
	Surname           string `json:"surname"`
	Name              string `json:"name"`
	Patronymic        string `json:"patronymic"`
	ExternalRiskLevel string `gorm:"column:external_risk_level" json:"external_risk_level"`
	NeedsSecondPart   bool   `gorm:"column:needs_second_part" json:"needs_second_part"`
	SecondPartCreated bool   `gorm:"column:second_part_created" json:"second_part_created"`
}

func ListCurrentClients(gdb *gorm.DB, page, perPage int, needsSecondPart *bool) (items []ClientListItem, total int64, err error) {
	q := gdb.Model(&models.ClientVersion{}).
		Where("is_current = ?", true)

	if needsSecondPart != nil {
		q = q.Where("needs_second_part = ?", *needsSecondPart)
	}

	if err = q.Count(&total).Error; err != nil {
		return
	}

	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 500 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	err = q.
		Select([]string{
			"client_id",
			"surname",
			"name",
			"patronymic",
			"external_risk_level",
			"needs_second_part",
			"second_part_created",
		}).
		Order("client_id ASC").
		Limit(perPage).
		Offset(offset).
		Scan(&items).Error

	return
}

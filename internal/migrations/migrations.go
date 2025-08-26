package migrations

import (
	"fmt"
	"log"
	"vector/internal/db"
	appdb "vector/internal/db/app"
	"vector/internal/models"

	"gorm.io/gorm"
)

type Migrator struct {
	db *gorm.DB
}

func New() (*Migrator, error) {
	gdb, err := db.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &Migrator{db: gdb}, nil
}

func (m *Migrator) MigrateAll() error {
	log.Println("Running all migrations...")

	if err := m.MigrateStaging(); err != nil {
		return fmt.Errorf("external migration failed: %w", err)
	}

	if err := m.MigrateCoreClients(); err != nil {
		return fmt.Errorf("core clients migration failed: %w", err)
	}

	if err := m.MigrateCoreSecondPart(); err != nil {
		return fmt.Errorf("core second part migration failed: %w", err)
	}

	if err := m.MigrateCoreUsers(); err != nil {
		return fmt.Errorf("core users migration failed: %w", err)
	}

	if err := m.MigrateCoreChecks(); err != nil {
		return fmt.Errorf("core checks migration failed: %w", err)
	}
	log.Println("All migrations completed successfully")
	return nil
}

func (m *Migrator) MigrateStaging() error {
	log.Println("Migrating staging schema and tables...")
	if err := m.db.Exec("CREATE SCHEMA IF NOT EXISTS staging").Error; err != nil {
		return err
	}
	return m.db.AutoMigrate(&models.StagingExternalUser{})
}

func (m *Migrator) MigrateCoreClients() error {
	log.Println("Migrating core clients tables...")
	if err := m.db.Exec("CREATE SCHEMA IF NOT EXISTS core").Error; err != nil {
		return err
	}

	if err := m.db.AutoMigrate(&models.ClientVersion{}); err != nil {
		return err
	}

	// Создаем индексы для новых полей
	queries := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_versions_current
		 ON core.clients_versions (client_id)
		 WHERE is_current = true`,

		`CREATE INDEX IF NOT EXISTS idx_clients_versions_login
		 ON core.clients_versions (login)
		 WHERE login IS NOT NULL AND login != ''`,

		`CREATE INDEX IF NOT EXISTS idx_clients_versions_external_id_str
		 ON core.clients_versions (external_id_str)
		 WHERE external_id_str IS NOT NULL AND external_id_str != ''`,

		`CREATE INDEX IF NOT EXISTS idx_clients_versions_risk_level
		 ON core.clients_versions (risk_level)
		 WHERE risk_level IS NOT NULL AND risk_level != ''`,

		`CREATE INDEX IF NOT EXISTS idx_clients_versions_inn
		 ON core.clients_versions (inn)
		 WHERE inn IS NOT NULL AND inn != ''`,

		`CREATE INDEX IF NOT EXISTS idx_clients_versions_snils
		 ON core.clients_versions (snils)
		 WHERE snils IS NOT NULL AND snils != ''`,
	}

	for _, query := range queries {
		if err := m.db.Exec(query).Error; err != nil {
			log.Printf("Warning: could not create index: %v", err)
		}
	}

	return nil
}

func (m *Migrator) MigrateCoreSecondPart() error {
	log.Println("Migrating core second part tables...")
	if err := m.db.Exec("CREATE SCHEMA IF NOT EXISTS core").Error; err != nil {
		return err
	}
	if err := m.db.AutoMigrate(&models.SecondPartVersion{}); err != nil {
		return err
	}
	if err := m.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_second_part_current
		ON core.second_part_versions (client_id)
		WHERE is_current = true
	`).Error; err != nil {
		return err
	}

	return m.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sp_client_version_current
		ON core.second_part_versions (client_id, client_version)
		WHERE is_current = true
	`).Error
}

func (m *Migrator) MigrateCoreUsers() error {
	log.Println("Migrating core users tables...")
	if err := m.db.Exec("CREATE SCHEMA IF NOT EXISTS core").Error; err != nil {
		return nil
	}
	return m.db.AutoMigrate(&models.AppUser{})
}

func (m *Migrator) MigrateCoreChecks() error {
	log.Println("Migrating core checks table...")
	if err := m.db.Exec("CREATE SCHEMA IF NOT EXISTS core").Error; err != nil {
		return nil
	}
	return m.db.AutoMigrate(&models.SecondPartCheck{})
}

func (m *Migrator) SeedUsers() error {
	log.Println("Seeding default users...")
	return appdb.SeedAppUsers(m.db) // ИСПРАВЛЕНО: используем appdb
}

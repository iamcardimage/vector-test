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
	log.Println("Migrating core client schema and tables...")
	if err := m.db.Exec("CREATE SCHEMA IF NOT EXISTS core").Error; err != nil {
		return err
	}
	if err := m.db.AutoMigrate(&models.ClientVersion{}); err != nil {
		return err
	}

	return m.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_current
		ON core.clients_versions (client_id)
		WHERE is_current = true 
	`).Error
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

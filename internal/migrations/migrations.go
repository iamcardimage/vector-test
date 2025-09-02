package migrations

import (
	"fmt"
	"log"
	"vector/internal/db"
	"vector/internal/models"
	"vector/internal/repository"

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

	if err := m.MigrateCoreContracts(); err != nil {
		return fmt.Errorf("core contracts migration failed: %w", err)
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
		return err
	}

	// Миграция таблицы пользователей с новой структурой (JWT)
	if err := m.db.AutoMigrate(&models.AppUser{}); err != nil {
		return err
	}

	// Добавляем индексы для производительности
	queries := []string{
		`CREATE INDEX IF NOT EXISTS idx_app_users_email
		 ON core.app_users (email)`,

		`CREATE INDEX IF NOT EXISTS idx_app_users_role
		 ON core.app_users (role)`,

		`CREATE INDEX IF NOT EXISTS idx_app_users_active
		 ON core.app_users (is_active)`,
	}

	for _, query := range queries {
		if err := m.db.Exec(query).Error; err != nil {
			log.Printf("Warning: could not create user index: %v", err)
		}
	}

	return nil
}

func (m *Migrator) MigrateCoreChecks() error {
	log.Println("Migrating core checks table...")
	if err := m.db.Exec("CREATE SCHEMA IF NOT EXISTS core").Error; err != nil {
		return err
	}
	return m.db.AutoMigrate(&models.SecondPartCheck{})
}

func (m *Migrator) MigrateCoreContracts() error {
	log.Println("Migrating core contracts tables...")
	if err := m.db.Exec("CREATE SCHEMA IF NOT EXISTS core").Error; err != nil {
		return err
	}

	if err := m.db.AutoMigrate(&models.StagingExternalContract{}); err != nil {
		return err
	}

	if err := m.db.AutoMigrate(&models.Contract{}); err != nil {
		return err
	}

	queries := []string{
		`CREATE INDEX IF NOT EXISTS idx_contracts_user_id
		 ON core.contracts (user_id)`,

		`CREATE INDEX IF NOT EXISTS idx_contracts_status
		 ON core.contracts (status)`,

		`CREATE INDEX IF NOT EXISTS idx_contracts_kind
		 ON core.contracts (kind)`,

		`CREATE INDEX IF NOT EXISTS idx_contracts_inner_code
		 ON core.contracts (inner_code)`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_contracts_external_id
		 ON core.contracts (external_id)`,
	}

	for _, query := range queries {
		if err := m.db.Exec(query).Error; err != nil {
			log.Printf("Warning: could not create contracts index: %v", err)
		}
	}

	return nil
}

func (m *Migrator) SeedUsers() error {
	log.Println("Seeding default users...")

	// Создаем репозиторий пользователей
	userRepo := repository.NewUserRepository(m.db)

	// Используем метод Seed из репозитория
	if err := userRepo.Seed(); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	log.Println("✅ Default users seeded successfully")
	return nil
}

// MigrateUsersToJWT мигрирует существующих пользователей к новой JWT структуре
func (m *Migrator) MigrateUsersToJWT() error {
	log.Println("Migrating existing users to JWT structure...")

	// Шаг 1: Добавляем новые поля к существующей таблице
	if err := m.db.Exec(`
		ALTER TABLE core.app_users 
		ADD COLUMN IF NOT EXISTS first_name VARCHAR(255),
		ADD COLUMN IF NOT EXISTS last_name VARCHAR(255),
		ADD COLUMN IF NOT EXISTS middle_name VARCHAR(255),
		ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255),
		ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true,
		ADD COLUMN IF NOT EXISTS last_login TIMESTAMP
	`).Error; err != nil {
		return fmt.Errorf("failed to add new columns: %w", err)
	}

	// Шаг 2: Обновляем существующих пользователей с временными данными
	if err := m.db.Exec(`
		UPDATE core.app_users 
		SET 
			first_name = COALESCE(first_name, 'Имя'),
			last_name = COALESCE(last_name, 'Фамилия'),
			middle_name = COALESCE(middle_name, ''),
			password_hash = COALESCE(password_hash, '$2a$10$defaulthashedpassword'),
			is_active = COALESCE(is_active, true)
		WHERE first_name IS NULL OR last_name IS NULL OR password_hash IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to update existing users: %w", err)
	}

	// Шаг 3: Удаляем старое поле token (если существует)
	if err := m.db.Exec(`
		ALTER TABLE core.app_users 
		DROP COLUMN IF EXISTS token
	`).Error; err != nil {
		log.Printf("Warning: could not drop token column: %v", err)
	}

	// Шаг 4: Делаем обязательные поля NOT NULL
	if err := m.db.Exec(`
		ALTER TABLE core.app_users 
		ALTER COLUMN first_name SET NOT NULL,
		ALTER COLUMN last_name SET NOT NULL,
		ALTER COLUMN password_hash SET NOT NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to set NOT NULL constraints: %w", err)
	}

	log.Println("✅ User migration to JWT structure completed")
	log.Println("📝 NOTE: All existing users now have default passwords. Please reset them via admin panel.")

	return nil
}

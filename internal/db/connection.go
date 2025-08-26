package db

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	globalDB *gorm.DB
	once     sync.Once
)

// Connect создает подключение к базе данных (singleton)
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

	var openErr error
	once.Do(func() {
		lg := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		})

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: lg})
		if err != nil {
			openErr = err
			return
		}
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.SetMaxOpenConns(15)
			sqlDB.SetMaxIdleConns(5)
			sqlDB.SetConnMaxLifetime(5 * time.Minute)
			sqlDB.SetConnMaxIdleTime(1 * time.Minute)
		}
		globalDB = db
	})
	if openErr != nil {
		return nil, openErr
	}
	return globalDB, nil
}

// getenv возвращает переменную окружения или значение по умолчанию
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

package config

import (
	"os"
	"time"
	"vector/internal/models"
)

// GetJWTConfig возвращает конфигурацию JWT из переменных окружения
func GetJWTConfig() models.JWTConfig {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		// Временный ключ для разработки
		// В продакшене ОБЯЗАТЕЛЬНО должен быть установлен через переменную окружения
		secretKey = "your-very-secret-jwt-key-change-this-in-production"
	}

	// Время жизни токена (по умолчанию 24 часа)
	tokenDuration := 24 * time.Hour
	if durationStr := os.Getenv("JWT_TOKEN_DURATION"); durationStr != "" {
		if d, err := time.ParseDuration(durationStr); err == nil {
			tokenDuration = d
		}
	}

	return models.JWTConfig{
		SecretKey:     secretKey,
		TokenDuration: tokenDuration,
	}
}

package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	Token     string  `json:"token"`
	ExpiresAt int64   `json:"expires_at"`
	User      AppUser `json:"user"`
}

type CreateUserRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=6"`
	FirstName  string `json:"first_name" validate:"required"`
	LastName   string `json:"last_name" validate:"required"`
	MiddleName string `json:"middle_name"`
	Role       string `json:"role" validate:"required"`
}

type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
}

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

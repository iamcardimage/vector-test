package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	RoleAdministrator    = "Administrator"
	RolePodft            = "Podft"
	RoleClientManagement = "ClientManagement"
)

type AppUser struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	FirstName    string `gorm:"not null"` // Имя
	LastName     string `gorm:"not null"` // Фамилия
	MiddleName   string `gorm:""`         // Отчество (может быть пустым)
	Role         string `gorm:"type:text;not null"`
	IsActive     bool   `gorm:"default:true"` // Активен ли пользователь
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (AppUser) TableName() string {
	return "core.app_users"
}

// HashPassword хеширует пароль
func (u *AppUser) HashPassword(password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedBytes)
	return nil
}

// CheckPassword проверяет пароль
func (u *AppUser) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// GetFullName возвращает полное имя пользователя
func (u *AppUser) GetFullName() string {
	if u.MiddleName != "" {
		return u.LastName + " " + u.FirstName + " " + u.MiddleName
	}
	return u.LastName + " " + u.FirstName
}

// PublicUser возвращает пользователя без чувствительных данных
func (u *AppUser) PublicUser() AppUser {
	return AppUser{
		ID:         u.ID,
		Email:      u.Email,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		MiddleName: u.MiddleName,
		Role:       u.Role,
		IsActive:   u.IsActive,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

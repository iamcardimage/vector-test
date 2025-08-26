package app

import (
	"fmt"
	"strings"
	"vector/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ===== USER MANAGEMENT =====

// GetUserByToken находит пользователя по токену
func GetUserByToken(gdb *gorm.DB, token string) (models.AppUser, error) {
	var u models.AppUser
	err := gdb.Where("token = ?", token).Take(&u).Error
	return u, err
}

// SeedAppUsers создает начальных пользователей
func SeedAppUsers(gdb *gorm.DB) error {
	users := []models.AppUser{
		{Email: "admin@vector.com", Role: models.RoleAdministrator, Token: "admin-token"},
		{Email: "podft@vector.com", Role: models.RolePodft, Token: "podft-token"},
		{Email: "cm@vector.com", Role: models.RoleClientManagement, Token: "cm-token"},
	}
	for _, u := range users {
		var existing models.AppUser
		if err := gdb.Where("email = ?", u.Email).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := gdb.Create(&u).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// isValidRole проверяет валидность роли
func isValidRole(role string) bool {
	switch role {
	case models.RoleAdministrator, models.RolePodft, models.RoleClientManagement:
		return true
	}
	return false
}

// CreateAppUser создает нового пользователя
func CreateAppUser(gdb *gorm.DB, email, role, token string) (models.AppUser, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return models.AppUser{}, fmt.Errorf("email required")
	}
	if !isValidRole(role) {
		return models.AppUser{}, fmt.Errorf("invalid role")
	}
	if strings.TrimSpace(token) == "" {
		token = uuid.NewString()
	}
	u := models.AppUser{
		Email: email,
		Role:  role,
		Token: token,
	}
	if err := gdb.Create(&u).Error; err != nil {
		return models.AppUser{}, err
	}
	return u, nil
}

// ListAppUsers возвращает список всех пользователей
func ListAppUsers(gdb *gorm.DB) ([]models.AppUser, error) {
	var xs []models.AppUser
	return xs, gdb.Order("id ASC").Find(&xs).Error
}

// UpdateUserRole обновляет роль пользователя
func UpdateUserRole(gdb *gorm.DB, id uint, role string) (models.AppUser, error) {
	if !isValidRole(role) {
		return models.AppUser{}, fmt.Errorf("invalid role")
	}
	var u models.AppUser
	if err := gdb.First(&u, id).Error; err != nil {
		return u, err
	}
	u.Role = role
	return u, gdb.Save(&u).Error
}

// RotateUserToken обновляет токен пользователя
func RotateUserToken(gdb *gorm.DB, id uint) (models.AppUser, error) {
	var u models.AppUser
	if err := gdb.First(&u, id).Error; err != nil {
		return u, err
	}
	u.Token = uuid.NewString()
	return u, gdb.Save(&u).Error
}

// DeleteAppUser удаляет пользователя
func DeleteAppUser(gdb *gorm.DB, id uint) error {
	return gdb.Delete(&models.AppUser{}, id).Error
}

package repository

import (
	"vector/internal/models"

	"gorm.io/gorm"
)

type userRepository struct {
	database *gorm.DB
}

func NewUserRepository(database *gorm.DB) UserRepository {
	return &userRepository{database: database}
}

func (r *userRepository) GetByID(id uint) (models.AppUser, error) {
	var user models.AppUser
	err := r.database.First(&user, id).Error
	return user, err
}

func (r *userRepository) GetByEmail(email string) (models.AppUser, error) {
	var user models.AppUser
	err := r.database.Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *userRepository) Create(user models.AppUser) (models.AppUser, error) {
	err := r.database.Create(&user).Error
	return user, err
}

func (r *userRepository) List() ([]models.AppUser, error) {
	var users []models.AppUser
	err := r.database.Order("created_at DESC").Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateRole(id uint, role string) (models.AppUser, error) {
	var user models.AppUser
	if err := r.database.First(&user, id).Error; err != nil {
		return user, err
	}

	user.Role = role
	err := r.database.Save(&user).Error
	return user, err
}

func (r *userRepository) UpdatePassword(id uint, passwordHash string) error {
	return r.database.Model(&models.AppUser{}).Where("id = ?", id).Update("password_hash", passwordHash).Error
}

func (r *userRepository) SetActive(id uint, isActive bool) error {
	return r.database.Model(&models.AppUser{}).Where("id = ?", id).Update("is_active", isActive).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.database.Delete(&models.AppUser{}, id).Error
}

func (r *userRepository) Seed() error {
	// Создаем администратора по умолчанию
	admin := models.AppUser{
		Email:      "admin@vector.com",
		FirstName:  "Администратор",
		LastName:   "Системы",
		MiddleName: "",
		Role:       models.RoleAdministrator,
		IsActive:   true,
	}

	// Устанавливаем пароль по умолчанию
	if err := admin.HashPassword("admin123"); err != nil {
		return err
	}

	// Проверяем, существует ли уже администратор
	var existing models.AppUser
	if err := r.database.Where("email = ?", admin.Email).First(&existing).Error; err == gorm.ErrRecordNotFound {
		return r.database.Create(&admin).Error
	}

	return nil
}

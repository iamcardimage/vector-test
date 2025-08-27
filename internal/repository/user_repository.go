package repository

import (
	appdb "vector/internal/db/app"
	"vector/internal/models"

	"gorm.io/gorm"
)

type userRepository struct {
	database *gorm.DB
}

func NewUserRepository(database *gorm.DB) UserRepository {
	return &userRepository{database: database}
}

func (r *userRepository) GetByToken(token string) (models.AppUser, error) {
	return appdb.GetUserByToken(r.database, token)
}

func (r *userRepository) Create(email, role, token string) (models.AppUser, error) {
	return appdb.CreateAppUser(r.database, email, role, token)
}

func (r *userRepository) List() ([]models.AppUser, error) {
	return appdb.ListAppUsers(r.database)
}

func (r *userRepository) UpdateRole(id uint, role string) (models.AppUser, error) {
	return appdb.UpdateUserRole(r.database, id, role)
}

func (r *userRepository) RotateToken(id uint) (models.AppUser, error) {
	return appdb.RotateUserToken(r.database, id)
}

func (r *userRepository) Delete(id uint) error {
	return appdb.DeleteAppUser(r.database, id)
}

func (r *userRepository) Seed() error {
	return appdb.SeedAppUsers(r.database)
}

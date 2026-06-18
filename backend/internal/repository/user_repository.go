package repository

import (
	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetAllUsers() ([]models.User, error)
	CreateUser(user models.User) (models.User, error)
	GetUserByGoogleId(googleId string) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
	GetUserById(id uint) (models.User, error)
	UpdateUser(user models.User) (models.User, error)
	DeleteUser(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (ur *userRepository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	if err := ur.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (ur *userRepository) CreateUser(user models.User) (models.User, error) {
	if err := ur.db.Create(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (ur *userRepository) GetUserByGoogleId(googleId string) (models.User, error) {
	var user models.User
	if err := ur.db.Where("google_id = ?", googleId).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (ur *userRepository) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	if err := ur.db.Where("email = ?", email).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (ur *userRepository) GetUserById(id uint) (models.User, error) {
	var user models.User
	if err := ur.db.First(&user, id).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (ur *userRepository) UpdateUser(user models.User) (models.User, error) {
	if err := ur.db.Save(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (ur *userRepository) DeleteUser(id uint) error {
	return ur.db.Delete(&models.User{}, id).Error
}

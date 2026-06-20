package services

import (
	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
)

type UserService interface {
	GetAllUsers() ([]models.User, error)
	CreateUser(user models.User) (models.User, error)
	GetUserByGoogleId(googleId string) (models.User, error)
	GetUserById(id uint) (models.User, error)
	UpdateRole(id uint, role string) error
}

type userService struct {
	UserRepository repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{UserRepository: userRepo}
}

func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.UserRepository.GetAllUsers()
}

func (s *userService) CreateUser(user models.User) (models.User, error) {
	return s.UserRepository.CreateUser(user)
}

func (s *userService) GetUserByGoogleId(googleId string) (models.User, error) {
	return s.UserRepository.GetUserByGoogleId(googleId)
}

func (s *userService) GetUserById(id uint) (models.User, error) {
	return s.UserRepository.GetUserById(id)
}

func (s *userService) UpdateRole(id uint, role string) error {
	return s.UserRepository.UpdateRole(id, role)
}

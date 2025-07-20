package services

import (
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(user *models.User) error {
	// TODO: Implement user creation logic
	return s.userRepo.Create(user)
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	// TODO: Implement get user by ID logic
	return s.userRepo.GetByID(id)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	// TODO: Implement get user by email logic
	return s.userRepo.GetByEmail(email)
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	// TODO: Implement get user by username logic
	return s.userRepo.GetByUsername(username)
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(user *models.User) error {
	// TODO: Implement user update logic
	return s.userRepo.Update(user)
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(id string) error {
	// TODO: Implement user deletion logic
	return s.userRepo.Delete(id)
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers() ([]*models.User, error) {
	// TODO: Implement get all users logic
	return s.userRepo.GetAll()
}

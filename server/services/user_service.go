package services

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/server/model"
	"github.com/stsg/gophkeeper2/server/repositories"
)

//go:generate mockgen -source=user_service.go -destination=../mocks/services/user_service.go -package=services

// Defines an interface named UserService.
// This interface has several methods that represent operations that can be
// performed on a user.
type UserService interface {
	CreateUser(ctx context.Context, user *model.User) (int32, error)
	GetUser(ctx context.Context, username string) (*model.User, error)
	ValidatePassword(_ context.Context, user *model.User, password string) (bool, error)
}

// Implements UserService.
type userService struct {
	log  *zap.SugaredLogger
	repo repositories.UserRepository
}

// NewUserService creates a new instance of UserService with the provided UserRepository.
//
// Parameters:
// - repo: The UserRepository to be used by the UserService.
//
// Returns:
// - A pointer to the newly created UserService.
func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{log: logger.NewLogger("user-srv"), repo: repo}
}

// CreateUser creates a new user in the system.
//
// Parameters:
// - ctx: The context.Context object for the request.
// - user: The model.User object representing the user to be created.
//
// Returns:
// - int32: The ID of the created user.
// - error: An error if the user creation fails.
func (s *userService) CreateUser(ctx context.Context, user *model.User) (int32, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(user.Password, 8)
	if err != nil {
		s.log.Errorf("failed to generate password of '%s' user", user.Username)
		return 0, err
	}
	newUser := &model.User{
		Username: user.Username,
		Password: hashedPassword,
	}
	return s.repo.CreateUser(ctx, newUser)
}

// GetUser retrieves a user by username.
//
// Parameters:
// - ctx: The context.Context object for the request.
// - username: The username of the user to retrieve.
//
// Returns:
// - *model.User: The user object if found.
// - error: An error if the user retrieval fails.
func (s *userService) GetUser(ctx context.Context, username string) (*model.User, error) {
	return s.repo.GetUser(ctx, username)
}

// ValidatePassword checks if the provided password matches the hashed password of a user.
//
// Parameters:
// - _ context.Context: The context.Context object for the request. It is not used in this function.
// - user *model.User
func (s *userService) ValidatePassword(_ context.Context, user *model.User, password string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			s.log.Warnf("Invalid password of '%s' user, id: %d", user.Username, user.Id)
			return false, nil
		}
		return false, err
	}
	return true, nil
}

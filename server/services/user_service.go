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

type UserService interface {
	CreateUser(ctx context.Context, user *model.User) (int32, error)
	GetUser(ctx context.Context, username string) (*model.User, error)
	ValidatePassword(_ context.Context, user *model.User, password string) (bool, error)
}

type userService struct {
	log  *zap.SugaredLogger
	repo repositories.UserRepository
	ctx  context.Context
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{log: logger.NewLogger("user-srv"), repo: repo}
}

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

func (s *userService) GetUser(ctx context.Context, username string) (*model.User, error) {
	return s.repo.GetUser(ctx, username)
}

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

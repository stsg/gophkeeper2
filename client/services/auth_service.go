package services

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stsg/gophkeeper2/client/model"
	"github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/pkg/pb"
)

//go:generate mockgen -source=auth_service.go -destination=../mocks/services/auth_service.go -package=services

type AuthService interface {
	Register(ctx context.Context, username string, password string) (*pb.TokenData, error)
	Login(ctx context.Context, username string, password string) (*pb.TokenData, error)
}

type authService struct {
	log        *zap.SugaredLogger
	authClient pb.AuthClient
	// refreshTokenOnce sync.Once
	tokenHolder *model.TokenHolder
}

func NewAuthService(
	client pb.AuthClient,
	tokenHolder *model.TokenHolder,
) AuthService {
	return &authService{
		log:         logger.NewLogger("auth-service"),
		authClient:  client,
		tokenHolder: tokenHolder,
	}
}

func (s *authService) Register(ctx context.Context, username string, password string) (*pb.TokenData, error) {
	tokenData, err := s.authClient.Register(ctx, &pb.AuthData{
		Username: username,
		Password: password,
	})

	if err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.AlreadyExists {
			return nil, errors.New(e.Message())
		}
		s.log.Errorf("failed to register: %v", err)
		return nil, err
	}
	s.tokenHolder.Set(tokenData.Token)

	return tokenData, nil
}

func (s *authService) Login(ctx context.Context, username string, password string) (*pb.TokenData, error) {
	tokenData, err := s.authClient.Login(
		ctx,
		&pb.AuthData{
			Username: username,
			Password: password,
		},
	)

	if err != nil {
		if statusErr, ok := status.FromError(err); ok && statusErr.Code() == codes.NotFound {
			return nil, errors.New(statusErr.Message())
		}
		return nil, err
	}

	s.tokenHolder.Set(tokenData.Token)
	return tokenData, nil
}

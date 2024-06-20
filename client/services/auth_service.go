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

// This Go code defines an interface named AuthService that has two methods: Register and Login.
type AuthService interface {
	Register(ctx context.Context, username string, password string) (*pb.TokenData, error)
	Login(ctx context.Context, username string, password string) (*pb.TokenData, error)
}

// The authService struct is a Go struct that represents an authentication service.
type authService struct {
	log        *zap.SugaredLogger
	authClient pb.AuthClient
	// refreshTokenOnce sync.Once
	tokenHolder *model.TokenHolder
}

// NewAuthService creates a new instance of AuthService with the given pb.AuthClient and model.TokenHolder.
//
// Parameters:
// - client: The pb.AuthClient to be used for authentication requests.
// - tokenHolder: The model.TokenHolder to store the token data.
//
// Returns:
// - AuthService: A new instance of AuthService.
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

// Register registers a new user with the given username and password.
//
// Parameters:
// - ctx: The context.Context object for the request.
// - username: The username of the new user.
// - password: The password of the new user.
//
// Returns:
// - *pb.TokenData: The token data for the newly registered user.
// - error: An error if the registration failed.
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

// Login logs in a user with the given username and password.
//
// Parameters:
// - ctx: The context.Context object for the request.
// - username: The username of the user.
// - password: The password of the user.
//
// Returns:
// - *pb.TokenData: The token data for the logged-in user.
// - error: An error if the login failed.
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

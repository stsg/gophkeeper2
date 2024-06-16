package grpc_servers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/pkg/pb"
	"github.com/stsg/gophkeeper2/server/model"
	"github.com/stsg/gophkeeper2/server/model/errs"
	"github.com/stsg/gophkeeper2/server/services"
)

type authServer struct {
	log *zap.SugaredLogger
	pb.UnimplementedAuthServer
	userService  services.UserService
	tokenService services.TokenService
}

func NewAuthServer(userService services.UserService, tokenService services.TokenService) pb.AuthServer {
	return &authServer{
		log:          logger.NewLogger("auth-server"),
		userService:  userService,
		tokenService: tokenService,
	}
}

func (s *authServer) Register(ctx context.Context, authData *pb.AuthData) (*pb.TokenData, error) {
	s.log.Infof("Handle registration of '%s' user", authData.Username)
	if err := s.validateAuthData(authData); err != nil {
		return nil, err
	}

	user := &model.User{Username: authData.Username, Password: []byte(authData.Password)}
	id, err := s.userService.CreateUser(ctx, user)
	if errors.Is(err, errs.ErrUserAlreadyExist) {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	if err != nil {
		s.log.Errorf("failed to create user: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create user: %v", err))
	}
	s.log.Infof("User '%s' registered, id: %d", user.Username, id)
	return s.genToken(id)
}

func (s *authServer) Login(ctx context.Context, authData *pb.AuthData) (*pb.TokenData, error) {
	s.log.Infof("Handle logging of '%s' user", authData.Username)
	if err := s.validateAuthData(authData); err != nil {
		return nil, err
	}
	user, err := s.userService.GetUser(ctx, authData.Username)
	if errors.Is(err, errs.ErrUserNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		s.log.Errorf("failed to get user: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get user: %v", err))
	}
	ok, err := s.userService.ValidatePassword(ctx, user, authData.Password)
	if err != nil {
		s.log.Errorf("failed to check user password: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check user password: %v", err))
	}
	if !ok {
		s.log.Warn("password is incorrect")
		return nil, status.Error(codes.InvalidArgument, "password is incorrect")
	}
	s.log.Infof("User '%s' logged, id: %d", user.Username, user.Id)
	return s.genToken(user.Id)
}

func (s *authServer) validateAuthData(authData *pb.AuthData) error {
	s.log.Info("Validate auth request")
	if len(authData.Username) == 0 {
		s.log.Errorf("username is empty")
		return status.Error(codes.InvalidArgument, "invalid username format: must be nonempty")
	}
	if len(authData.Password) == 0 {
		s.log.Errorf("password is empty")
		return status.Error(codes.InvalidArgument, "invalid password format: must be nonempty")
	}
	return nil
}

func (s *authServer) genToken(id int32) (*pb.TokenData, error) {
	s.log.Infof("Generating token for user %d", id)
	expireAt := time.Now().UTC().Add(time.Hour)
	token, err := s.tokenService.Generate(id, expireAt)
	if err != nil {
		s.log.Errorf("failed to generate token: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("token generation error: %v", err))
	}
	s.log.Infof("Token generated successfully: %v", zap.Time("expireAt", expireAt))
	return &pb.TokenData{Token: token, ExpireAt: timestamppb.New(expireAt)}, nil
}

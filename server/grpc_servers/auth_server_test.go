package grpc_servers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stsg/gophkeeper2/pkg/pb"
	"github.com/stsg/gophkeeper2/server/mocks/services"
	"github.com/stsg/gophkeeper2/server/model"
	"github.com/stsg/gophkeeper2/server/model/errs"
)

func TestAuthServer_Register_UsernameError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userService := services.NewMockUserService(ctrl)
	tokenService := services.NewMockTokenService(ctrl)
	authServer := NewAuthServer(userService, tokenService)

	data := &pb.AuthData{
		Username: "",
		Password: "test",
	}

	token, err := authServer.Register(ctx, data)
	assert.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid username format: must be nonempty"))
	assert.Nil(t, token)
}

func TestAuthServer_Register_PasswordError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userService := services.NewMockUserService(ctrl)
	tokenService := services.NewMockTokenService(ctrl)
	authServer := NewAuthServer(userService, tokenService)

	data := &pb.AuthData{
		Username: "test",
		Password: "",
	}

	token, err := authServer.Register(ctx, data)
	assert.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid password format: must be nonempty"))
	assert.Nil(t, token)
}

func TestAuthServer_Register_Success(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userService := services.NewMockUserService(ctrl)
	tokenService := services.NewMockTokenService(ctrl)
	authServer := NewAuthServer(userService, tokenService)

	user := &model.User{
		Username: "test",
		Password: []byte("test"),
	}
	id := int32(1)
	userService.
		EXPECT().
		CreateUser(ctx, gomock.Eq(user)).Return(id, nil)

	token := "iAmToken"
	tokenService.
		EXPECT().
		Generate(id, gomock.AssignableToTypeOf(time.Time{})).
		Return(token, nil)

	data := &pb.AuthData{
		Username: user.Username,
		Password: string(user.Password),
	}

	tokenData, err := authServer.Register(ctx, data)
	assert.NoError(t, err)
	assert.NotNil(t, tokenData)
	assert.NotNil(t, tokenData.ExpireAt)

	assert.Equal(t, token, tokenData.Token)
}

func TestAuthServer_Register_TokenGenerateError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userService := services.NewMockUserService(ctrl)
	tokenService := services.NewMockTokenService(ctrl)
	authServer := NewAuthServer(userService, tokenService)

	user := &model.User{
		Username: "test",
		Password: []byte("test"),
	}
	id := int32(1)
	userService.
		EXPECT().
		CreateUser(ctx, gomock.Eq(user)).Return(id, nil)

	tokenService.
		EXPECT().
		Generate(id, gomock.AssignableToTypeOf(time.Time{})).
		Return("", errors.New("do not want to generate token"))

	data := &pb.AuthData{
		Username: user.Username,
		Password: string(user.Password),
	}

	tokenData, err := authServer.Register(ctx, data)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "token generation error: do not want to generate token")
	assert.Nil(t, tokenData)
}

func TestAuthServer_Register(t *testing.T) {
	// TODO: add tests
}

func TestAuthServer_Login_Success(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userService := services.NewMockUserService(ctrl)
	tokenService := services.NewMockTokenService(ctrl)
	authServer := NewAuthServer(userService, tokenService)

	id := int32(1)
	user := &model.User{
		Id:       id,
		Username: "test",
		Password: []byte("test"),
	}
	userService.
		EXPECT().
		GetUser(ctx, user.Username).Return(user, nil)

	userService.
		EXPECT().
		ValidatePassword(ctx, user, "test").
		Return(true, nil)

	token := "iAmToken"
	tokenService.
		EXPECT().
		Generate(id, gomock.AssignableToTypeOf(time.Time{})).
		Return(token, nil)

	data := &pb.AuthData{
		Username: user.Username,
		Password: string(user.Password),
	}

	tokenData, err := authServer.Login(ctx, data)

	assert.NoError(t, err)
	assert.NotNil(t, tokenData)
	assert.NotNil(t, tokenData.ExpireAt)

	assert.Equal(t, token, tokenData.Token)
}

// Successful registration with valid username and password
func TestRegister_SuccessfulRegistration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := services.NewMockUserService(ctrl)
	mockTokenService := services.NewMockTokenService(ctrl)
	logger := zap.NewNop().Sugar()

	authServer := &authServer{
		log:          logger,
		userService:  mockUserService,
		tokenService: mockTokenService,
	}

	ctx := context.Background()
	authData := &pb.AuthData{
		Username: "testuser",
		Password: "testpassword",
	}

	user := &model.User{Username: authData.Username, Password: []byte(authData.Password)}
	mockUserService.EXPECT().CreateUser(ctx, user).Return(int32(1), nil)
	mockTokenService.EXPECT().Generate(int32(1), gomock.Any()).Return("testtoken", nil)

	tokenData, err := authServer.Register(ctx, authData)

	require.NoError(t, err)
	require.NotNil(t, tokenData)
	require.Equal(t, "testtoken", tokenData.Token)
}

// Registration with an already existing username
func TestRegister_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := services.NewMockUserService(ctrl)
	mockTokenService := services.NewMockTokenService(ctrl)
	logger := zap.NewNop().Sugar()

	authServer := &authServer{
		log:          logger,
		userService:  mockUserService,
		tokenService: mockTokenService,
	}

	ctx := context.Background()
	authData := &pb.AuthData{
		Username: "existinguser",
		Password: "testpassword",
	}

	user := &model.User{Username: authData.Username, Password: []byte(authData.Password)}
	mockUserService.EXPECT().CreateUser(ctx, user).Return(int32(0), errs.ErrUserAlreadyExist)

	tokenData, err := authServer.Register(ctx, authData)

	require.Error(t, err)
	require.Nil(t, tokenData)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.AlreadyExists, st.Code())
}

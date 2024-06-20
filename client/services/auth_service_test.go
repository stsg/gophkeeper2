package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	services "github.com/stsg/gophkeeper2/client/mocks/services"
	"github.com/stsg/gophkeeper2/pkg/pb"
)

func Register_Test(t *testing.T) {
	// TODO: add test
}

func Login_Test(t *testing.T) {
	// TODO: add test
}

func TestNewAuthService_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := services.NewMockAuthService(ctrl)

	assert.NotNil(t, authService)
}

func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := services.NewMockAuthService(ctrl)

	ctx := context.Background()
	username := "testuser"
	password := "testpassword"

	authService.EXPECT().Register(ctx, username, password).AnyTimes().Return(&pb.TokenData{Token: "testtoken"}, nil)

}

func TestLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := services.NewMockAuthService(ctrl)

	ctx := context.Background()
	username := "testuser"
	password := "testpassword"

	authService.EXPECT().Login(ctx, username, password).AnyTimes().Return(&pb.TokenData{Token: "testtoken"}, nil)

}

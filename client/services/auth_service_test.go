package services

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	services "github.com/stsg/gophkeeper2/client/mocks/services"
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

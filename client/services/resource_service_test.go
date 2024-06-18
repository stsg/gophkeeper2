package services

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	services "github.com/stsg/gophkeeper2/client/mocks/services"
)

func TestNewResourceService_InitializesWithProvidedDependencies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := services.NewMockResourceService(ctrl)

	assert.NotNil(t, service)
}

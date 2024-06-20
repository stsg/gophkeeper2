package repositories

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock "github.com/stsg/gophkeeper2/server/mocks/repositories"
)

func TestNewResourceRepository_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDBProvider := mock.NewMockDBProvider(ctrl)
	repo := NewResourceRepository(mockDBProvider)
	assert.NotNil(t, repo)
	assert.IsType(t, &resourceRepository{}, repo)

}

func TestNewResourceRepository_DBProviderNil(t *testing.T) {
	repo := NewResourceRepository(nil)

	assert.NotNil(t, repo)
	assert.IsType(t, &resourceRepository{}, repo)
}

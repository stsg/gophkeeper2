package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stsg/gophkeeper2/pkg/model/enum"
	"github.com/stsg/gophkeeper2/server/mocks/repositories"
	"github.com/stsg/gophkeeper2/server/model"
)

// Successfully save a file description with valid userId, meta, and data
func TestSaveFileDescription_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockResourceRepository(ctrl)
	service := NewResourceService(mockRepo)

	ctx := context.Background()
	userId := int32(1)
	meta := []byte("meta")
	data := []byte("data")

	resource := &model.Resource{
		UserId: userId,
		Data:   data,
		ResourceDescription: model.ResourceDescription{
			Type: enum.File,
			Meta: meta,
		},
	}

	mockRepo.EXPECT().Save(ctx, resource).Return(nil)

	id, err := service.SaveFileDescription(ctx, userId, meta, data)

	assert.NoError(t, err)
	assert.Equal(t, resource.Id, id)
}

// Handle error when repository fails to save the resource
func TestSaveFileDescription_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockResourceRepository(ctrl)
	service := NewResourceService(mockRepo)

	ctx := context.Background()
	userId := int32(1)
	meta := []byte("meta")
	data := []byte("data")
	resource := &model.Resource{
		UserId: userId,
		Data:   data,
		ResourceDescription: model.ResourceDescription{
			Type: enum.File,
			Meta: meta,
		},
	}

	expectedErr := errors.New("repository error")
	mockRepo.EXPECT().Save(ctx, resource).Return(expectedErr)

	id, err := service.SaveFileDescription(ctx, userId, meta, data)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, int32(0), id)
}

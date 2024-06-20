package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stsg/gophkeeper2/client/model/resources"

	"github.com/stretchr/testify/assert"
	services "github.com/stsg/gophkeeper2/client/mocks/services"
	"github.com/stsg/gophkeeper2/pkg/model/enum"
	"github.com/stsg/gophkeeper2/server/model"
)

func TestNewResourceService_InitializesWithProvidedDependencies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := services.NewMockResourceService(ctrl)

	assert.NotNil(t, service)
}

func TestGetDescriptions_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := services.NewMockResourceService(ctrl)

	ctx := context.Background()
	resType := enum.Nan

	resDescription := &model.ResourceDescription{
		Id:   int32(1),
		Meta: []byte("meta1"),
		Type: resType,
	}

	service.EXPECT().GetDescriptions(ctx, resType).AnyTimes().Return([]*model.ResourceDescription{resDescription}, nil)
	descriptions, err := service.GetDescriptions(ctx, resType)
	assert.NoError(t, err)
	assert.Len(t, descriptions, 1)
	assert.Equal(t, int32(1), descriptions[0].Id)
	assert.Equal(t, []byte("meta1"), descriptions[0].Meta)
	assert.Equal(t, resType, descriptions[0].Type)

}

func TestGetDescriptions_HandleEOF(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := services.NewMockResourceService(ctrl)

	ctx := context.Background()
	resType := enum.Nan

	resDescription := &model.ResourceDescription{}
	service.EXPECT().GetDescriptions(ctx, resType).Return([]*model.ResourceDescription{resDescription}, nil)

	descriptions, err := service.GetDescriptions(ctx, resType)

	assert.NoError(t, err)
	assert.Len(t, descriptions, 1)
}

func TestGet_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := services.NewMockResourceService(ctrl)

	ctx := context.Background()
	resType := int32(0)

	service.EXPECT().Get(ctx, resType).AnyTimes().Return(&resources.Info{}, nil)
	info, err := service.Get(ctx, resType)
	assert.NoError(t, err)
	assert.Equal(t, []byte(nil), info.Meta)
}

func TestSave_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := services.NewMockResourceService(ctrl)

	path := []byte("../../cfg/config.json")
	meta := []byte("file meta")

	ctx := context.Background()
	resType := enum.File

	service.EXPECT().Save(ctx, resType, path, meta).AnyTimes().Return(int32(0), nil)
	info, err := service.Save(ctx, resType, path, meta)
	assert.NoError(t, err)
	assert.Equal(t, int32(0), info)
}

func TestGetFile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := services.NewMockResourceService(ctrl)

	ctx := context.Background()
	resType := enum.File

	service.EXPECT().GetFile(ctx, int32(resType)).AnyTimes().Return("", nil)
	info, err := service.GetFile(ctx, int32(resType))
	assert.NoError(t, err)
	assert.Equal(t, "", info)
}

func TestGetDescriptions_ThreadSafety(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := services.NewMockResourceService(ctrl)

	ctx := context.Background()
	resType := enum.ResourceType(1)

	resDescription := &model.ResourceDescription{}
	service.EXPECT().GetDescriptions(ctx, resType).Return([]*model.ResourceDescription{resDescription}, nil)

	descriptions, err := service.GetDescriptions(ctx, resType)

	assert.NoError(t, err)
	assert.Len(t, descriptions, 1)
	assert.Equal(t, int32(0), descriptions[0].Id)
	assert.Equal(t, []uint8([]byte(nil)), descriptions[0].Meta)
	assert.Equal(t, enum.ResourceType(0), descriptions[0].Type)
}

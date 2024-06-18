package services

import (
	"context"

	"go.uber.org/zap"

	"github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/pkg/model/enum"
	"github.com/stsg/gophkeeper2/server/model"
	"github.com/stsg/gophkeeper2/server/repositories"
)

//go:generate mockgen -source=resource_service.go -destination=../mocks/services/resource_service.go -package=services

// Defines an interface named ResourceService.
// This interface has several methods that represent operations that can be
// performed on a resource.
type ResourceService interface {
	Save(ctx context.Context, res *model.Resource) error
	Update(ctx context.Context, res *model.Resource) error
	Delete(ctx context.Context, resId, userId int32) error
	GetDescriptions(ctx context.Context, userId int32, resType enum.ResourceType) ([]*model.ResourceDescription, error)
	Get(ctx context.Context, resId int32, userId int32) (*model.Resource, error)
	SaveFileDescription(ctx context.Context, userId int32, meta []byte, data []byte) (int32, error)
	GetFileDescription(ctx context.Context, resource *model.Resource) ([]byte, error)
}

// The resourceService struct is a type that represents a service for managing
// resources. It has two fields: log, which is a logger for logging messages,
// and repo, which is an instance of the ResourceRepository interface for
// interacting with the resource repository.
type resourceService struct {
	log  *zap.SugaredLogger
	repo repositories.ResourceRepository
}

func NewResourceService(repo repositories.ResourceRepository) ResourceService {
	return &resourceService{log: logger.NewLogger("res-service"), repo: repo}
}

func (s *resourceService) Save(ctx context.Context, data *model.Resource) error {
	return s.repo.Save(ctx, data)
}

func (s *resourceService) Update(ctx context.Context, data *model.Resource) error {
	return s.repo.Update(ctx, data)
}

func (s *resourceService) Delete(ctx context.Context, resId int32, userId int32) error {
	return s.repo.Delete(ctx, resId, userId)
}

func (s *resourceService) GetDescriptions(ctx context.Context, userId int32, resType enum.ResourceType) ([]*model.ResourceDescription, error) {
	return s.repo.GetResDescriptionsByType(ctx, userId, resType)
}

func (s *resourceService) Get(ctx context.Context, resId int32, userId int32) (*model.Resource, error) {
	return s.repo.Get(ctx, resId, userId)
}

func (s *resourceService) SaveFileDescription(ctx context.Context, userId int32, meta []byte, data []byte) (int32, error) {
	resource := &model.Resource{
		UserId: userId,
		Data:   data,
	}
	resource.Type = enum.File
	resource.Meta = meta

	err := s.repo.Save(ctx, resource)
	if err != nil {
		return 0, err
	}

	return resource.Id, nil
}

func (s *resourceService) GetFileDescription(ctx context.Context, resource *model.Resource) ([]byte, error) {
	res, err := s.repo.Get(ctx, resource.Id, resource.UserId)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

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

// NewResourceService creates a new instance of the ResourceService interface.
//
// Parameters:
// - repo: an instance of the ResourceRepository interface for interacting with the resource repository.
//
// Returns:
// - ResourceService: a pointer to the resourceService struct that implements the ResourceService interface.
func NewResourceService(repo repositories.ResourceRepository) ResourceService {
	return &resourceService{log: logger.NewLogger("res-service"), repo: repo}
}

// Save saves the given resource data using the resource repository.
//
// Parameters:
// - ctx: the context for the operation.
// - data: the resource data to be saved.
// Return type: error.
func (s *resourceService) Save(ctx context.Context, data *model.Resource) error {
	return s.repo.Save(ctx, data)
}

// Update updates the given resource data using the resource repository.
//
// Parameters:
// - ctx: the context for the operation.
// - data: the resource data to be updated.
//
// Returns:
// - error: an error if the update operation fails.
func (s *resourceService) Update(ctx context.Context, data *model.Resource) error {
	return s.repo.Update(ctx, data)
}

// Delete deletes a resource from the resource repository.
//
// Parameters:
// - ctx: the context for the operation.
// - resId: the ID of the resource to be deleted.
// - userId: the ID of the user who is deleting the resource.
//
// Returns:
// - error: an error if the deletion fails, nil otherwise.
func (s *resourceService) Delete(ctx context.Context, resId int32, userId int32) error {
	return s.repo.Delete(ctx, resId, userId)
}

// GetDescriptions retrieves the descriptions of resources based on the provided user ID and resource type.
//
// Parameters:
// - ctx: The context.Context object for the function.
// - userId: The ID of the user.
// - resType: The type of the resource.
//
// Returns:
// - []*model.ResourceDescription: The descriptions of the resources.
// - error: An error if the retrieval fails.
func (s *resourceService) GetDescriptions(ctx context.Context, userId int32, resType enum.ResourceType) ([]*model.ResourceDescription, error) {
	return s.repo.GetResDescriptionsByType(ctx, userId, resType)
}

// Get retrieves a resource from the repository based on the provided resource ID and user ID.
//
// Parameters:
//   - ctx: The context.Context object for the function.
//   - resId: The ID of the resource to retrieve.
//   - userId: The ID of the user.
//
// Returns:
//   - *model.Resource: The retrieved resource, or nil if not found.
//   - error: An error if the retrieval fails.
func (s *resourceService) Get(ctx context.Context, resId int32, userId int32) (*model.Resource, error) {
	return s.repo.Get(ctx, resId, userId)
}

// SaveFileDescription saves the file description with the provided user ID, meta, and data.
//
// Parameters:
// - ctx: The context for the operation.
// - userId: The ID of the user.
// - meta: The metadata of the file.
// - data: The data content of the file.
// Return type: int32, error.
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

// GetFileDescription retrieves the file description for a given resource.
//
// Parameters:
// - ctx: The context.Context object for the function.
// - resource: A pointer to the model.Resource object representing the resource.
//
// Returns:
// - []byte: The file description data.
// - error: An error if the retrieval fails.
func (s *resourceService) GetFileDescription(ctx context.Context, resource *model.Resource) ([]byte, error) {
	res, err := s.repo.Get(ctx, resource.Id, resource.UserId)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

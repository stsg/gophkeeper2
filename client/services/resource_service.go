package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/stsg/gophkeeper2/client/model/resources"
	"github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/pkg/model/enum"
	"github.com/stsg/gophkeeper2/pkg/pb"
	intsrv "github.com/stsg/gophkeeper2/pkg/services"
	"github.com/stsg/gophkeeper2/server/model"
)

//go:generate mockgen -source=resource_service.go -destination=../mocks/services/resource_service.go -package=services

type ResourceService interface {
	Save(ctx context.Context, resType enum.ResourceType, data []byte, meta []byte) (int32, error)
	Update(ctx context.Context, resId int32, resType enum.ResourceType, data []byte, meta []byte) error
	Delete(ctx context.Context, resId int32) error
	GetDescriptions(ctx context.Context, resType enum.ResourceType) ([]*model.ResourceDescription, error)
	Get(ctx context.Context, resId int32) (*resources.Info, error)
	SaveFile(ctx context.Context, path string, meta []byte) (int32, error)
	GetFile(ctx context.Context, resId int32) (string, error)
}

type resourceService struct {
	log            *zap.SugaredLogger
	resourceClient pb.ResourcesClient
	fileService    intsrv.FileService
	cryptoService  CryptService
}

func NewResourceService(
	client pb.ResourcesClient,
	fileService intsrv.FileService,
	cryptoService CryptService,
) ResourceService {
	return &resourceService{
		log:            logger.NewLogger("res-service"),
		resourceClient: client,
		fileService:    fileService,
		cryptoService:  cryptoService,
	}
}

func (s *resourceService) Save(
	ctx context.Context,
	resType enum.ResourceType,
	data []byte,
	meta []byte,
) (int32, error) {
	encryptedData, err := s.cryptoService.Encrypt(data)
	if err != nil {
		return 0, err
	}
	resId, err := s.resourceClient.Save(ctx, &pb.Resource{
		Type: pb.TYPE(resType),
		Data: encryptedData,
		Meta: meta,
	})
	if err != nil {
		return 0, err
	}
	return resId.GetId(), nil
}

func (s *resourceService) Update(
	ctx context.Context,
	resId int32,
	resType enum.ResourceType,
	data []byte,
	meta []byte,
) error {
	encryptedData, err := s.cryptoService.Encrypt(data)
	if err != nil {
		return err
	}
	_, err = s.resourceClient.Update(ctx, &pb.Resource{
		Id:   resId,
		Type: pb.TYPE(resType),
		Data: encryptedData,
		Meta: meta,
	})
	return err
}

func (s *resourceService) Delete(ctx context.Context, resId int32) error {
	_, err := s.resourceClient.Delete(ctx, &pb.ResourceId{Id: resId})
	return err
}

func (s *resourceService) GetDescriptions(ctx context.Context, resType enum.ResourceType) ([]*model.ResourceDescription, error) {
	stream, err := s.resourceClient.GetDescriptions(ctx, &pb.Query{ResourceType: pb.TYPE(resType)})
	if err != nil {
		return nil, err
	}
	results := make([]*model.ResourceDescription, 0)
	for {
		descr, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, &model.ResourceDescription{
			Id:   descr.Id,
			Meta: descr.Meta,
			Type: enum.ResourceType(descr.Type),
		})
	}
	return results, nil
}

func (s *resourceService) Get(ctx context.Context, resId int32) (*resources.Info, error) {
	resource, err := s.resourceClient.Get(ctx, &pb.ResourceId{Id: resId})
	if err != nil {
		return nil, err
	}
	decryptedData, err := s.cryptoService.Decrypt(resource.Data)
	if err != nil {
		return nil, err
	}
	resource.Data = decryptedData
	return s.parseResource(resource)
}

func (s *resourceService) parseResource(resource *pb.Resource) (*resources.Info, error) {
	switch enum.ResourceType(resource.Type) {
	case enum.LoginPassword:
		var loginPassword resources.LoginPassword
		if err := json.Unmarshal(resource.Data, &loginPassword); err != nil {
			return nil, err
		}
		return &resources.Info{Resource: &loginPassword, Meta: resource.Meta}, nil
	case enum.BankCard:
		var bankCard resources.BankCard
		if err := json.Unmarshal(resource.Data, &bankCard); err != nil {
			return nil, err
		}
		return &resources.Info{Resource: &bankCard, Meta: resource.Meta}, nil
	case enum.File:
		var file resources.File
		if err := json.Unmarshal(resource.Data, &file); err != nil {
			return nil, err
		}
		return &resources.Info{Resource: &file, Meta: resource.Meta}, nil
	default:
		return nil, fmt.Errorf("undefined type %v", resource.Type)
	}
}

func (s *resourceService) SaveFile(ctx context.Context, path string, meta []byte) (int32, error) {
	stream, err := s.resourceClient.SaveFile(ctx)
	if err != nil {
		return 0, err
	}
	errCh := make(chan error)
	chunks, stat, err := s.fileService.ReadFile(path, errCh)
	if err != nil {
		return 0, err
	}
	fileDescriptionJson, err := json.Marshal(resources.File{
		Name:      stat.Name(),
		Extension: filepath.Ext(path),
		Size:      stat.Size(),
	})
	if err != nil {
		return 0, err
	}

	err = stream.Send(&pb.FileChunk{
		Meta: meta,
		Data: fileDescriptionJson,
	})
	if err != nil {
		return 0, err
	}
	for {
		chunk, ok := <-chunks
		if !ok {
			break
		}
		encrypt, err := s.cryptoService.Encrypt(chunk)
		if err != nil {
			return 0, err
		}
		err = stream.Send(&pb.FileChunk{
			Data: encrypt,
		})
		if err != nil {
			errCh <- err
			return 0, err
		}
	}
	resId, err := stream.CloseAndRecv()
	if err != nil {
		return 0, err
	}
	return resId.Id, nil
}

func (s *resourceService) GetFile(ctx context.Context, resId int32) (string, error) {
	stream, err := s.resourceClient.GetFile(ctx, &pb.ResourceId{Id: resId})
	if err != nil {
		return "", err
	}
	chunk, err := stream.Recv()
	if err != nil {
		return "", err
	}
	var fileDescription resources.File
	err = json.Unmarshal(chunk.Data, &fileDescription)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("./%s", fileDescription.Name)
	chunks := make(chan []byte)
	errCh, err := s.fileService.SaveFile(path, chunks)
	if err != nil {
		return "", err
	}
Loop:
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			close(chunks)
			break Loop
		}
		if err != nil {
			close(chunks)
			s.log.Errorf("failed to recieve file stream chunk: %v", err)
			return "", err
		}

		decrypt, err := s.cryptoService.Decrypt(chunk.Data)
		if err != nil {
			s.log.Errorf("failed to decrypt file stream chunk: %v", err)
			return "", err
		}
		select {
		case chunks <- decrypt:
		case <-errCh:
			close(chunks)
			break Loop
		}
	}
	return path, err
}

package grpc_servers

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/pkg/model/enum"
	"github.com/stsg/gophkeeper2/pkg/pb"
	intsrv "github.com/stsg/gophkeeper2/pkg/services"
	"github.com/stsg/gophkeeper2/pkg/shutdown"
	"github.com/stsg/gophkeeper2/server/model"
	"github.com/stsg/gophkeeper2/server/model/consts"
	"github.com/stsg/gophkeeper2/server/model/errs"
	"github.com/stsg/gophkeeper2/server/services"
)

type ResourceServer struct {
	log *zap.SugaredLogger
	pb.UnimplementedResourcesServer
	service     services.ResourceService
	fileService intsrv.FileService
	eh          shutdown.ExitHandler
}

func NewResourcesServer(
	service services.ResourceService,
	fileService intsrv.FileService,
	eh shutdown.ExitHandler,
) pb.ResourcesServer {
	return &ResourceServer{
		log:         logger.NewLogger("res-service"),
		service:     service,
		fileService: fileService,
		eh:          eh,
	}
}

func (s *ResourceServer) Save(ctx context.Context, resource *pb.Resource) (*pb.ResourceId, error) {
	res := &model.Resource{
		UserId: s.getUserIdFromCtx(ctx),
		Data:   resource.Data,
	}
	res.Meta = resource.Meta

	res.Type = enum.ResourceType(resource.Type)
	s.log.Infof("Saving resource: %v", res.ResourceDescription)
	err := s.service.Save(ctx, res)
	if err != nil {
		s.log.Errorf("failed to save resource: %v", res.ResourceDescription)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ResourceId{Id: res.Id}, nil
}

func (s *ResourceServer) Update(ctx context.Context, resource *pb.Resource) (*emptypb.Empty, error) {
	res := &model.Resource{

		UserId: s.getUserIdFromCtx(ctx),
		Data:   resource.Data,
	}
	res.Id = resource.Id
	res.Meta = resource.Meta
	res.Type = enum.ResourceType(resource.Type)

	s.log.Infof("Updating resource: %v", res.ResourceDescription)
	err := s.service.Update(ctx, res)
	if err != nil {
		s.log.Errorf("failed to update resource: %v", res.ResourceDescription)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *ResourceServer) Delete(ctx context.Context, resId *pb.ResourceId) (*emptypb.Empty, error) {
	s.log.Infof("Deleting resource: %d", resId.Id)
	if err := s.service.Delete(ctx, resId.Id, s.getUserIdFromCtx(ctx)); err != nil {
		s.log.Errorf("failed to delete resource: %d", resId.Id)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *ResourceServer) GetDescriptions(query *pb.Query, stream pb.Resources_GetDescriptionsServer) error {
	t := enum.ResourceType(query.ResourceType)
	userId := s.getUserIdFromCtx(stream.Context())
	s.log.Infof("Getting list descriptions of resources for user: %d", userId)
	resourceDescriptions, err := s.service.GetDescriptions(stream.Context(), userId, t)
	if err != nil {
		s.log.Errorf("failed to collect list descriptions of resources for user: %d", userId)
		return status.Error(codes.Internal, err.Error())
	}

	for _, resDescription := range resourceDescriptions {
		err := stream.Send(&pb.ResourceDescription{
			Id:   resDescription.Id,
			Type: pb.TYPE(resDescription.Type),
			Meta: resDescription.Meta,
		})
		if err != nil {
			s.log.Errorf("failed to send '%v' of  user %d: %v", resDescription, userId, err)
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}

func (s *ResourceServer) Get(ctx context.Context, id *pb.ResourceId) (*pb.Resource, error) {
	s.log.Infof("Getting resource: %d", id.GetId())
	result, err := s.service.Get(ctx, id.Id, s.getUserIdFromCtx(ctx))
	if err != nil {
		s.log.Errorf("failed to get resource '%d': %v", id.GetId(), err)
		if errors.Is(err, errs.ErrResNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Resource{
		Type: pb.TYPE(result.Type),
		Data: result.Data,
		Meta: result.Meta,
	}, nil
}

func (s *ResourceServer) SaveFile(stream pb.Resources_SaveFileServer) error {
	userId := s.getUserIdFromCtx(stream.Context())
	s.log.Infof("Saving file resource, user: %d", userId)
	s.eh.AddFuncInProcessing(fmt.Sprintf("saving file for user: %d", userId))
	defer s.eh.FuncFinished(fmt.Sprintf("saving file for user: %d", userId))
	chunk, err := stream.Recv()
	if err == io.EOF {
		s.log.Errorf("failed to save file resource for '%d' user: empty stream", userId)
		return status.Error(codes.InvalidArgument, "empty stream")
	}
	if err != nil {
		s.log.Errorf("failed to save file resource for '%d' user: %v", userId, err)
		return err
	}
	chunks := make(chan []byte)

	resId, err := s.service.SaveFileDescription(
		stream.Context(),
		userId,
		chunk.Meta,
		chunk.Data,
	)
	if err != nil {
		s.log.Errorf("failed to save file '%s' description for '%d' user: %v", string(chunk.Meta), userId, err)
		return err
	}
	errCh, err := s.fileService.SaveFile(fmt.Sprintf("./cmd/server/%d", resId), chunks)
	if err != nil {
		s.log.Errorf("failed to save file '%d' for '%d' user: %v", resId, userId, err)
		return status.Error(codes.Internal, err.Error())
	}
Loop:
	for {
		chunk, err = stream.Recv()
		if err == io.EOF {
			s.log.Debugf("End of stream, resource: %d", resId)
			close(chunks)
			break Loop
		}
		if err != nil {
			close(chunks)
			s.log.Errorf("failed to get stream chunk, resource: %d", resId)
			return status.Error(codes.Internal, errs.StreamError{Err: err}.Error())
		}
		select {
		case chunks <- chunk.Data:
		case err := <-errCh:
			s.log.Errorf("failed to save stream chunk of '%d' resource: %v", resId, err)
			close(chunks)
			return status.Error(codes.Internal, err.Error())
		}
	}

	id := &pb.ResourceId{Id: resId}
	s.log.Infof("File '%d' was saved successfully", resId)
	return stream.SendAndClose(id)
}

func (s *ResourceServer) GetFile(resId *pb.ResourceId, stream pb.Resources_GetFileServer) error {
	s.log.Infof("Sending file resource: %d", resId.GetId())
	s.eh.AddFuncInProcessing(fmt.Sprintf("sending file: %d", resId.GetId()))
	defer s.eh.FuncFinished(fmt.Sprintf("sending file: %d", resId.GetId()))
	userId := s.getUserIdFromCtx(stream.Context())
	resource, err := s.service.Get(stream.Context(), resId.GetId(), userId)
	if err != nil {
		s.log.Errorf("failed to get '%d' file description for '%d' user: %v", resId.GetId(), userId, err)
		return status.Error(codes.Internal, err.Error())
	}
	err = stream.Send(&pb.FileChunk{
		Meta: resource.Meta,
		Data: resource.Data,
	})
	if err != nil {
		s.log.Errorf("failed to send '%d' file description for '%d' user: %v", resId.GetId(), userId, err)
		return status.Error(codes.Internal, errs.StreamError{Err: err}.Error())
	}
	errCh := make(chan error)
	chunks, _, err := s.fileService.ReadFile(fmt.Sprintf("./cmd/server/%d", resource.Id), errCh)
	if err != nil {
		s.log.Errorf("failed to read file '%d': %v", resource.Id, err)
		return status.Error(codes.Internal, err.Error())
	}

Loop:
	for {
		chunk, ok := <-chunks
		if !ok {
			break Loop
		}
		err := stream.Send(&pb.FileChunk{
			Meta: nil,
			Data: chunk,
		})
		if err != nil {
			s.log.Errorf("failed to send '%d' file's chunk: %v", resource.Id, err)
			errCh <- err
			return status.Error(codes.Internal, errs.StreamError{Err: err}.Error())
		}
	}
	return nil
}

func (s *ResourceServer) getUserIdFromCtx(ctx context.Context) int32 {
	return ctx.Value(consts.UserIDCtxKey).(int32)
}

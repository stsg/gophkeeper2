package interceptors

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/server/model"
	"github.com/stsg/gophkeeper2/server/model/consts"
	"github.com/stsg/gophkeeper2/server/model/errs"
	"github.com/stsg/gophkeeper2/server/services"
)

//go:generate mockgen -source=token_processing.go -destination=../mocks/interceptors/token_processing.go -package=interceptors

type RequestTokenProcessor interface {
	TokenInterceptor() grpc.UnaryServerInterceptor
	TokenStreamInterceptor() grpc.StreamServerInterceptor
}

type requestTokenProcessor struct {
	log             *zap.SugaredLogger
	tokenService    services.TokenService
	nonSecureMethod map[string]struct{}
}

func NewRequestTokenProcessor(tokenService services.TokenService, nonSecureMethods ...string) RequestTokenProcessor {
	validator := &requestTokenProcessor{
		log:             logger.NewLogger("token-itr"),
		tokenService:    tokenService,
		nonSecureMethod: make(map[string]struct{}),
	}
	for _, method := range nonSecureMethods {
		validator.nonSecureMethod[method] = struct{}{}
	}
	return validator
}

func (tp *requestTokenProcessor) isSecureMethod(method string) bool {
	if _, ok := tp.nonSecureMethod[method]; ok {
		return true
	}
	return false
}

func (tp *requestTokenProcessor) TokenInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if !tp.isSecureMethod(info.FullMethod) {
			userId, err := tp.tokenService.ExtractUserId(ctx)
			if err != nil {
				tp.log.Errorf("failed to extract userId from request token: %tp", err)
				return nil, errs.TokenError{Err: err}
			}
			ctxWithUserId := context.WithValue(ctx, consts.UserIDCtxKey, userId)
			tp.log.Infof("Retrieved from token userId: %d", userId)
			return handler(ctxWithUserId, req)
		}
		return handler(ctx, req)
	}
}

func (tp *requestTokenProcessor) TokenStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if !tp.isSecureMethod(info.FullMethod) {
			userId, err := tp.tokenService.ExtractUserId(ss.Context())
			if err != nil {
				tp.log.Errorf("failed to extract userId from request token: %tp", err)
				return errs.TokenError{Err: err}
			}
			ctxWithUserId := context.WithValue(ss.Context(), consts.UserIDCtxKey, userId)
			tp.log.Infof("Retrieved from token userId: %d", userId)
			ss.Context()
			return handler(srv, &model.ServerStreamWithCtx{
				ServerStream: ss,
				Ctx:          ctxWithUserId,
			})
		}
		return handler(srv, ss)
	}
}

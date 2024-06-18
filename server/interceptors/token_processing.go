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

// RequestTokenProcessor interface for token processing
type RequestTokenProcessor interface {
	TokenInterceptor() grpc.UnaryServerInterceptor
	TokenStreamInterceptor() grpc.StreamServerInterceptor
}

type requestTokenProcessor struct {
	log             *zap.SugaredLogger
	tokenService    services.TokenService
	nonSecureMethod map[string]struct{}
}

// NewRequestTokenProcessor creates a new instance of the RequestTokenProcessor interface.
//
// It takes a tokenService of type services.TokenService and a variadic parameter nonSecureMethods of type string.
// It returns a RequestTokenProcessor.
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

// isSecureMethod checks if the method is secure.
//
// It takes a string method as a parameter and returns a boolean.
func (tp *requestTokenProcessor) isSecureMethod(method string) bool {
	if _, ok := tp.nonSecureMethod[method]; ok {
		return true
	}
	return false
}

// TokenInterceptor returns a grpc.UnaryServerInterceptor that checks if the method is secure before extracting the userId from the request token.
//
// It takes a context.Context, an interface{} representing the request, a *grpc.UnaryServerInfo, and a grpc.UnaryHandler as parameters.
// It returns an interface{} representing the response and an error.
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

// TokenStreamInterceptor is a stream server interceptor that extracts the user ID from the request token and adds it to the context if the method is not secure.
//
// It takes the following parameters:
// - srv: the server instance
// - ss: the server stream
// - info: the stream server info
// - handler: the stream handler
//
// It returns an error.
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

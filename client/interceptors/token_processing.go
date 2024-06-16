package interceptors

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/stsg/gophkeeper2/client/model"
	"github.com/stsg/gophkeeper2/pkg/logger"
)

//go:generate mockgen -source=token_processing.go -destination=../mocks/interceptors/token_processing.go -package=interceptors

type RequestTokenProcessor interface {
	TokenInterceptor() grpc.UnaryClientInterceptor
	TokenStreamInterceptor() grpc.StreamClientInterceptor
}

type requestTokenProcessor struct {
	log         *zap.SugaredLogger
	tokenHolder *model.TokenHolder
}

func NewRequestTokenProcessor(tokenHolder *model.TokenHolder) RequestTokenProcessor {
	return &requestTokenProcessor{
		log:         logger.NewLogger("interceptors"),
		tokenHolder: tokenHolder,
	}
}

func (tp *requestTokenProcessor) TokenInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(tp.ctxWithToken(ctx), method, req, reply, cc, opts...)
	}
}

func (tp *requestTokenProcessor) TokenStreamInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		return streamer(tp.ctxWithToken(ctx), desc, cc, method, opts...)
	}
}

func (tp *requestTokenProcessor) ctxWithToken(ctx context.Context) context.Context {
	token := tp.tokenHolder.Get()
	if token != "" {
		return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"token": token}))
	}
	return ctx
}

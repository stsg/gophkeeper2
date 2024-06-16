package model

import (
	"context"

	"google.golang.org/grpc"
)

type ServerStreamWithCtx struct {
	grpc.ServerStream
	Ctx context.Context
}

func (w *ServerStreamWithCtx) Context() context.Context {
	return w.Ctx
}

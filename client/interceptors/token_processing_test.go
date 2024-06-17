// Token is successfully added to the context when present
package interceptors

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stsg/gophkeeper2/client/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestTokenInterceptor_AddsTokenToContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tokenHolder := &model.TokenHolder{}
	tokenHolder.Set("test-token")

	tp := NewRequestTokenProcessor(tokenHolder)

	interceptor := tp.TokenInterceptor()

	ctx := context.Background()
	method := "/test/method"
	req := struct{}{}
	reply := struct{}{}
	cc := &grpc.ClientConn{}

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, []string{"test-token"}, md["token"])
		return nil
	}

	err := interceptor(ctx, method, req, reply, cc, invoker)
	assert.NoError(t, err)
}

// Token is empty and context remains unchanged
func TestTokenInterceptor_EmptyTokenContextUnchanged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tokenHolder := &model.TokenHolder{}
	tokenHolder.Set("")

	tp := NewRequestTokenProcessor(tokenHolder)

	interceptor := tp.TokenInterceptor()

	ctx := context.Background()
	method := "/test/method"
	req := struct{}{}
	reply := struct{}{}
	cc := &grpc.ClientConn{}

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		assert.False(t, ok)
		assert.Nil(t, md)
		return nil
	}

	err := interceptor(ctx, method, req, reply, cc, invoker)
	assert.NoError(t, err)
}

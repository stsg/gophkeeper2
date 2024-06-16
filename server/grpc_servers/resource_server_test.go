package grpc_servers

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	intsrv "github.com/stsg/gophkeeper2/pkg/mocks/services"
	"github.com/stsg/gophkeeper2/pkg/mocks/shutdown"
	"github.com/stsg/gophkeeper2/pkg/model/enum"
	"github.com/stsg/gophkeeper2/pkg/pb"
	"github.com/stsg/gophkeeper2/server/mocks/services"
	"github.com/stsg/gophkeeper2/server/model"
	"github.com/stsg/gophkeeper2/server/model/consts"
)

// Another type of tests
func TestResourceServer(t *testing.T) {
	tests := []struct {
		name    string
		testing func(t *testing.T)
	}{
		{
			name:    "Successful save new loginPassword resource",
			testing: testResourceServerGet,
		},
	}

	for _, test := range tests {
		t.Parallel()
		t.Run(test.name, func(t *testing.T) {
			test.testing(t)
		})
	}
}

func testResourceServerGet(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	resourceService := services.NewMockResourceService(ctrl)
	fileService := intsrv.NewMockFileService(ctrl)
	exitHandler := shutdown.NewMockExitHandler(ctrl)

	resourcesServer := NewResourcesServer(resourceService, fileService, exitHandler)

	resRequest := &pb.Resource{
		Type: pb.TYPE_LOGIN_PASSWORD,
		Meta: []byte("meta"),
		Data: []byte("data"),
	}

	userId := int32(1)
	ctx = context.WithValue(ctx, consts.UserIDCtxKey, userId)
	res := &model.Resource{
		UserId: userId,
		Data:   resRequest.Data,
		ResourceDescription: model.ResourceDescription{
			Meta: resRequest.Meta,
			Type: enum.ResourceType(resRequest.Type),
		},
	}
	resId := int32(2)
	resourceService.
		EXPECT().
		Save(ctx, gomock.Eq(res)).
		Do(func(ctx context.Context, r *model.Resource) {
			r.Id = resId
		}).
		Return(nil)

	resourceId, err := resourcesServer.Save(ctx, resRequest)
	assert.NoError(t, err)
	assert.NotNil(t, resourceId)
	assert.Equal(t, resId, resourceId.Id)
}

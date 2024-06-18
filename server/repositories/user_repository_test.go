package repositories

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock "github.com/stsg/gophkeeper2/server/mocks/repositories"
)

// func init() {
// 	err := os.Chdir("../..")
// 	if err != nil {
// 		panic(err)
// 	}
// 	appConfig, err = configs.InitAppConfig(defaultConfigPath)
// 	if err != nil {
// 		panic(err)
// 	}
// }

func TestNewUserRepository_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDBProvider := mock.NewMockDBProvider(ctrl)
	repo := NewUserRepository(mockDBProvider)
	assert.NotNil(t, repo)
	assert.IsType(t, &userRepository{}, repo)

}

func TestNewUserRepository_DBProviderNil(t *testing.T) {
	repo := NewUserRepository(nil)

	assert.NotNil(t, repo)
	assert.IsType(t, &userRepository{}, repo)
}

// func TestCreateUser_Success(t *testing.T) {
// 	ctx := context.Background()

// 	assert.NotNil(t, appConfig)

// 	dbProvider, err := NewPgProvider(ctx, appConfig)
// 	assert.NoError(t, err)

// 	repo := NewUserRepository(dbProvider)
// 	user := &model.User{Username: "testuser", Password: []byte("testpass")}
// 	userID, err := repo.CreateUser(context.Background(), user)

// 	assert.NoError(t, err)
// 	assert.Equal(t, int32(1), userID)
// }

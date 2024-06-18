package repositories

import (
	"context"
	"fmt"
	"os"
	"testing"

	// "github.com/golang/mock/gomock"

	// "github.com/jackc/pgconn"
	// pgxpool "github.com/paradoxedge/pgxpoolmock"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stsg/gophkeeper2/server/configs"
)

var (
	defaultConfigPath = "cfg/config.json"
)

var (
	appConfig *configs.AppConfig
)

func init() {
	err := os.Chdir("../..")
	if err != nil {
		panic(err)
	}
	dir, _ := os.Getwd()
	fmt.Print("Current dir: ", dir)
	appConfig, err = configs.InitAppConfig(defaultConfigPath)
	if err != nil {
		panic(err)
	}
}

func TestNewPgProvider_Success(t *testing.T) {
	ctx := context.Background()

	assert.NotNil(t, appConfig)

	dbProvider, err := NewPgProvider(ctx, appConfig)

	assert.NoError(t, err)
	assert.NotNil(t, dbProvider)
}

func TestPgProvider_GetConnection(t *testing.T) {
	ctx := context.Background()

	assert.NotNil(t, appConfig)

	dbProvider, err := NewPgProvider(ctx, appConfig)

	assert.NoError(t, err)
	assert.NotNil(t, dbProvider)

	conn, err := dbProvider.GetConnection(ctx)

	assert.NotNil(t, conn)
	assert.IsType(t, &pgxpool.Conn{}, conn)
	assert.NoError(t, err)
}

func TestPgProvider_HealthCheck(t *testing.T) {
	ctx := context.Background()

	assert.NotNil(t, appConfig)

	dbProvider, err := NewPgProvider(ctx, appConfig)

	assert.NoError(t, err)
	assert.NotNil(t, dbProvider)

	err = dbProvider.HealthCheck(ctx)
	assert.NoError(t, err)
}

package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stsg/gophkeeper2/server/configs"
	"github.com/stsg/gophkeeper2/server/repositories"
)

func TestCheckDBHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	appConfig, _ := configs.InitAppConfig("cfg/config.json")
	dbProvider, _ := repositories.NewPgProvider(ctx, appConfig)
	hc := &HealthChecker{ctx: ctx, db: dbProvider}
	assert.NotNil(t, hc)
}

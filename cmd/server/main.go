package main

import (
	"context"
	"os"

	"github.com/stsg/gophkeeper2/pkg/logger"
	intsrv "github.com/stsg/gophkeeper2/pkg/services"
	"github.com/stsg/gophkeeper2/pkg/shutdown"
	"github.com/stsg/gophkeeper2/server/configs"
	servers "github.com/stsg/gophkeeper2/server/grpc_servers"
	"github.com/stsg/gophkeeper2/server/repositories"
	"github.com/stsg/gophkeeper2/server/services"
)

var (
	BuildVersion      = "N/A"
	BuildDate         = "N/A"
	configPathEnvVar  = "CONFIG"
	defaultConfigPath = "cfg/config.json"
)

func main() {
	log := logger.NewLogger("main")
	log.Infof("Build version: %s", BuildVersion)
	log.Debugf("Build date: %s", BuildDate)

	ctx, ctxCancel := context.WithCancel(context.Background())

	log.Infof("Server args: %s", os.Args[1:])
	configPath := os.Getenv(configPathEnvVar)
	if configPath == "" {
		configPath = defaultConfigPath
	}
	appConfig, err := configs.InitAppConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	dbProvider, err := repositories.NewPgProvider(ctx, appConfig)
	if err != nil {
		log.Fatalln(err)
	}
	exitHandler := shutdown.NewExitHandlerWithCtx(ctxCancel)

	userRepo := repositories.NewUserRepository(dbProvider)
	resRepo := repositories.NewResourceRepository(dbProvider)

	userSrv := services.NewUserService(userRepo)
	resSrv := services.NewResourceService(resRepo)
	tokenSrv := services.NewTokenService(appConfig.TokenKey)
	fileProcessor := intsrv.NewFileService()

	authServer := servers.NewAuthServer(userSrv, tokenSrv)
	resourcesServer := servers.NewResourcesServer(resSrv, fileProcessor, exitHandler)

	serverManager, err := servers.NewServerManager(tokenSrv)
	if err != nil {
		log.Fatalf("failed to init grpc server: %v", err)
	}
	serverManager.RegisterResourcesServer(resourcesServer)
	serverManager.RegisterAuthServer(authServer)
	server, err := serverManager.Start(appConfig.ServerPort)
	if err != nil {
		log.Fatalf("failed to start grpc server: %v", err)
	}
	exitHandler.ShutdownGrpcServerBeforeExit(server)
	exit := exitHandler.ProperExitDefer()
	<-exit
	log.Info("program is going to be closed")
	<-ctx.Done()
}

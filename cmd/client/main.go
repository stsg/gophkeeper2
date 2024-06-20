package main

import (
	"context"
	"io"
	"os"

	clients "github.com/stsg/gophkeeper2/client"
	"github.com/stsg/gophkeeper2/client/configs"
	"github.com/stsg/gophkeeper2/client/model"
	"github.com/stsg/gophkeeper2/client/services"
	"github.com/stsg/gophkeeper2/client/terminal"
	"github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/pkg/pb"
	intsrv "github.com/stsg/gophkeeper2/pkg/services"
	"github.com/stsg/gophkeeper2/pkg/shutdown"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"

	configPathEnvVar  = "CONFIG"
	defaultConfigPath = "cfg/gpk_config.json"
)

func main() {
	log := logger.NewLogger("main")
	log.Infof("Client args: %s", os.Args[1:])
	ctx, ctxClose := context.WithCancel(context.Background())
	configPath := os.Getenv(configPathEnvVar)
	if configPath == "" {
		configPath = defaultConfigPath
	}
	appConfig, err := configs.InitAppConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	exitHandler := shutdown.NewExitHandlerWithCtx(ctxClose)

	tokenHolder := &model.TokenHolder{}

	grpcConn, err := clients.CreateGrpcConnection(appConfig.ServerPort, tokenHolder)
	if err != nil {
		log.Fatalf("failed to create grpc connection: %v", err)
	}
	exitHandler.ToClose([]io.Closer{grpcConn})

	authService := services.NewAuthService(pb.NewAuthClient(grpcConn), tokenHolder)
	fileService := intsrv.NewFileService()
	cryptoService := services.NewCryptService(appConfig.PrivateKey)
	resourceService := services.NewResourceService(pb.NewResourcesClient(grpcConn), fileService, cryptoService)
	exit := exitHandler.ProperExitDefer()

	commandProcessor := terminal.NewCommandParser(buildVersion, buildDate, authService, resourceService, exitHandler, exit)
	commandProcessor.InitScanner()
	commandProcessor.Start(exit)
	<-ctx.Done()
}

package grpc_servers

import (
	"net"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/pkg/pb"
	"github.com/stsg/gophkeeper2/server/interceptors"
	"github.com/stsg/gophkeeper2/server/services"
)

const (
	registerMethod = "/gophkeeper.Auth/Register"
	loginMethod    = "/gophkeeper.Auth/Login"
	serverCert     = "cert/server-cert.pem"
	serverKey      = "cert/server-key.pem"
)

//go:generate mockgen -source=server_manager.go -destination=../mocks/grpc_servers/server_manager.go -package=grpc_servers

type ServerManager interface {
	RegisterAuthServer(authServer pb.AuthServer)
	RegisterResourcesServer(resServer pb.ResourcesServer)
	Start(port string) (*grpc.Server, error)
}

type serverManager struct {
	log    *zap.SugaredLogger
	server *grpc.Server
}

func NewServerManager(tokenService services.TokenService) (ServerManager, error) {
	sm := &serverManager{log: logger.NewLogger("server-mngr")}
	tokenValidator := interceptors.NewRequestTokenProcessor(tokenService, registerMethod, loginMethod)
	tlsCredentials, err := sm.loadTLSCredentials()
	if err != nil {
		return nil, err
	}
	server := grpc.NewServer(
		grpc.Creds(tlsCredentials),
		grpc.UnaryInterceptor(tokenValidator.TokenInterceptor()),
		grpc.StreamInterceptor(tokenValidator.TokenStreamInterceptor()),
	)
	sm.server = server
	return sm, nil
}

func (s *serverManager) Start(port string) (*grpc.Server, error) {
	listen, err := net.Listen("tcp", port)
	if err != nil {
		s.log.Errorf("failed to listen '%s': %v", port, err)
		return nil, err
	}

	go func() {
		s.log.Infof("proto server started on %s", port)
		if err := s.server.Serve(listen); err != nil {
			s.log.Fatalf("failed to start server: %v", err)
		}
	}()
	return s.server, nil
}

func (s *serverManager) RegisterAuthServer(authServer pb.AuthServer) {
	pb.RegisterAuthServer(s.server, authServer)
}

func (s *serverManager) RegisterResourcesServer(resServer pb.ResourcesServer) {
	pb.RegisterResourcesServer(s.server, resServer)
}

func (s *serverManager) loadTLSCredentials() (credentials.TransportCredentials, error) {
	config, err := credentials.NewServerTLSFromFile(serverCert, serverKey)
	if err != nil {
		s.log.Errorf("failed to load TLC config: %v", err)
		return nil, errors.Wrap(err, "tls-error")
	}

	return config, nil
}

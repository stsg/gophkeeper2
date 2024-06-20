package clients

import (
	"log"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/stsg/gophkeeper2/client/interceptors"
	"github.com/stsg/gophkeeper2/client/model"
)

const (
	serverCert = "cert/server-cert.pem"
	serverName = ""
)

func CreateGrpcConnection(targetPort string, tokenHolder *model.TokenHolder) (*grpc.ClientConn, error) {
	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("cannot load TLS credentials: ", err)
		return nil, err
	}
	tokenProcessor := interceptors.NewRequestTokenProcessor(tokenHolder)
	return grpc.NewClient(
		targetPort,
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithUnaryInterceptor(tokenProcessor.TokenInterceptor()),
		grpc.WithStreamInterceptor(tokenProcessor.TokenStreamInterceptor()),
	)
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	creds, err := credentials.NewClientTLSFromFile(serverCert, serverName)
	if err != nil {
		return nil, errors.Wrap(err, "tls-error")
	}

	return creds, nil
}

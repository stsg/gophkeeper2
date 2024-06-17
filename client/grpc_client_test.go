package clients

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stsg/gophkeeper2/client/model"
	"google.golang.org/grpc/credentials"
)

// Successfully creates a gRPC connection with valid TLS credentials
func TestCreateGrpcConnection_Success(t *testing.T) {

	err := os.Chdir("..")
	if err != nil {
		panic(err)
	}

	// Arrange
	targetPort := "localhost:50051"
	tokenHolder := &model.TokenHolder{}

	// Act
	conn, err := CreateGrpcConnection(targetPort, tokenHolder)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, conn)
}

// Fails to create a connection if TLS credentials cannot be loaded
func TestCreateGrpcConnection_FailTLSLoad(t *testing.T) {
	originalLoadTLSCredentials := loadTLSCredentials
	loadTLSCredentials := func() (credentials.TransportCredentials, error) {
		return nil, errors.New("tls-error")
	}
	defer func() { loadTLSCredentials = originalLoadTLSCredentials }()

	// Act
	credentials, err := loadTLSCredentials()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, credentials)
}

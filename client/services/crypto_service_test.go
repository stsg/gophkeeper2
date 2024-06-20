package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCryptServiceWithValidPrivateKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	service := NewCryptService(privateKey)

	assert.NotNil(t, service)
	assert.Equal(t, privateKey, service.(*cryptService).privateKey)
}

func TestNewCryptServiceWithNilPrivateKey(t *testing.T) {
	service := NewCryptService(nil)

	assert.NotNil(t, service)
	assert.Nil(t, service.(*cryptService).privateKey)
}

func TestNewCryptServiceReturnsNonNilInstance(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	service := NewCryptService(privateKey)

	assert.NotNil(t, service)
	assert.Equal(t, privateKey, service.(*cryptService).privateKey)
}

func TestNewCryptServiceWithInvalidPrivateKey(t *testing.T) {
	privateKey := &rsa.PrivateKey{} // Creating an empty or invalid private key

	service := NewCryptService(privateKey)

	assert.NotNil(t, service)
	assert.Equal(t, privateKey, service.(*cryptService).privateKey)
}

func TestDecryptWithValidPrivateKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	cryptService := NewCryptService(privateKey)

	originalData := []byte("test data")
	encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &privateKey.PublicKey, originalData, []byte("cartople"))
	assert.NoError(t, err)

	decryptedData, err := cryptService.Decrypt(encryptedData)
	assert.NoError(t, err)
	assert.Equal(t, originalData, decryptedData)
}

func TestDecryptWithEmptyData(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	cryptService := NewCryptService(privateKey)

	emptyData := []byte{}
	decryptedData, err := cryptService.Decrypt(emptyData)
	assert.NoError(t, err)
	assert.Equal(t, emptyData, decryptedData)
}

func TestErrorOnDecryptionFailure(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	cryptService := NewCryptService(privateKey)

	// Create invalid encrypted data
	encryptedData := []byte("invalid data")

	// Ensure error is returned on decryption failure
	decryptedData, err := cryptService.Decrypt(encryptedData)
	assert.Error(t, err)
	assert.Nil(t, decryptedData)
}

func TestDecryptWithInvalidPrivateKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	cryptService := NewCryptService(privateKey)

	encryptedData := []byte("$2a$08$fIPQ0m1fzdkx8CQvuQB/AuCD6uixrlwQDI6EqCN1bmZX2pvbYyG26")
	_, err = cryptService.Decrypt(encryptedData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decrypt data")
}

func TestDecryptHandlesDataSmallerThanOneBlockSize(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	cryptService := NewCryptService(privateKey)

	originalData := []byte("test data")
	encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &privateKey.PublicKey, originalData, []byte("cartople"))
	assert.NoError(t, err)

	decryptedData, err := cryptService.Decrypt(encryptedData)
	assert.NoError(t, err)
	assert.Equal(t, originalData, decryptedData)
}

func TestEnsureNoDataLossDuringDecryption(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	cryptService := NewCryptService(privateKey)

	originalData := []byte("test data")
	encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &privateKey.PublicKey, originalData, []byte("cartople"))
	assert.NoError(t, err)

	decryptedData, err := cryptService.Decrypt(encryptedData)
	assert.NoError(t, err)
	assert.Equal(t, originalData, decryptedData)
}

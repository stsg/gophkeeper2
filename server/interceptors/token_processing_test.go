package interceptors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stsg/gophkeeper2/server/mocks/services"
)

// Creates a requestTokenProcessor with a valid tokenService
func TestNewRequestTokenProcessor(t *testing.T) {
	mockTokenService := new(services.MockTokenService)
	nonSecureMethods := []string{"GET", "POST"}

	processor := NewRequestTokenProcessor(mockTokenService, nonSecureMethods...)

	assert.NotNil(t, processor)
	assert.NotNil(t, processor.TokenInterceptor())
	assert.NotNil(t, processor.TokenStreamInterceptor())
}

// Handles nil tokenService without panicking
func TestNewRequestTokenProcessor_NilTokenService(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked with nil tokenService")
		}
	}()

	nonSecureMethods := []string{}

	processor := NewRequestTokenProcessor(nil, nonSecureMethods...)
	assert.NotNil(t, processor)
}

func TestIsSecureMethodInMap(t *testing.T) {
	mockTokenService := new(services.MockTokenService)
	nonSecureMethods := []string{
		"/test.Method",
		"GET",
		"POST",
	}
	processor := NewRequestTokenProcessor(mockTokenService, nonSecureMethods...)

	result := processor.(*requestTokenProcessor).isSecureMethod("/test.Method")
	assert.True(t, result)
	result = processor.(*requestTokenProcessor).isSecureMethod("GET")
	assert.True(t, result)
	result = processor.(*requestTokenProcessor).isSecureMethod("POST")
	assert.True(t, result)
	result = processor.(*requestTokenProcessor).isSecureMethod("POSt")
	assert.False(t, result)
	result = processor.(*requestTokenProcessor).isSecureMethod("test.method")
	assert.False(t, result)
}

package services

// Generates a valid JWT token with correct user ID

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stsg/gophkeeper2/server/model"
)

func TestGenerateValidJWTToken(t *testing.T) {
	key := "test_secret_key"
	service := NewTokenService(key)
	userID := int32(12345)
	expireAt := time.Now().Add(time.Hour)

	tokenStr, err := service.Generate(userID, expireAt)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	token, err := jwt.ParseWithClaims(tokenStr, &model.AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*model.AuthClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims.Id)
}

// Handles very large user IDs without error
func TestGenerateWithLargeUserID(t *testing.T) {
	key := "test_secret_key"
	service := NewTokenService(key)
	largeUserID := int32(2147483647) // Max int32 value
	expireAt := time.Now().Add(time.Hour)

	tokenStr, err := service.Generate(largeUserID, expireAt)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	token, err := jwt.ParseWithClaims(tokenStr, &model.AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*model.AuthClaims)
	assert.True(t, ok)
	assert.Equal(t, largeUserID, claims.Id)
}

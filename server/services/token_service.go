package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/metadata"

	"github.com/stsg/gophkeeper2/server/model"
	"github.com/stsg/gophkeeper2/server/model/errs"
)

const (
	token = "token"
)

//go:generate mockgen -source=token_service.go -destination=../mocks/services/token_service.go -package=services

type TokenService interface {
	Generate(id int32, expireAt time.Time) (string, error)
	ExtractUserId(ctx context.Context) (int32, error)
}

type tokenService struct {
	key string
}

func NewTokenService(key string) TokenService {
	return &tokenService{key}
}

func (s *tokenService) Generate(id int32, expireAt time.Time) (string, error) {
	claims := &model.AuthClaims{
		Id:               id,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expireAt)},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.key))
}

func (s *tokenService) ExtractUserId(ctx context.Context) (int32, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, errors.New("failed to read request metadata")
	}
	var tokenStr string
	if values := md.Get(token); len(values) == 0 {
		return 0, errs.TokenError{Err: errs.ErrTokenNotFound}
	} else {
		tokenStr = values[0]
	}

	return s.extract(tokenStr)
}

func (s *tokenService) extract(tokenStr string) (int32, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &model.AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.key), nil
	})

	if claims, ok := token.Claims.(*model.AuthClaims); ok && token.Valid {
		return claims.Id, nil
	}
	if !token.Valid {
		return 0, errs.TokenError{Err: errs.ErrTokenInvalid}
	}
	return 0, err
}

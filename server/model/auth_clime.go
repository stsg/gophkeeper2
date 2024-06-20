package model

import "github.com/golang-jwt/jwt/v4"

type AuthClaims struct {
	Id int32 `json:"id"`
	jwt.RegisteredClaims
}

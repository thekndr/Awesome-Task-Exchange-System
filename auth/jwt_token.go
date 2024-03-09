package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type JwtToken struct {
	Secret string
	TTL    time.Duration
}

func (jt JwtToken) Issue(publicId, email, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    publicId,
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(jt.TTL).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jt.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

type validatedToken struct {
	Email string
	Role  string
}

func (jt JwtToken) Validate(token string) (bool, validatedToken, error) {
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(jt.Secret), nil
	})
	if err != nil {
		return false, validatedToken{}, fmt.Errorf("Token parse failed: %w", err)
	}

	if claims, ok := parsed.Claims.(jwt.MapClaims); ok && parsed.Valid {
		email := claims["email"].(string)
		role := claims["role"].(string)
		return true, validatedToken{Email: email, Role: role}, nil
	}

	return false, validatedToken{}, fmt.Errorf("Token is invalid")
}

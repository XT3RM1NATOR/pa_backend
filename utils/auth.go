package utils

import (
	"fmt"
	"github.com/Point-AI/backend/internal/auth/domain/entity"
	"github.com/Point-AI/backend/internal/auth/infrastructure/model"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func CreateAccessToken(user *model.User, secret string, expiry int) (accessToken string, err error) {
	// Calculate token expiration time
	expirationTime := time.Now().Add(time.Hour * time.Duration(expiry))

	// Create JWT claims
	claims := &entity.JwtCustomClaims{
		ID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{
				Time: expirationTime,
			},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err = token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func CreateRefreshToken(user *model.User, secret string, expiry int) (refreshToken string, err error) {
	expirationTime := time.Now().Add(time.Hour * time.Duration(expiry))

	claimsRefresh := &entity.JwtCustomRefreshClaims{
		ID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{
				Time: expirationTime,
			},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsRefresh)

	refreshToken, err = token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func IsAuthorized(requestToken string, secret string) (bool, error) {
	_, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func ExtractIDFromToken(requestToken string, secret string) (string, error) {
	token, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok && !token.Valid {
		return "", fmt.Errorf("Invalid Token")
	}

	return claims["id"].(string), nil
}

package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

func GenerateJWTToken(id int, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateJWTToken(signedToken string, secretKey string) (int, error) {
	parsedToken, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return -1, err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		id := claims["id"].(int)
		return id, nil
	}
	return -1, errors.New("invalid token")
}

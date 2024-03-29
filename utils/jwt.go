package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type TokenType string

const (
	AccessToken  TokenType = "access_token"
	RefreshToken TokenType = "refresh_token"
	ResetToken   TokenType = "reset_token"
)

func GenerateJWTToken(tokenType TokenType, id primitive.ObjectID, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"id":   id,
		"type": tokenType,
	}

	switch tokenType {
	case AccessToken:
		claims["exp"] = time.Now().Add(time.Minute * 90).Unix()
	case RefreshToken:
		claims["exp"] = time.Now().Add(time.Hour * 24 * 90).Unix()
	case ResetToken:
		claims["exp"] = time.Now().Add(time.Minute * 60).Unix()
	default:
		return "", errors.New("wrong token type")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateJWTToken(expectedTokenType TokenType, signedToken, secretKey string) (primitive.ObjectID, error) { // Changed return type
	parsedToken, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return primitive.ObjectID{}, err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		tokenType, typeExists := claims["type"].(string)
		if !typeExists || TokenType(tokenType) != expectedTokenType {
			return primitive.ObjectID{}, errors.New("invalid token type")
		}
		expFloat, exists := claims["exp"].(float64)
		if !exists {
			return primitive.ObjectID{}, errors.New("expiration claim missing or invalid")
		}

		expTime := time.Unix(int64(expFloat), 0)
		if expTime.Before(time.Now()) {
			return primitive.ObjectID{}, errors.New("token expired")
		}

		idStr, idExists := claims["id"].(string)
		if !idExists {
			return primitive.ObjectID{}, errors.New("id field missing or invalid")
		}

		objectID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return primitive.ObjectID{}, errors.New("invalid id format")
		}

		return objectID, nil
	}

	return primitive.ObjectID{}, errors.New("invalid token")
}

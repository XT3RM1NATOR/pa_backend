package middleware

import (
	"context"
	"github.com/Point-AI/backend/utils"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

func ValidateAccessTokenMiddleware(secretKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Authorization header required"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Invalid authorization header format"})
			}

			token := parts[1]

			userId, err := utils.ValidateJWTToken(utils.AccessToken, token, secretKey)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			}

			ctx := context.WithValue(c.Request().Context(), "userId", userId)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func ValidateServerMiddleware(secretKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Authorization header required"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid authorization header format"})
			}

			key := parts[1]

			if key != secretKey {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization token"})
			}

			return next(c)
		}
	}
}

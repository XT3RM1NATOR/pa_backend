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
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Authorization header required"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid authorization header format"})
			}

			accessToken := parts[1]

			userId, err := utils.ValidateJWTToken("access_token", accessToken, secretKey)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}

			ctx := context.WithValue(c.Request().Context(), "userId", userId)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

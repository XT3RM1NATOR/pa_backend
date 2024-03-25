package middleware

import (
	"context"
	"github.com/Point-AI/backend/internal/auth/delivery/model"
	"github.com/Point-AI/backend/utils"
	"github.com/labstack/echo/v4"
	"net/http"
)

func ValidateAccessTokenMiddleware(secretKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var accessTokenInput model.AccessTokenInput
			if err := c.Bind(&accessTokenInput); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request format"})
			}

			userId, err := utils.ValidateJWTToken("access_token", accessTokenInput.AccessToken, secretKey)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}

			ctx := context.WithValue(c.Request().Context(), "userId", userId)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

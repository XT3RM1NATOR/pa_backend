package authDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterAuthRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database) {
	authGroup := e.Group("/auth")

	authGroup.POST("/signin", signInHandler)
	authGroup.POST("/signup", signUpHandler)
	authGroup.POST("/verify", verifyHandler)
	authGroup.POST("/recover", recoverHandler)
}

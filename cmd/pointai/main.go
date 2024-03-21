package main

import (
	"github.com/Point-AI/backend/config"
	app "github.com/Point-AI/backend/internal/app/db"
	"github.com/Point-AI/backend/internal/app/server"
)

func main() {
	cfg := config.Load()
	db := app.ConnectToDB(cfg)

	server.RunHTTPServer(cfg, db)
}

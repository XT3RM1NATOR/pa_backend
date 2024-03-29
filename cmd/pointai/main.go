package main

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/app/db"
	"github.com/Point-AI/backend/internal/app/server"
)

func main() {
	cfg := config.Load()
	mongodb := db.ConnectToDB(cfg)

	server.RunHTTPServer(cfg, mongodb)
}

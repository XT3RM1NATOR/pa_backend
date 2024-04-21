package main

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/app/db"
	"github.com/Point-AI/backend/internal/app/server"
	"github.com/Point-AI/backend/internal/app/storage"
)

func main() {
	cfg := config.Load()
	mongodb := db.ConnectToDB(cfg)
	str := storage.ConnectToStorage(cfg)

	server.RunHTTPServer(cfg, mongodb, str)
}

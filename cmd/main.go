package main

import (
	"fmt"
	"server-monitoring/internal/routes"
	"server-monitoring/pkg/config"
	"server-monitoring/pkg/mongo"
)

func main() {
	cfg := config.Load()

	mongo.Init(cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password)

	srv := routes.Setup()
	srv.Run(fmt.Sprintf(":%d", cfg.Port))
}

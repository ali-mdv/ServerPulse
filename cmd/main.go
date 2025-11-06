package main

import (
	"fmt"
	"server-monitoring/internal/routes"
	"server-monitoring/pkg/config"
)

func main() {
	cfg := config.Load()

	srv := routes.Setup()
	srv.Run(fmt.Sprintf(":%d", cfg.Port))
}

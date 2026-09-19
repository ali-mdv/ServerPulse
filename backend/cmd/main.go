package main

import (
	"context"
	"fmt"
	"server-monitoring/internal/routes"
	v1 "server-monitoring/internal/routes/v1"
	"server-monitoring/internal/services"
	"server-monitoring/pkg/config"
	database "server-monitoring/pkg/mongo"
)

const dbName = "server_monitoring"

func main() {
	cfg := config.Load()

	database.Init(cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password)

	pm2Service := services.NewPM2Service(cfg.PM2SocketPath)
	dockerService := services.NewDockerService()
	systemService := services.NewSystemService()
	historyService, err := services.NewHistoryService(dbName, cfg.HistoryRetention)
	if err != nil {
		panic(fmt.Sprintf("history service: %v", err))
	}
	stateService := services.NewStateService(dbName)

	scheduler := services.NewHistoryScheduler(pm2Service, dockerService, systemService, historyService, stateService, cfg.HistoryPollInterval)
	scheduler.Start(context.Background())

	srv := routes.Setup(cfg, &v1.Services{
		PM2:     pm2Service,
		Docker:  dockerService,
		System:  systemService,
		History: historyService,
		State:   stateService,
	})
	srv.Run(fmt.Sprintf(":%d", cfg.Port))
}

package main

import (
	"context"
	"fmt"
	"log"
	"server-monitoring/internal/repository"
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
	db := database.GetDatabase(dbName)

	pm2Service := services.NewPM2Service(cfg.PM2SocketPath)
	dockerService := services.NewDockerService()
	systemService := services.NewSystemService()
	historyService, err := services.NewHistoryService(dbName, cfg.HistoryRetention)
	if err != nil {
		panic(fmt.Sprintf("history service: %v", err))
	}
	stateService := services.NewStateService(dbName)
	serverService := services.NewServerService(dbName)

	// AgentService needs the raw server repo so it can stamp heartbeats
	// without the serverService having to expose it.
	agentService := services.NewAgentService(
		repository.NewServerRepository(db),
		stateService,
		historyService,
	)

	// Ensure the local host exists as a server record so the in-process
	// scheduler's LocalServerID matches a real row. Best-effort — a
	// failure here is logged but does not abort startup so the rest of
	// the dashboard stays usable.
	if _, err := serverService.EnsureLocalServer(context.Background()); err != nil {
		log.Printf("server: ensure local server failed: %v", err)
	}

	scheduler := services.NewHistoryScheduler(pm2Service, dockerService, systemService, historyService, stateService, cfg.HistoryPollInterval)
	scheduler.Start(context.Background())

	srv := routes.Setup(cfg, &v1.Services{
		PM2:     pm2Service,
		Docker:  dockerService,
		System:  systemService,
		History: historyService,
		State:   stateService,
		Servers: serverService,
		Agent:   agentService,
	})
	srv.Run(fmt.Sprintf(":%d", cfg.Port))
}
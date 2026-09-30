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

	"go.mongodb.org/mongo-driver/v2/bson"
)

const dbName = "server_monitoring"

func main() {
	cfg := config.Load()

	database.Init(cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password)
	db := database.GetDatabase(dbName)

	pm2Service := services.NewPM2Service(cfg.PM2SocketPath)
	dockerService := services.NewDockerService()
	systemService := services.NewSystemService()
	settingsService := services.NewSettingsService(dbName)
	historyService := services.NewHistoryService(dbName, settingsService)
	stateService := services.NewStateService(dbName)
	serverService := services.NewServerService(dbName)
	notificationService := services.NewNotificationService(dbName)

	// Control channel: agents connect over Socket.IO; user actions for
	// remote servers are forwarded through the hub and wait for an ack.
	agentHub := services.NewAgentHub()
	agentSocket := services.NewAgentSocket(serverService, agentHub)
	commandService := services.NewCommandService(agentHub, dockerService, pm2Service, serverService)

	// Best-effort index creation for the notification lookup pattern.
	if err := notificationService.EnsureSchema(context.Background()); err != nil {
		log.Printf("notifications: ensure schema failed: %v", err)
	}

	// AgentService needs the raw server repo so it can stamp heartbeats
	// without the serverService having to expose it.
	agentService := services.NewAgentService(
		repository.NewServerRepository(db),
		stateService,
		historyService,
		notificationService,
	)

	// Ensure the local host exists as a server record so the in-process
	// scheduler writes state under the same server ID the dashboard uses.
	// Best-effort — a failure here is logged but does not abort startup.
	localServerID := bson.NewObjectID()
	if localServer, err := serverService.EnsureLocalServer(context.Background()); err != nil {
		log.Printf("server: ensure local server failed: %v", err)
	} else if localServer != nil {
		localServerID = localServer.ID
	}

	scheduler := services.NewHistoryScheduler(pm2Service, dockerService, systemService, historyService, stateService, serverService, notificationService, settingsService, localServerID)
	scheduler.Start(context.Background())

	srv := routes.Setup(cfg, &v1.Services{
		PM2:           pm2Service,
		Docker:        dockerService,
		System:        systemService,
		History:       historyService,
		State:         stateService,
		Servers:       serverService,
		Agent:         agentService,
		Notifications: notificationService,
		Settings:      settingsService,
		Commands:      commandService,
		AgentSocket:   agentSocket,
	})
	srv.Run(fmt.Sprintf(":%d", cfg.Port))
}
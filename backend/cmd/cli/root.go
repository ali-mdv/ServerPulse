package main

import (
	"os"

	"server-monitoring/internal/services"
	"server-monitoring/pkg/config"
	database "server-monitoring/pkg/mongo"

	"github.com/spf13/cobra"
)

const dbName = "server_monitoring"

type app struct {
	users services.UserService
}

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	a := &app{}

	cmd := &cobra.Command{
		Use:          "serverpulse-cli",
		Short:        "ServerPulse management CLI",
		Long:         "ServerPulse CLI for managing backend resources such as users.",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// The server config requires PORT even though the CLI never
			// listens; supply a harmless default when it is unset so the
			// shared config loader can be used.
			if os.Getenv("PORT") == "" {
				_ = os.Setenv("PORT", "12000")
			}

			cfg := config.Load()
			database.Init(cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password)
			a.users = services.NewUserService(dbName)
			return nil
		},
	}

	cmd.AddCommand(newUserCmd(a))
	return cmd
}

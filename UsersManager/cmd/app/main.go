package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"usersmanager/internal/app"
	userspsqlstorage "usersmanager/internal/storage/users/psql"
	"usersmanager/pkg/config"
	"usersmanager/pkg/lib/logger"
)

func main() {
	config := config.MustLoad()

	log := logger.SetupLogger(config.Env)

	log.Info("application", slog.Any("config", config))

	psqlStorage := userspsqlstorage.New(log, config.PsqlConnStr, config.PsqlUsersTableName)

	application := app.New(log, config.Port, psqlStorage)

	go func() {
		application.GRPCApp.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	psqlStorage.Close()
	application.GRPCApp.Stop()
}

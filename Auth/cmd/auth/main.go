package main

import (
	"auth/pkg/config"
	"auth/pkg/lib/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("application", slog.Any("config", cfg))

	// usersStorage := usersgrpcstorage.New(log, cfg.UsersGrpcStorageHost, cfg.cfg.UsersGrpcStoragePort)

	// application := app.New(log, cfg.Port, usersStorage)

	/*
		go func() {
			application.GRPCApp.MustRun()
		}())
	*/

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	// application.GRPCApp.Stop()

	// usersStorage.Close()
}

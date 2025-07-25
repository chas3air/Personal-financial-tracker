package main

import (
	"apigateway/internal/app"
	usersgrpcstorage "apigateway/internal/storage/users/grpc"
	"apigateway/pkg/config"
	"apigateway/pkg/lib/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("application config", slog.Any("config", cfg))

	storage := usersgrpcstorage.New(log, cfg.UsersStorageHost, cfg.UsersStoragePort)

	application := app.New(log, cfg.Port, storage)

	go func() {
		application.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	storage.Close()

}

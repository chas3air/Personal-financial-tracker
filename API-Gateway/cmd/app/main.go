package main

import (
	"apigateway/internal/app"
	"apigateway/pkg/config"
	"apigateway/pkg/lib/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoadEnv()

	log := logger.SetupLogger(cfg.Env)

	log.Info("application config", slog.Any("config", cfg))

	application := app.New(log, cfg.Port, nil)

	go func() {
		application.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<- stop

	// storage.Close()

	// application.
}

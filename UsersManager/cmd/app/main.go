package main

import (
	"usersmanager/pkg/config"
	"usersmanager/pkg/lib/logger"
)

func main() {
	config := config.MustLoad()

	log := logger.SetupLogger(config.Env)

	_ = log
}

package main

import (
	"net/http"
	"os"

	"filedownloader-20240926/internal/config"
	"filedownloader-20240926/internal/handler"
	"filedownloader-20240926/internal/service"
	"filedownloader-20240926/pkg/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	setupLogging(cfg)

	logger.Logger.Info("Starting File Downloader Service",
		"server_port", cfg.Server.Port,
		"worker_count", cfg.Worker.Count,
		"debug_mode", cfg.IsDebugMode())

	logger.Logger.Info("Initializing components")
	taskManager := service.NewTaskManager()
	workerPool := service.NewWorkerPool(cfg.Worker.Count, taskManager)
	workerPool.Start()

	logger.Logger.Info("Recovering incomplete tasks")
	taskManager.RecoverIncompleteTasks()

	logger.Logger.Info("Setting up HTTP server")
	th := handler.NewTaskHandler(taskManager, workerPool)
	server := &http.Server{
		Addr:    cfg.GetServerAddr(),
		Handler: handler.SetupRoutes(th),
	}

	logger.Logger.Info("Setting up graceful shutdown")
	graceful := service.NewGracefulShutdown(server, workerPool, taskManager)

	logger.Logger.Info("Server starting", "addr", cfg.GetServerAddr())
	if err := graceful.Start(); err != nil {
		logger.Logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

// setupLogging configures logging based on the configuration
func setupLogging(cfg *config.Config) {
	if cfg.IsDebugMode() {
		logger.SetDebug()
	} else {
		logger.SetProduction()
	}

	logger.Logger.Info("Configuration loaded",
		"config_file", "config.yaml",
		"server_port", cfg.Server.Port,
		"worker_count", cfg.Worker.Count,
		"log_level", cfg.Logging.Level,
		"log_format", cfg.Logging.Format)
}

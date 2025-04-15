package main

import (
	"log/slog"
	"os"

	"github.com/kaustavdm/awwdio/config"
	"github.com/kaustavdm/awwdio/server"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Load configuration using config.LoadConfig
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create a new server instance
	srv := server.NewServer(cfg, logger)

	// Start the server
	logger.Info("Server starting...")
	if err := srv.Start(); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

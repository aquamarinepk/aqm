package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aquamarinepk/aqm/examples/ticked/services/authn/config"
	"github.com/aquamarinepk/aqm/examples/ticked/services/authn/internal"
	log "github.com/aquamarinepk/aqm/log"
)

func main() {
	// Create logger
	logger := log.NewLogger("info")

	// Load configuration
	logger.Info("Loading configuration...")
	cfg, err := config.New(logger)
	if err != nil {
		logger.Errorf("Failed to load config: %v", err)
		os.Exit(1)
	}

	// Create service
	logger.Info("Creating service...")
	svc, err := internal.New(cfg)
	if err != nil {
		logger.Errorf("Failed to create service: %v", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start service in background
	errCh := make(chan error, 1)
	go func() {
		if err := svc.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	logger.Info("Service started. Press Ctrl+C to stop.")
	logger.Infof("API available at http://localhost%s", cfg.Server.Port)

	// Wait for shutdown signal or error
	select {
	case sig := <-sigCh:
		logger.Infof("Received signal: %v", sig)
	case err := <-errCh:
		logger.Errorf("Service error: %v", err)
	}

	// Graceful shutdown
	logger.Info("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := svc.Stop(shutdownCtx); err != nil {
		logger.Errorf("Shutdown error: %v", err)
		os.Exit(1)
	}

	logger.Info("Service stopped")
}

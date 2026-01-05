package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aquamarinepk/aqm/examples/ticked/services/authn/config"
	"github.com/aquamarinepk/aqm/examples/ticked/services/authn/internal"
)

func main() {
	// Load configuration
	log.Println("Loading configuration...")
	cfg, err := config.Load("AUTHN_")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config loaded: driver=%s, port=%s, log=%s",
		cfg.Database.Driver, cfg.Server.Port, cfg.Log.Level)

	// Create service
	log.Println("Creating service...")
	svc, err := internal.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
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

	log.Println("Service started. Press Ctrl+C to stop.")
	log.Printf("API available at http://localhost%s", cfg.Server.Port)

	// Wait for shutdown signal or error
	select {
	case sig := <-sigCh:
		log.Printf("Received signal: %v", sig)
	case err := <-errCh:
		log.Printf("Service error: %v", err)
	}

	// Graceful shutdown
	log.Println("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := svc.Stop(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	log.Println("Service stopped")
}

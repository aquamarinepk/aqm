package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aquamarinepk/aqm/app"
	"github.com/aquamarinepk/aqm/examples/ticked/services/authn/config"
	"github.com/aquamarinepk/aqm/examples/ticked/services/authn/internal"
	logger "github.com/aquamarinepk/aqm/log"
)

const (
	name    = "authn"
	version = "0.1.0"
)

func main() {
	// Create logger
	log := logger.NewLogger("info")

	// Load configuration
	cfg, err := config.New(log)
	if err != nil {
		log.Errorf("Cannot load config: %v", err)
		os.Exit(1)
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create router
	router := app.NewRouter(log, app.WithPing(), app.WithDebugRoutes())

	// Build dependencies
	var deps []any

	svc, err := internal.New(cfg)
	if err != nil {
		log.Errorf("Cannot create service: %v", err)
		os.Exit(1)
	}

	deps = append(deps, svc)

	// Setup lifecycle
	starts, stops, registrars := app.Setup(ctx, router, deps...)

	// Start components
	if err := app.Start(ctx, log, starts, stops, registrars, router); err != nil {
		log.Errorf("Cannot start %s(%s): %v", name, version, err)
		os.Exit(1)
	}

	log.Infof("%s(%s) started successfully", name, version)

	// Start HTTP server in background
	go func() {
		log.Infof("Server listening on %s", cfg.Server.Port)
		if err := app.Serve(router, cfg.Server.Port); err != nil {
			log.Errorf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop

	// Graceful shutdown
	log.Infof("Shutting down %s(%s)...", name, version)
	cancel()

	for i := len(stops) - 1; i >= 0; i-- {
		if err := stops[i](context.Background()); err != nil {
			log.Errorf("Error stopping component: %v", err)
		}
	}

	fmt.Println("Goodbye!")
}

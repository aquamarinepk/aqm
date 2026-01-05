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
	"github.com/aquamarinepk/aqm/log"
)

const (
	name    = "authn"
	version = "0.1.0"
)

func main() {
	logger := log.NewLogger("info")

	cfg, err := config.New(logger)
	if err != nil {
		logger.Errorf("Cannot load config: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := app.NewRouter(logger)
	app.ApplyRouterOptions(router,
		app.WithDefaultInternalMiddlewares(),
		app.WithPing(),
		app.WithDebugRoutes(),
		app.WithHealthChecks(name, version),
	)

	var deps []any

	svc, err := internal.New(cfg)
	if err != nil {
		logger.Errorf("Cannot create service: %v", err)
		os.Exit(1)
	}

	deps = append(deps, svc)

	starts, stops, registrars := app.Setup(ctx, router, deps...)

	if err := app.Start(ctx, logger, starts, stops, registrars, router); err != nil {
		logger.Errorf("Cannot start %s(%s): %v", name, version, err)
		os.Exit(1)
	}

	logger.Infof("%s(%s) started successfully", name, version)

	go func() {
		logger.Infof("Server listening on %s", cfg.Server.Port)
		if err := app.Serve(router, cfg.Server.Port); err != nil {
			logger.Errorf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop

	logger.Infof("Shutting down %s(%s)...", name, version)
	cancel()

	for i := len(stops) - 1; i >= 0; i-- {
		if err := stops[i](context.Background()); err != nil {
			logger.Errorf("Error stopping component: %v", err)
		}
	}

	fmt.Println("Goodbye!")
}

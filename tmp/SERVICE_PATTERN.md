# AQM Service Pattern

This document describes the ideal pattern for building services with AQM. All services should follow this pattern for consistency, maintainability, and clean architecture.

## Overview

AQM services use a **declarative lifecycle pattern** where components implement standard interfaces (`Startable`, `Stoppable`, `RouteRegistrar`) and the `aqm/app` package manages their lifecycle automatically.

This approach provides:
- **Clean main.go**: Just dependency construction + lifecycle delegation
- **Explicit dependencies**: No globals, no magic, no framework
- **Automatic lifecycle management**: Start/stop in correct order with rollback
- **Interface-based composition**: Components declare capabilities via interfaces
- **Homogeneous structure**: All services look similar and predictable

## The Pattern

### main.go Structure

Every AQM service main.go should follow this structure:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/aquamarinepk/aqm/app"
    "github.com/yourservice/config"
    "github.com/yourservice/internal"
    logger "github.com/aquamarinepk/aqm/log"
)

const (
    name    = "your-service"
    version = "0.1.0"
)

func main() {
    // 1. Create logger
    log := logger.NewLogger("info")

    // 2. Load configuration
    cfg, err := config.New(log)
    if err != nil {
        log.Errorf("Cannot load config: %v", err)
        os.Exit(1)
    }

    // 3. Create context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 4. Create router with options
    router := app.NewRouter(log)
    app.ApplyRouterOptions(router,
        app.WithDefaultInternalStack(),
        app.WithPing(),
        app.WithDebugRoutes(),
        app.WithHealthChecks(name, version),
    )

    // 5. Build dependencies
    var deps []any

    svc, err := internal.New(cfg)
    if err != nil {
        log.Errorf("Cannot create service: %v", err)
        os.Exit(1)
    }

    deps = append(deps, svc)

    // 6. Setup lifecycle
    starts, stops, registrars := app.Setup(ctx, router, deps...)

    // 7. Start components
    if err := app.Start(ctx, log, starts, stops, registrars, router); err != nil {
        log.Errorf("Cannot start %s(%s): %v", name, version, err)
        os.Exit(1)
    }

    log.Infof("%s(%s) started successfully", name, version)

    // 8. Start HTTP server
    go func() {
        log.Infof("Server listening on %s", cfg.Server.Port)
        if err := app.Serve(router, cfg.Server.Port); err != nil {
            log.Errorf("Server error: %v", err)
        }
    }()

    // 9. Wait for shutdown signal
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
    <-stop

    // 10. Graceful shutdown
    log.Infof("Shutting down %s(%s)...", name, version)
    cancel()

    for i := len(stops) - 1; i >= 0; i-- {
        if err := stops[i](context.Background()); err != nil {
            log.Errorf("Error stopping component: %v", err)
        }
    }

    fmt.Println("Goodbye!")
}
```

## Component Interfaces

Components declare their capabilities by implementing these interfaces:

### Startable

Components that need initialization implement `Startable`:

```go
type Startable interface {
    Start(context.Context) error
}
```

**Example:**

```go
func (s *Service) Start(ctx context.Context) error {
    // Initialize resources, bootstrap data, etc.
    if s.cfg.IsBootstrapEnabled() {
        if err := s.bootstrap(ctx); err != nil {
            return fmt.Errorf("bootstrap failed: %w", err)
        }
    }

    log.Println("Service started successfully")
    return nil
}
```

### Stoppable

Components that need cleanup implement `Stoppable`:

```go
type Stoppable interface {
    Stop(context.Context) error
}
```

**Example:**

```go
func (s *Service) Stop(ctx context.Context) error {
    // Close database connections, release resources, etc.
    if s.db != nil {
        if err := s.db.Close(); err != nil {
            return fmt.Errorf("database close error: %w", err)
        }
    }

    log.Println("Service stopped successfully")
    return nil
}
```

### RouteRegistrar

Components that expose HTTP endpoints implement `RouteRegistrar`:

```go
type RouteRegistrar interface {
    RegisterRoutes(chi.Router)
}
```

**Example:**

```go
func (s *Service) RegisterRoutes(r chi.Router) {
    // Add middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // Register handler routes
    s.authnHandler.RegisterRoutes(r)
    s.authzHandler.RegisterRoutes(r)
}
```

## Router Options

AQM provides a clean RouterOptions pattern for configuring the HTTP router. All configuration happens in `app.ApplyRouterOptions()`.

### Standard Setup

The recommended pattern for all services:

```go
router := app.NewRouter(logger)
app.ApplyRouterOptions(router,
    app.WithDefaultInternalStack(),  // Middleware + InternalOnly
    app.WithPing(),                   // GET /ping
    app.WithDebugRoutes(),            // GET /debug/routes
    app.WithHealthChecks(name, version), // GET /health
)
```

### Available Options

**Middleware Options:**

- `WithDefaultStack()` - Standard middleware (RequestID, RealIP, Logger, Recoverer)
- `WithDefaultInternalStack()` - DefaultStack + InternalOnly restriction

**Route Options:**

- `WithPing()` - Adds `GET /ping` endpoint returning `{"status":"ok"}`
- `WithDebugRoutes()` - Adds `GET /debug/routes` listing all registered routes
- `WithHealthChecks(name, version)` - Adds `GET /health` with service info

### Internal vs Public Services

**Internal services** (not exposed to public internet):
```go
app.WithDefaultInternalStack()  // Includes InternalOnly restriction
```

**Public services** (exposed to internet):
```go
app.WithDefaultStack()  // No IP restrictions
```

The `InternalOnly` middleware restricts access to:
- Localhost (127.0.0.1, ::1)
- Private IPv4 ranges (10.x, 172.16-31.x, 192.168.x)
- IPv6 ULA (fc00::/7, fd00::/8)

**Defense-in-depth**: This complements (does not replace) network policies at the infrastructure level.

### Custom Middleware

Additional middleware can be added directly:

```go
router := app.NewRouter(logger)
app.ApplyRouterOptions(router, app.WithDefaultStack())
router.Use(customMiddleware1)
router.Use(customMiddleware2)
app.ApplyRouterOptions(router, app.WithPing(), app.WithDebugRoutes())
```

**Important**: Middleware must be added BEFORE routes.

## Lifecycle Flow

### Startup

1. **Logger creation**: First thing in main()
2. **Config loading**: Uses logger for observability
3. **Dependency construction**: Explicit, in order
4. **app.Setup()**: Extracts interfaces from dependencies
5. **app.Start()**: Calls all `Start()` methods with rollback on failure
6. **Route registration**: All `RegisterRoutes()` called
7. **HTTP server**: Started in background goroutine

**Key points:**
- Components start in dependency order
- If any Start() fails, already-started components are stopped in reverse
- Routes registered only after all components successfully start

### Shutdown

1. **Signal received**: SIGINT, SIGTERM, or SIGQUIT
2. **Context cancel**: Propagated to all goroutines
3. **Component shutdown**: All `Stop()` called in **reverse order** (LIFO)

**Key points:**
- Reverse order ensures dependency cleanup cascade
- Each component gets context for graceful shutdown
- Errors logged but shutdown continues

## What NOT to Do in main.go

❌ **Don't create HTTP server manually**

```go
// BAD
server := &http.Server{
    Addr:    cfg.Server.Port,
    Handler: router,
}
server.ListenAndServe()
```

✅ **Use app.Serve()**

```go
// GOOD
app.Serve(router, cfg.Server.Port)
```

---

❌ **Don't manage goroutines and error channels**

```go
// BAD
errCh := make(chan error, 1)
go func() {
    if err := svc.Start(ctx); err != nil {
        errCh <- err
    }
}()

select {
case sig := <-sigCh:
    // ...
case err := <-errCh:
    // ...
}
```

✅ **Use app.Start()**

```go
// GOOD
if err := app.Start(ctx, log, starts, stops, registrars, router); err != nil {
    log.Errorf("Cannot start: %v", err)
    os.Exit(1)
}
```

---

❌ **Don't manage shutdown logic manually**

```go
// BAD
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := server.Shutdown(shutdownCtx); err != nil {
    log.Printf("Shutdown error: %v", err)
}

if err := svc.Stop(shutdownCtx); err != nil {
    log.Printf("Stop error: %v", err)
}
```

✅ **Use stop functions from app.Setup()**

```go
// GOOD
for i := len(stops) - 1; i >= 0; i-- {
    if err := stops[i](context.Background()); err != nil {
        log.Errorf("Error stopping component: %v", err)
    }
}
```

## Real Example: authn Service

See [`examples/ticked/services/authn`](examples/ticked/services/authn) for a complete example following this pattern.

**Key files:**

- [`main.go`](examples/ticked/services/authn/main.go): Clean main following the pattern (86 lines)
- [`internal/service.go`](examples/ticked/services/authn/internal/service.go): Implements `Startable`, `Stoppable`, `RouteRegistrar`
- [`config/config.go`](examples/ticked/services/authn/config/config.go): Uses `aqm/config` base

**Metrics:**
- main.go: 86 lines (vs 71 before, but cleaner and more standard)
- No manual signal handling
- No error channels
- No manual HTTP server management
- Follows exact pattern every service should use

## Benefits

### Consistency

All services have identical structure:
1. Logger
2. Config
3. Context
4. Router
5. Dependencies
6. Setup
7. Start
8. Serve
9. Signal
10. Shutdown

### Maintainability

- **Predictable**: Any developer can navigate any service
- **Testable**: Components are isolated and injectable
- **Debuggable**: Clear startup/shutdown order
- **Extensible**: Add components to deps slice

### Safety

- **Automatic rollback**: Failed startup cleans up properly
- **Reverse shutdown**: Dependencies cleaned in correct order
- **No leaks**: Resources properly released
- **Graceful**: Context propagation for clean shutdown

## Migration Guide

### From Manual Pattern

If your service has manual lifecycle management:

**Before:**

```go
func main() {
    // ...
    errCh := make(chan error, 1)
    go func() {
        errCh <- svc.Start(ctx)
    }()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    select {
    case sig := <-sigCh:
        logger.Printf("Received signal: %v", sig)
    case err := <-errCh:
        logger.Printf("Service error: %v", err)
    }

    // Manual shutdown...
}
```

**After:**

```go
func main() {
    // ...
    var deps []any
    deps = append(deps, svc)

    starts, stops, registrars := app.Setup(ctx, router, deps...)

    if err := app.Start(ctx, log, starts, stops, registrars, router); err != nil {
        log.Errorf("Cannot start: %v", err)
        os.Exit(1)
    }

    // ...
}
```

### Component Changes

1. **Remove HTTP server from component**:
   - Delete `server *http.Server` field
   - Remove server creation from `Start()`
   - Remove server shutdown from `Stop()`

2. **Add RegisterRoutes method**:
   - Move route registration from `Start()` to `RegisterRoutes(r chi.Router)`

3. **Simplify Start()**:
   - Only initialization logic
   - No HTTP server, no blocking

4. **Simplify Stop()**:
   - Only cleanup logic
   - No server shutdown

## Checklist

When creating or refactoring a service, ensure:

- [ ] main.go follows the exact pattern (10 steps)
- [ ] Component implements `Startable` if needs initialization
- [ ] Component implements `Stoppable` if needs cleanup
- [ ] Component implements `RouteRegistrar` if exposes HTTP endpoints
- [ ] No HTTP server management in components
- [ ] No manual signal handling in main.go
- [ ] No error channels or goroutine management
- [ ] Uses `app.NewRouter()` with options
- [ ] Uses `app.Setup()` for lifecycle extraction
- [ ] Uses `app.Start()` for component startup
- [ ] Uses `app.Serve()` for HTTP server
- [ ] Shutdown iterates stops in reverse

## Related Documentation

- [`aqm/app` package](app/lifecycle.go): Lifecycle management implementation
- [`aqm/config` package](config/README.md): Configuration system
- [authn service](examples/ticked/services/authn/README.md): Complete example

## Summary

**The pattern is simple:**

1. main.go constructs dependencies
2. app.Setup() extracts capabilities
3. app.Start() manages lifecycle
4. app.Serve() runs HTTP server
5. Reverse iteration stops components

**Every service should look like this. No exceptions.**

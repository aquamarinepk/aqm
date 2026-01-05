# AQM Configuration System

A hybrid configuration system providing both static baseline configuration and dynamic service-specific parameters for AQM services.

## Overview

The `aqm/config` package provides:

- **Static Baseline Configuration**: Type-safe structs for common config (Server, Database, NATS, Log, Assets)
- **Dynamic Configuration Access**: Map-based access for service-specific parameters via `GetString`, `GetInt`, `GetBool`, etc.
- **Options Pattern**: Flexible initialization with `WithPrefix`, `WithFile`, `WithDefaults`
- **Logger Integration**: Injected logger for observability during config loading
- **Configuration Precedence**: defaults → file → environment variables

This approach allows services to reuse ~80% of configuration infrastructure while maintaining flexibility for custom parameters.

## Quick Start

```go
package main

import (
    "github.com/aquamarinepk/aqm/config"
    "github.com/aquamarinepk/aqm/log"
)

func main() {
    // Create logger first
    logger := log.NewLogger("info")

    // Load configuration
    cfg, err := config.New(logger,
        config.WithPrefix("SERVICE_"),
        config.WithFile("config.yaml"),
        config.WithDefaults(map[string]interface{}{
            "server.port": ":8080",
            "custom.param": "value",
        }),
    )
    if err != nil {
        logger.Errorf("Failed to load config: %v", err)
        return
    }

    // Access static baseline config (type-safe)
    port := cfg.Server.Port
    dbHost := cfg.Database.Host
    logLevel := cfg.Log.Level

    // Access dynamic service-specific config
    customParam := cfg.GetString("custom.param")
    maxRetries := cfg.GetInt("custom.maxretries")
    enableFeature := cfg.GetBool("custom.enablefeature")
}
```

## Configuration Structure

### Static Baseline Configuration

All services inherit these common configuration sections:

```go
type Config struct {
    Server   ServerConfig   `koanf:"server"`
    Database DatabaseConfig `koanf:"database"`
    NATS     NATSConfig     `koanf:"nats"`
    Log      LogConfig      `koanf:"log"`
    Assets   AssetsConfig   `koanf:"assets"`
}
```

#### Server Configuration

```go
type ServerConfig struct {
    Port string `koanf:"port"` // Default: ":8080"
}
```

Environment variable: `PREFIX_SERVER_PORT`

#### Database Configuration

```go
type DatabaseConfig struct {
    Driver   string `koanf:"driver"`   // "fake", "postgres", "mongo"
    Host     string `koanf:"host"`
    Port     int    `koanf:"port"`
    User     string `koanf:"user"`
    Password string `koanf:"password"`
    Database string `koanf:"database"`
    SSLMode  string `koanf:"sslmode"`
    Schema   string `koanf:"schema"`
}
```

Environment variables:
- `PREFIX_DATABASE_DRIVER`
- `PREFIX_DATABASE_HOST`
- `PREFIX_DATABASE_PORT`
- `PREFIX_DATABASE_USER`
- `PREFIX_DATABASE_PASSWORD`
- `PREFIX_DATABASE_DATABASE`
- `PREFIX_DATABASE_SSLMODE`
- `PREFIX_DATABASE_SCHEMA`

#### NATS Configuration

```go
type NATSConfig struct {
    URL          string `koanf:"url"`
    ClusterID    string `koanf:"clusterid"`
    ClientID     string `koanf:"clientid"`
    MaxReconnect int    `koanf:"maxreconnect"`
}
```

Environment variables:
- `PREFIX_NATS_URL`
- `PREFIX_NATS_CLUSTERID`
- `PREFIX_NATS_CLIENTID`
- `PREFIX_NATS_MAXRECONNECT`

#### Log Configuration

```go
type LogConfig struct {
    Level  string `koanf:"level"`  // "debug", "info", "error"
    Format string `koanf:"format"` // "text", "json"
}
```

Environment variables:
- `PREFIX_LOG_LEVEL`
- `PREFIX_LOG_FORMAT`

#### Assets Configuration

```go
type AssetsConfig struct {
    Path string `koanf:"path"`
}
```

Environment variable: `PREFIX_ASSETS_PATH`

### Dynamic Service-Specific Configuration

Use dynamic access methods for service-specific parameters:

```go
// String values
tokenTTL := cfg.GetString("auth.tokenttl")
apiKey := cfg.GetString("external.apikey")

// Integer values
maxRetries := cfg.GetInt("client.maxretries")
poolSize := cfg.GetInt("pool.size")

// Boolean values
enableFeature := cfg.GetBool("features.experimental")
debug := cfg.GetBool("debug.enabled")

// Float values
timeout := cfg.GetFloat("client.timeout")

// Duration values
ttl, err := cfg.GetDuration("cache.ttl") // Parses "24h", "5m", etc.

// Check existence
if cfg.Exists("optional.setting") {
    value := cfg.GetString("optional.setting")
}
```

## Options Pattern

### WithPrefix

Sets the environment variable prefix for your service:

```go
cfg, err := config.New(logger, config.WithPrefix("AUTHN_"))
// Reads: AUTHN_SERVER_PORT, AUTHN_DATABASE_DRIVER, etc.
```

### WithFile

Loads configuration from a YAML file:

```go
cfg, err := config.New(logger, config.WithFile("config.yaml"))
```

Example `config.yaml`:

```yaml
server:
  port: ":8080"

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "myuser"
  database: "mydb"
  sslmode: "disable"

log:
  level: "info"
  format: "text"

# Service-specific config
auth:
  tokenttl: "24h"
  passwordlength: 32

custom:
  feature: true
  maxretries: 3
```

### WithDefaults

Provides default values for your service:

```go
cfg, err := config.New(logger,
    config.WithDefaults(map[string]interface{}{
        "server.port": ":8080",
        "auth.tokenttl": "24h",
        "auth.passwordlength": 32,
        "features.experimental": false,
    }),
)
```

### WithEnvExpansion

Enables environment variable expansion in config files:

```go
cfg, err := config.New(logger,
    config.WithFile("config.yaml"),
    config.WithEnvExpansion(),
)
```

In `config.yaml`:

```yaml
database:
  password: "${DB_PASSWORD}"  # Expands to value of DB_PASSWORD env var
  host: "${DB_HOST:-localhost}"  # Defaults to "localhost" if DB_HOST not set
```

### Combining Options

Options can be combined and are processed in order:

```go
cfg, err := config.New(logger,
    config.WithPrefix("MYSERVICE_"),
    config.WithDefaults(defaults),
    config.WithFile("config.yaml"),
    config.WithEnvExpansion(),
)
```

**Loading order** (later sources override earlier):
1. Defaults from `WithDefaults()`
2. Config file from `WithFile()`
3. Environment variables from `WithPrefix()`

## Environment Variable Naming

Environment variables follow this pattern:

```
PREFIX_SECTION_FIELDNAME
```

**Important Rules:**

1. **No underscores in field names**: Field names must match koanf tags exactly
   - ✅ `AUTHN_CRYPTO_ENCRYPTIONKEY` → `crypto.encryptionkey`
   - ❌ `AUTHN_CRYPTO_ENCRYPTION_KEY` (won't work)

2. **Sections separated by underscores**:
   - `PREFIX_SERVER_PORT` → `server.port`
   - `PREFIX_DATABASE_HOST` → `database.host`
   - `PREFIX_AUTH_TOKENTTL` → `auth.tokenttl`

3. **Prefix is required**: Without `WithPrefix()`, no env vars are loaded

## Validation

Configuration is automatically validated during `New()`:

```go
cfg, err := config.New(logger, opts...)
if err != nil {
    // Validation failed - err contains details
}
```

**Validation rules:**

- `server.port` must not be empty
- `database.driver` must be "fake", "postgres", or "mongo"
- If `database.driver` is "postgres" or "mongo", `database.host` is required
- `log.level` must be "debug", "info", or "error"
- `log.format` must be "text" or "json" (if set)

## Service-Specific Configuration

Services should wrap `aqm/config.Config` and add their own validation:

```go
package config

import (
    "fmt"
    aqmconfig "github.com/aquamarinepk/aqm/config"
    "github.com/aquamarinepk/aqm/log"
)

// Config wraps aqm.Config and adds service-specific helpers
type Config struct {
    *aqmconfig.Config
}

func New(logger log.Logger) (*Config, error) {
    // Service-specific defaults
    defaults := map[string]interface{}{
        "auth.tokenttl": "24h",
        "auth.passwordlength": 32,
        "crypto.encryptionkey": "",
        "crypto.signingkey": "",
    }

    base, err := aqmconfig.New(logger,
        aqmconfig.WithPrefix("MYSERVICE_"),
        aqmconfig.WithFile("config.yaml"),
        aqmconfig.WithDefaults(defaults),
    )
    if err != nil {
        return nil, err
    }

    cfg := &Config{Config: base}

    // Service-specific validation
    if err := cfg.ValidateService(); err != nil {
        return nil, err
    }

    return cfg, nil
}

// ValidateService validates service-specific configuration
func (c *Config) ValidateService() error {
    encKey := c.GetString("crypto.encryptionkey")
    if encKey == "" {
        return fmt.Errorf("crypto.encryptionkey is required")
    }
    if len(encKey) != 64 {
        return fmt.Errorf("crypto.encryptionkey must be 64 hex characters")
    }
    return nil
}

// Service-specific helper methods
func (c *Config) GetTokenTTL() (time.Duration, error) {
    return c.GetDuration("auth.tokenttl")
}

func (c *Config) GetPasswordLength() int {
    return c.GetInt("auth.passwordlength")
}
```

## Migration Guide

### From Old Config System

**Before:**

```go
// Old config loading
cfg, err := config.LoadConfig("config.yaml", "SERVICE_", os.Args)
```

**After:**

```go
// New config with logger injection
logger := log.NewLogger("info")
cfg, err := config.New(logger,
    config.WithPrefix("SERVICE_"),
    config.WithFile("config.yaml"),
)
```

**Key changes:**

1. Logger is created first and passed to config
2. Options pattern replaces multiple parameters
3. Config validation happens during `New()`
4. All logging goes through injected logger (no more `fmt.Printf`)

### From Service-Specific Config

If your service reimplemented configuration (like `authn` did):

**Before (249 lines of duplicated config code):**

```go
package config

// Full config implementation: structs, loading, validation, helpers
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Crypto   CryptoConfig
    Auth     AuthConfig
    Log      LogConfig
}

func Load(prefix string) (*Config, error) {
    // 200+ lines of config loading logic
}
```

**After (151 lines, 39% reduction):**

```go
package config

import aqmconfig "github.com/aquamarinepk/aqm/config"

// Config wraps aqm.Config
type Config struct {
    *aqmconfig.Config
}

func New(logger log.Logger) (*Config, error) {
    defaults := map[string]interface{}{
        "crypto.encryptionkey": "",
        // service-specific defaults only
    }

    base, err := aqmconfig.New(logger,
        aqmconfig.WithPrefix("SERVICE_"),
        aqmconfig.WithDefaults(defaults),
    )
    if err != nil {
        return nil, err
    }

    return &Config{Config: base}, nil
}
```

**Benefits:**

- Reuse baseline config infrastructure (Server, Database, Log, etc.)
- Only implement service-specific validation and helpers
- Automatic inheritance of new base features
- Reduced maintenance burden

## Testing

### Test with Fake Config

```go
func TestWithConfig(t *testing.T) {
    logger := log.NewLogger("info")

    cfg, err := config.New(logger,
        config.WithDefaults(map[string]interface{}{
            "server.port": ":8080",
            "database.driver": "fake",
            "log.level": "info",
        }),
    )
    if err != nil {
        t.Fatalf("Failed to create config: %v", err)
    }

    // Test with config
    if cfg.Server.Port != ":8080" {
        t.Errorf("Expected port :8080, got %s", cfg.Server.Port)
    }
}
```

### Test with Environment Variables

```go
func TestWithEnvVars(t *testing.T) {
    // Set test env vars
    os.Setenv("TEST_SERVER_PORT", ":9090")
    os.Setenv("TEST_DATABASE_DRIVER", "fake")
    os.Setenv("TEST_LOG_LEVEL", "debug")
    defer func() {
        os.Unsetenv("TEST_SERVER_PORT")
        os.Unsetenv("TEST_DATABASE_DRIVER")
        os.Unsetenv("TEST_LOG_LEVEL")
    }()

    logger := log.NewLogger("info")
    cfg, err := config.New(logger, config.WithPrefix("TEST_"))
    if err != nil {
        t.Fatalf("Failed to create config: %v", err)
    }

    if cfg.Server.Port != ":9090" {
        t.Errorf("Expected port :9090, got %s", cfg.Server.Port)
    }
}
```

## Best Practices

1. **Create logger first**: Always create logger before config
2. **Use WithDefaults**: Provide sensible defaults for all parameters
3. **Service-specific wrapper**: Wrap `aqm/config.Config` for service-specific logic
4. **Validate early**: Validate during config initialization, not at use time
5. **Document env vars**: List all env vars in your service README
6. **Never commit secrets**: Use env vars or secret management for sensitive data
7. **Test with fakes**: Use `database.driver: fake` for tests
8. **Explicit over implicit**: Prefer explicit config over "magic" defaults

## Examples

See these services for complete examples:

- [`examples/ticked/services/authn`](../examples/ticked/services/authn): Full authentication service with crypto config
- More examples coming soon

## Coverage

The config package maintains 85%+ test coverage with comprehensive table-driven tests for:

- Options pattern combinations
- Dynamic access methods
- Environment variable parsing
- Validation rules
- Configuration precedence
- Error handling

Run tests:

```bash
cd config
go test -v
go test -cover  # Shows coverage percentage
```

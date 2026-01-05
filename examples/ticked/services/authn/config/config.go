package config

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	aqmconfig "github.com/aquamarinepk/aqm/config"
	log "github.com/aquamarinepk/aqm/log"
)

// Config wraps aqm.Config and adds authn-specific helpers.
type Config struct {
	*aqmconfig.Config // Embed for baseline access
}

// New creates authn service configuration.
func New(logger log.Logger) (*Config, error) {
	// Default values for authn-specific config
	defaults := map[string]interface{}{
		"crypto.encryptionkey":   "",
		"crypto.signingkey":      "",
		"crypto.tokenprivatekey": "",
		"auth.tokenttl":          "24h",
		"auth.passwordlength":    32,
		"auth.enablebootstrap":   true,
		"log.format":             "text",
	}

	base, err := aqmconfig.New(logger,
		aqmconfig.WithPrefix("AUTHN_"),
		aqmconfig.WithFile("config.yaml"),
		aqmconfig.WithDefaults(defaults),
	)
	if err != nil {
		return nil, err
	}

	cfg := &Config{Config: base}

	// Validate authn-specific config
	if err := cfg.ValidateAuthn(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ValidateAuthn validates authn-specific configuration.
func (c *Config) ValidateAuthn() error {
	// Validate crypto.encryptionkey
	encKeyStr := c.GetString("crypto.encryptionkey")
	if encKeyStr == "" {
		return fmt.Errorf("crypto.encryptionkey is required")
	}

	encKey, err := hex.DecodeString(encKeyStr)
	if err != nil {
		return fmt.Errorf("crypto.encryptionkey must be hex-encoded: %w", err)
	}
	if len(encKey) != 32 {
		return fmt.Errorf("crypto.encryptionkey must be 32 bytes (64 hex chars), got %d bytes", len(encKey))
	}

	// Validate crypto.signingkey
	signKeyStr := c.GetString("crypto.signingkey")
	if signKeyStr == "" {
		return fmt.Errorf("crypto.signingkey is required")
	}

	signKey, err := hex.DecodeString(signKeyStr)
	if err != nil {
		return fmt.Errorf("crypto.signingkey must be hex-encoded: %w", err)
	}
	if len(signKey) != 32 {
		return fmt.Errorf("crypto.signingkey must be 32 bytes (64 hex chars), got %d bytes", len(signKey))
	}

	// Validate crypto.tokenprivatekey
	tokenKeyStr := c.GetString("crypto.tokenprivatekey")
	if tokenKeyStr == "" {
		return fmt.Errorf("crypto.tokenprivatekey is required")
	}

	tokenKey, err := base64.StdEncoding.DecodeString(tokenKeyStr)
	if err != nil {
		return fmt.Errorf("crypto.tokenprivatekey must be base64-encoded: %w", err)
	}
	if len(tokenKey) != 64 {
		return fmt.Errorf("crypto.tokenprivatekey must be 64 bytes (Ed25519), got %d bytes", len(tokenKey))
	}

	// Validate auth.tokenttl
	tokenTTL := c.GetString("auth.tokenttl")
	if tokenTTL != "" {
		if _, err := time.ParseDuration(tokenTTL); err != nil {
			return fmt.Errorf("auth.tokenttl invalid duration: %w", err)
		}
	}

	// Validate auth.passwordlength
	passwordLength := c.GetInt("auth.passwordlength")
	if passwordLength < 16 {
		return fmt.Errorf("auth.passwordlength must be at least 16, got %d", passwordLength)
	}

	// Validate log.format
	logFormat := c.GetString("log.format")
	validFormats := map[string]bool{"text": true, "json": true}
	if !validFormats[logFormat] {
		return fmt.Errorf("log.format must be 'text' or 'json', got '%s'", logFormat)
	}

	return nil
}

// DecodeEncryptionKey decodes the hex-encoded encryption key.
func (c *Config) DecodeEncryptionKey() ([]byte, error) {
	return hex.DecodeString(c.GetString("crypto.encryptionkey"))
}

// DecodeSigningKey decodes the hex-encoded signing key.
func (c *Config) DecodeSigningKey() ([]byte, error) {
	return hex.DecodeString(c.GetString("crypto.signingkey"))
}

// DecodeTokenPrivateKey decodes the base64-encoded token private key.
func (c *Config) DecodeTokenPrivateKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(c.GetString("crypto.tokenprivatekey"))
}

// ParseTokenTTL parses the token TTL duration.
func (c *Config) ParseTokenTTL() (time.Duration, error) {
	return c.GetDuration("auth.tokenttl")
}

// GetLogFormat returns the log format (text or json).
func (c *Config) GetLogFormat() string {
	return c.GetString("log.format")
}

// GetPasswordLength returns the password length.
func (c *Config) GetPasswordLength() int {
	return c.GetInt("auth.passwordlength")
}

// IsBootstrapEnabled returns whether bootstrap is enabled.
func (c *Config) IsBootstrapEnabled() bool {
	return c.GetBool("auth.enablebootstrap")
}

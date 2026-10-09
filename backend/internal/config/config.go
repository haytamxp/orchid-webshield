package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	minHTTPMaxHeaderBytes = 1 << 10
	maxHTTPMaxHeaderBytes = 1 << 20
	maxRequestBodyBytes   = 10 << 20
	maxServerTimeout      = 10 * time.Minute
)

// Config contains validated HTTP server configuration.
//
// Environment variables are parsed once during startup. Invalid configuration
// causes startup to fail instead of silently disabling security controls.
type Config struct {
	HTTPAddr            string
	ReadTimeout         time.Duration
	ReadHeaderTimeout   time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	ShutdownTimeout     time.Duration
	MaxHeaderBytes      int
	MaxRequestBodyBytes int64
	LogLevel            string
}

// Load reads configuration from environment variables, applies defaults,
// and validates the resulting configuration.
func Load() (Config, error) {
	readTimeout, err := durationFromEnv(
		"HTTP_READ_TIMEOUT",
		10*time.Second,
	)
	if err != nil {
		return Config{}, err
	}

	readHeaderTimeout, err := durationFromEnv(
		"HTTP_READ_HEADER_TIMEOUT",
		5*time.Second,
	)
	if err != nil {
		return Config{}, err
	}

	writeTimeout, err := durationFromEnv(
		"HTTP_WRITE_TIMEOUT",
		15*time.Second,
	)
	if err != nil {
		return Config{}, err
	}

	idleTimeout, err := durationFromEnv(
		"HTTP_IDLE_TIMEOUT",
		60*time.Second,
	)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := durationFromEnv(
		"HTTP_SHUTDOWN_TIMEOUT",
		10*time.Second,
	)
	if err != nil {
		return Config{}, err
	}

	maxHeaderBytes, err := int64FromEnv(
		"HTTP_MAX_HEADER_BYTES",
		maxHTTPMaxHeaderBytes,
	)
	if err != nil {
		return Config{}, err
	}

	maxBodyBytes, err := int64FromEnv(
		"MAX_REQUEST_BODY_BYTES",
		1<<20,
	)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:            stringFromEnv("HTTP_ADDR", "127.0.0.1:8080"),
		ReadTimeout:         readTimeout,
		ReadHeaderTimeout:   readHeaderTimeout,
		WriteTimeout:        writeTimeout,
		IdleTimeout:         idleTimeout,
		ShutdownTimeout:     shutdownTimeout,
		MaxHeaderBytes:      int(maxHeaderBytes),
		MaxRequestBodyBytes: maxBodyBytes,
		LogLevel:            stringFromEnv("LOG_LEVEL", "info"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate checks the configuration and normalizes the log level.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.HTTPAddr) == "" {
		return fmt.Errorf("HTTP_ADDR must not be empty")
	}

	_, portText, err := net.SplitHostPort(c.HTTPAddr)
	if err != nil {
		return fmt.Errorf(
			"HTTP_ADDR must be a valid host:port address: %w",
			err,
		)
	}

	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("HTTP_ADDR must contain a valid TCP port")
	}

	timeouts := []struct {
		name  string
		value time.Duration
	}{
		{"HTTP_READ_TIMEOUT", c.ReadTimeout},
		{"HTTP_READ_HEADER_TIMEOUT", c.ReadHeaderTimeout},
		{"HTTP_WRITE_TIMEOUT", c.WriteTimeout},
		{"HTTP_IDLE_TIMEOUT", c.IdleTimeout},
		{"HTTP_SHUTDOWN_TIMEOUT", c.ShutdownTimeout},
	}

	for _, timeout := range timeouts {
		if timeout.value <= 0 {
			return fmt.Errorf("%s must be greater than zero", timeout.name)
		}

		if timeout.value > maxServerTimeout {
			return fmt.Errorf(
				"%s must not exceed %s",
				timeout.name,
				maxServerTimeout,
			)
		}
	}

	if c.ReadTimeout < c.ReadHeaderTimeout {
		return fmt.Errorf(
			"HTTP_READ_TIMEOUT must be >= HTTP_READ_HEADER_TIMEOUT",
		)
	}

	if c.MaxHeaderBytes < minHTTPMaxHeaderBytes ||
		c.MaxHeaderBytes > maxHTTPMaxHeaderBytes {
		return fmt.Errorf(
			"HTTP_MAX_HEADER_BYTES must be between %d and %d",
			minHTTPMaxHeaderBytes,
			maxHTTPMaxHeaderBytes,
		)
	}

	if c.MaxRequestBodyBytes < 1 ||
		c.MaxRequestBodyBytes > maxRequestBodyBytes {
		return fmt.Errorf(
			"MAX_REQUEST_BODY_BYTES must be between 1 and %d",
			maxRequestBodyBytes,
		)
	}

	c.LogLevel = strings.ToLower(strings.TrimSpace(c.LogLevel))

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
		// Valid log level.
	default:
		return fmt.Errorf(
			"LOG_LEVEL must be one of: debug, info, warn, error",
		)
	}

	return nil
}

func stringFromEnv(name, fallback string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fallback
	}

	return strings.TrimSpace(value)
}

func durationFromEnv(
	name string,
	fallback time.Duration,
) (time.Duration, error) {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", name, err)
	}

	return parsed, nil
}

func int64FromEnv(name string, fallback int64) (int64, error) {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}

	return parsed, nil
}

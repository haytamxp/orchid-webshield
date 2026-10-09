package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/haytamxp/orchid-webshield/backend/internal/api"
	"github.com/haytamxp/orchid-webshield/backend/internal/config"
)

func main() {
	bootstrapLogger := newLogger("info")

	cfg, err := config.Load()
	if err != nil {
		bootstrapLogger.Error(
			"configuration_invalid",
			"error",
			err,
		)
		os.Exit(1)
	}

	logger := newLogger(cfg.LogLevel)

	if err := run(cfg, logger); err != nil {
		logger.Error("server_failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, logger *slog.Logger) error {
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.NewRouter(logger, cfg.MaxRequestBodyBytes),
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	signalContext, stopSignals := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopSignals()

	serverErrors := make(chan error, 1)

	logger.Info("http_server_starting", "address", cfg.HTTPAddr)

	go func() {
		err := server.ListenAndServe()

		if errors.Is(err, http.ErrServerClosed) {
			serverErrors <- nil
			return
		}

		serverErrors <- err
	}()

	select {
	case err := <-serverErrors:
		if err != nil {
			return fmt.Errorf("HTTP server stopped unexpectedly: %w", err)
		}

		return nil

	case <-signalContext.Done():
		logger.Info("http_server_shutdown_started")
	}

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful_shutdown_failed", "error", err)

		closeErr := server.Close()
		if closeErr != nil {
			return errors.Join(
				fmt.Errorf("graceful shutdown failed: %w", err),
				fmt.Errorf("forced server close failed: %w", closeErr),
			)
		}

		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	if err := <-serverErrors; err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}

	logger.Info("http_server_shutdown_completed")

	return nil
}

func newLogger(level string) *slog.Logger {
	logLevel := slog.LevelInfo

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	return slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: logLevel,
			},
		),
	)
}

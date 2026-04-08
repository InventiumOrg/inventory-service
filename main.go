package main

import (
	"context"
	"inventory-service/api"
	"inventory-service/config"
	"inventory-service/observability"
	"inventory-service/processors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/segmentio/kafka-go"
)

var conn *pgx.Conn
var consumer *kafka.Reader
var dialer *kafka.Dialer

const attemptThreshold = 5

// setupLogging configures logging to use OTLP
func setupLogging(cfg config.Config) error {
	// Use OTLP for logs if configured
	if cfg.OTELExporterOTLPEndpoint != "" {
		endpoint := "http://" + cfg.OTELExporterOTLPEndpoint
		if err := observability.SetupOTLPLogging(endpoint, cfg.ServiceName); err == nil {
			slog.Info("OTLP logging configured", slog.String("endpoint", endpoint))
			return nil
		} else {
			slog.Warn("OTLP logging failed, falling back to stdout", slog.Any("error", err))
		}
	}

	// Fallback to stdout JSON logging
	slog.Info("Using default stdout logging")
	return nil
}

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		slog.Error("Failed to load config: ", slog.Any("ERROR", err))
		os.Exit(1)
	}

	slog.Info("Set Up Logging.....")
	// Setup logging based on configuration
	if err := setupLogging(config); err != nil {
		slog.Error("Failed to setup logging", slog.Any("error", err))
		// Continue with stdout logging if setup fails
	}

	time.Sleep(10 * time.Second)
	slog.Info("Connecting to database")

	attempt := 1
	for attempt <= attemptThreshold {
		conn, err = pgx.Connect(context.Background(), config.DBSource)
		if err == nil {
			slog.Info("Connected to database successfully")
			break
		}
		slog.Error("Failed to connect to database",
			slog.Int("attempt", attempt),
			slog.Int("maxAttempts", attemptThreshold),
			slog.Any("error", err),
		)

		if attempt == attemptThreshold {
			slog.Error("Max connection attempts reached, exiting", slog.Any("ERROR", err))
			os.Exit(1)
		}

		backoffDuration := time.Duration(1<<(attempt-1)) * time.Second
		slog.Info("Retrying connection",
			slog.Int("attempt", attempt+1),
			slog.Duration("backoff", backoffDuration),
		)

		time.Sleep(backoffDuration)
		attempt++
	}

	// Create server with inventory-specific service name
	router := api.NewServer(conn, config.ServiceName, "1.0.0", config.OTELExporterOTLPEndpoint, config.OTELExporterOTLPHeaders)
	invetoryProcessor := processors.NewInventoryProcessor(conn, consumer, dialer)
	err = invetoryProcessor.Init(config)
	if err != nil {
		slog.Error("Error bootstrapping Kafka consumer!", slog.Any("ERROR", err))
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to listen for interrupt signal to trigger shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start Kafka consumer in a goroutine
	go func() {
		slog.Info("Starting Kafka consumer...")
		invetoryProcessor.Start(ctx)
	}()

	// Start HTTP server in a goroutine
	go func() {
		slog.Info("Starting HTTP server on :13740...")
		if err := router.Run(":13740", config.ServiceName); err != nil {
			slog.Error("HTTP server failed", "error", err)
			cancel()
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	<-quit
	slog.Info("Shutting down server...")

	// Cancel context to stop Kafka consumer
	cancel()

	slog.Info("Server exited")
}

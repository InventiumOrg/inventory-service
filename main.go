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

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/kafka-go"
)

var conn *pgx.Conn
var consumer *kafka.Reader
var dialer *kafka.Dialer

const attemptThreshold = 5

// setupLogging configures logging based on environment variables
func setupLogging(cfg config.Config) error {
	// Priority order: OTLP > Loki > Syslog > File > Stdout

	// Option 1: Direct OTLP Logs (recommended for OpenTelemetry)
	if cfg.OTELExporterOTLPEndpoint != "" {
		endpoint := "http://" + cfg.OTELExporterOTLPEndpoint
		if err := observability.SetupOTLPLogging(endpoint, cfg.ServiceName); err == nil {
			slog.Info("Using OTLP logging", slog.String("endpoint", endpoint))
			return nil
		}
		slog.Warn("OTLP logging failed, trying next option")
	}

	// Option 2: Direct Loki HTTP (no file needed)
	if cfg.LokiURL != "" {
		if err := observability.SetupDirectLokiLogging(cfg.LokiURL, cfg.ServiceName); err == nil {
			slog.Info("Using direct Loki logging", slog.String("url", cfg.LokiURL))
			return nil
		}
		slog.Warn("Direct Loki logging failed, trying next option")
	}

	// Option 3: Syslog (for traditional setups)
	if cfg.SyslogAddress != "" {
		network := cfg.SyslogNetwork
		if network == "" {
			network = "udp"
		}
		if err := observability.SetupSyslogLogging(network, cfg.SyslogAddress, cfg.ServiceName); err == nil {
			slog.Info("Using syslog logging", slog.String("address", cfg.SyslogAddress))
			return nil
		}
		slog.Warn("Syslog logging failed, trying next option")
	}

	// Option 4: File logging (fallback)
	if cfg.LogFilePath != "" {
		logConfig := observability.LogConfig{
			FilePath:   cfg.LogFilePath,
			MaxSizeMB:  100,
			MaxBackups: 5,
			MaxAgeDays: 30,
			Compress:   true,
		}
		if err := observability.SetupAdvancedFileLogger(logConfig); err == nil {
			slog.Info("Using file logging", slog.String("path", cfg.LogFilePath))
			return nil
		}
		slog.Warn("File logging failed, using stdout")
	}

	// Option 5: Default stdout JSON logging
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

	clerk.SetKey(config.ClerkKey)
	slog.Info("Connecting to database", slog.String("db_source", config.DBSource))

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

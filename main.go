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

// setupLogging configures logging based on environment variables
func setupLogging(cfg config.Config) error {
	var handlers []slog.Handler

	// Collect all configured handlers
	handlersConfigured := false

	// Option 1: OTLP Logs (for OpenTelemetry)
	if cfg.OTELExporterOTLPEndpoint != "" {
		endpoint := "http://" + cfg.OTELExporterOTLPEndpoint
		if err := observability.SetupOTLPLogging(endpoint, cfg.ServiceName); err == nil {
			slog.Info("OTLP logging enabled", slog.String("endpoint", endpoint))
			// Note: SetupOTLPLogging sets the default logger internally
			// We'll need to refactor to get the handler instead
			handlersConfigured = true
		} else {
			slog.Warn("OTLP logging failed", slog.Any("error", err))
		}
	}

	// Option 2: Loki HTTP (can run alongside OTLP)
	if cfg.LokiURL != "" {
		lokiConfig := observability.LokiConfig{
			URL: cfg.LokiURL,
			Labels: map[string]string{
				"service": cfg.ServiceName,
				"job":     "go-direct",
				"source":  "application",
			},
			Level: slog.LevelInfo,
		}
		lokiHandler := observability.NewLokiHandler(lokiConfig)
		handlers = append(handlers, lokiHandler)
		slog.Info("Loki logging enabled", slog.String("url", cfg.LokiURL))
		handlersConfigured = true
	}

	// Option 3: Syslog (can run alongside others)
	if cfg.SyslogAddress != "" {
		network := cfg.SyslogNetwork
		if network == "" {
			network = "udp"
		}
		if err := observability.SetupSyslogLogging(network, cfg.SyslogAddress, cfg.ServiceName); err == nil {
			slog.Info("Syslog logging enabled", slog.String("address", cfg.SyslogAddress))
			handlersConfigured = true
		} else {
			slog.Warn("Syslog logging failed", slog.Any("error", err))
		}
	}

	// Option 4: File logging (can run alongside others)
	if cfg.LogFilePath != "" {
		logConfig := observability.LogConfig{
			FilePath:   cfg.LogFilePath,
			MaxSizeMB:  100,
			MaxBackups: 5,
			MaxAgeDays: 30,
			Compress:   true,
		}
		if err := observability.SetupAdvancedFileLogger(logConfig); err == nil {
			slog.Info("File logging enabled", slog.String("path", cfg.LogFilePath))
			handlersConfigured = true
		} else {
			slog.Warn("File logging failed", slog.Any("error", err))
		}
	}

	// If we have multiple handlers, combine them
	if len(handlers) > 1 {
		multiHandler := observability.NewMultiHandler(handlers...)
		logger := slog.New(multiHandler)
		slog.SetDefault(logger)
		slog.Info("Multi-handler logging configured", slog.Int("handlers", len(handlers)))
	} else if len(handlers) == 1 {
		logger := slog.New(handlers[0])
		slog.SetDefault(logger)
	}

	// Option 5: Default stdout JSON logging (if nothing else configured)
	if !handlersConfigured {
		slog.Info("Using default stdout logging")
	}

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

package api

import (
	"context"
	"inventory-service/observability"
	"inventory-service/routes"
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Server struct {
	router            *gin.Engine
	routes            *routes.Route
	db                *pgx.Conn
	otelShutdown      func(context.Context) error
	metrics           *observability.AppMetrics
	prometheusMetrics *observability.PrometheusMetrics
}

func NewServer(db *pgx.Conn, serviceName, serviceVersion, otelEndpoint, otelHeaders string) *Server {
	// Setup OpenTelemetry
	// ctx := context.Background()
	// otelShutdown, err := observability.SetupOTelSDK(ctx, serviceName, serviceVersion, otelEndpoint, otelHeaders)
	// if err != nil {
	// 	slog.Error("Failed to setup OpenTelemetry", slog.Any("error", err))
	// 	// Continue without OpenTelemetry
	// 	otelShutdown = func(context.Context) error { return nil }
	// }

	// Create OTEL metrics
	metrics, err := observability.CreateMetrics()
	if err != nil {
		slog.Error("Failed to create OTEL metrics", slog.Any("error", err))
	}

	// Create Prometheus metrics
	prometheusMetrics := observability.NewPrometheusMetrics(serviceName)

	router := gin.Default()

	// Add metrics middleware
	server := &Server{
		router: router,
		db:     db,
		// otelShutdown:      otelShutdown,
		metrics:           metrics,
		prometheusMetrics: prometheusMetrics,
	}

	// Add Prometheus middleware (this will collect HTTP metrics)
	router.Use(prometheusMetrics.PrometheusMiddleware())

	// Add OTEL middleware for request tracing
	router.Use(server.metricsMiddleware())

	// Setup Prometheus /metrics endpoint
	observability.SetupPrometheusEndpoint(router)

	server.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origins", "Content-Type", "Authorization", "Bearer"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// Setup routes
	server.routes = routes.NewRoute(db, prometheusMetrics)
	server.routes.AddHealthRoutes(router)
	server.routes.AddInventoryRoutes(router)

	return server
}

func (s *Server) Run(addr string, serviceName string) error {
	slog.Info("Starting inventory service server",
		slog.String("address", addr),
		slog.String("service", serviceName))

	return s.router.Run(addr)
}

// Shutdown gracefully shuts down the server and OpenTelemetry
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down inventory service server")

	if s.otelShutdown != nil {
		if err := s.otelShutdown(ctx); err != nil {
			slog.Error("Failed to shutdown OpenTelemetry", slog.Any("error", err))
			return err
		}
	}

	if s.db != nil {
		if err := s.db.Close(ctx); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}

	return nil
}

// metricsMiddleware records HTTP request metrics
func (s *Server) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics if available
		if s.metrics != nil {
			duration := time.Since(start).Seconds()

			// Record request counter
			s.metrics.RequestCounter.Add(c.Request.Context(), 1,
				metric.WithAttributes(
					attribute.String("method", c.Request.Method),
					attribute.String("route", c.FullPath()),
					attribute.Int("status_code", c.Writer.Status()),
				))

			// Record request duration
			s.metrics.RequestDuration.Record(c.Request.Context(), duration,
				metric.WithAttributes(
					attribute.String("method", c.Request.Method),
					attribute.String("route", c.FullPath()),
					attribute.Int("status_code", c.Writer.Status()),
				))
		}
	}
}

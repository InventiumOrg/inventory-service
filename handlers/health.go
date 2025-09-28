package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// HealthzHandler handles the /healthz endpoint for basic liveness check
func (h *Handlers) HealthzHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "inventory-service",
	})
}

// ReadyzHandler handles the /readyz endpoint for readiness check
func (h *Handlers) ReadyzHandler(ctx *gin.Context) {
	// Start a new span for this operation
	spanCtx, span := h.tracer.Start(ctx.Request.Context(), "Health Check")
	defer span.End()

	// Measure database ping duration
	dbStart := time.Now()
	err := h.db.Ping(context.Background())
	dbDuration := time.Since(dbStart).Seconds()

	// Record database operation metrics
	if h.businessMetrics != nil {
		h.businessMetrics.DBOperationDuration.Record(spanCtx, dbDuration,
			metric.WithAttributes(
				attribute.String("operation", "health_check_ping"),
				attribute.String("table", "connection"),
			))
	}

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "database_ping_failed"))

		// Record database error
		if h.businessMetrics != nil {
			h.businessMetrics.DBOperationErrors.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "health_check_ping"),
					attribute.String("error_type", "connection_failed"),
				))
		}

		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "sale-service",
			"error":     "database connection failed",
			"details":   err.Error(),
		})
		return
	}

	// All checks passed
	span.SetAttributes(attribute.String("health.status", "ready"))

	ctx.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "sale-service",
		"checks": gin.H{
			"database": "ok",
		},
	})
}

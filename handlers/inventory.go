package handlers

import (
	models "inventory-service/models/sqlc"
	"inventory-service/observability"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	queries           *models.Queries
	tracer            trace.Tracer
	db                *pgx.Conn
	prometheusMetrics *observability.PrometheusMetrics
}

func NewHandlers(db *pgx.Conn, prometheusMetrics *observability.PrometheusMetrics) *Handlers {
	return &Handlers{
		db:                db,
		queries:           models.New(db),
		tracer:            otel.Tracer("inventory-service/handlers"),
		prometheusMetrics: prometheusMetrics,
	}
}

func (h *Handlers) GetInventory(ctx *gin.Context) {
	// Start a new span for this operation
	_, span := h.tracer.Start(ctx.Request.Context(), "GetInventory")
	defer span.End()

	// Get inventory ID from URL parameter
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid inventory ID",
		})
		return
	}

	// Add attributes to the span
	span.SetAttributes(attribute.Int64("inventory.id", id))

	dbStart := time.Now()
	inventory, err := h.queries.GetInventory(ctx, id)
	dbDuration := time.Since(dbStart)

	// Record database operation duration (Prometheus)
	if h.prometheusMetrics != nil {
		h.prometheusMetrics.RecordDBOperation("get", "inventory", dbDuration, err)
	}

	if err != nil {
		slog.Error("Got an error while getting inventory: ", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get inventory item",
		})
		return
	}

	// Record successful retrieval (Prometheus)
	if h.prometheusMetrics != nil {
		h.prometheusMetrics.RecordInventoryOperation("get", inventory.Category, inventory.Location)
	}

	// Record successful operation
	span.SetAttributes(
		attribute.String("inventory.name", inventory.Name),
		attribute.String("operation.status", "success"),
	)

	ctx.JSON(200, gin.H{
		"message": "Get Inventory Successfully",
		"data":    inventory,
	})
}

func (h *Handlers) ListInventory(ctx *gin.Context) {
	// Start a new span for this operation
	_, span := h.tracer.Start(ctx.Request.Context(), "ListInventory")
	defer span.End()

	// Add attributes to the span
	span.SetAttributes(
		attribute.Int("inventory.limit", 10),
		attribute.Int("inventory.offset", 0),
	)

	dbStart := time.Now()
	inventories, err := h.queries.ListInventory(ctx, models.ListInventoryParams{
		Limit:  10,
		Offset: 0,
	})
	dbDuration := time.Since(dbStart)

	// Record database operation duration (Prometheus)
	if h.prometheusMetrics != nil {
		h.prometheusMetrics.RecordDBOperation("list", "inventory", dbDuration, err)
	}

	if err != nil {
		slog.Error("Got an error while listing inventory: ", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list inventory items",
		})
		return
	}

	// Record successful list operation (Prometheus)
	if h.prometheusMetrics != nil {
		h.prometheusMetrics.RecordInventoryOperation("list", "all", "all")
	}

	span.SetAttributes(
		attribute.Int("inventory.count", len(inventories)),
		attribute.String("operation.status", "success"),
	)

	ctx.JSON(200, gin.H{
		"message": "List Inventory Successfully",
		"data":    inventories,
		"count":   len(inventories),
	})
}

func (h *Handlers) CreateInventory(ctx *gin.Context) {
	// Start a new span for this operation
	_, span := h.tracer.Start(ctx.Request.Context(), "CreateInventory")
	defer span.End()

	// Parse quantity from string to int32
	quantityStr := ctx.PostForm("Quantity")
	quantity, err := strconv.ParseInt(quantityStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid quantity format",
		})
		return
	}

	param := models.CreateInventoryParams{
		Name:     ctx.PostForm("Name"),
		Unit:     ctx.PostForm("Unit"),
		Quantity: int32(quantity),
		Measure:  ctx.PostForm("Measure"),
		Category: ctx.PostForm("Category"),
		Location: ctx.PostForm("Location"),
	}

	// Add attributes to the span
	span.SetAttributes(
		attribute.String("inventory.name", param.Name),
		attribute.String("inventory.category", param.Category),
		attribute.String("inventory.location", param.Location),
		attribute.Int("inventory.quantity", int(param.Quantity)),
	)

	dbStart := time.Now()
	inventory, err := h.queries.CreateInventory(ctx, param)
	dbDuration := time.Since(dbStart)

	// Record database operation duration (Prometheus)
	if h.prometheusMetrics != nil {
		h.prometheusMetrics.RecordDBOperation("create", "inventory", dbDuration, err)
	}

	if err != nil {
		slog.Error("Could not create inventory: ", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create inventory",
		})
		return
	}

	// Record successful creation (Prometheus)
	if h.prometheusMetrics != nil {
		h.prometheusMetrics.RecordInventoryOperation("create", inventory.Category, inventory.Location)
		// Update active inventory count (you'd need to query the total count or maintain it)
		// For now, we'll increment by 1 (in a real app, you'd want to track the actual count)
		h.prometheusMetrics.UpdateInventoryCount(1) // This should be the actual total count
	}

	// Record successful operation
	span.SetAttributes(
		attribute.Int64("inventory.id", inventory.ID),
		attribute.String("operation.status", "success"),
	)

	ctx.JSON(200, gin.H{
		"message": "Create Inventory Successfully",
		"data":    inventory,
	})
}

func (h *Handlers) UpdateInventory(ctx *gin.Context) {
	// Start a new span for this operation
	_, span := h.tracer.Start(ctx.Request.Context(), "UpdateInventory")
	defer span.End()

	// Get inventory ID from URL parameter
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid inventory ID",
		})
		return
	}

	// Parse quantity from string to int32
	quantityStr := ctx.PostForm("Quantity")
	quantity, err := strconv.ParseInt(quantityStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid quantity format",
		})
		return
	}

	// Add attributes to the span
	span.SetAttributes(
		attribute.Int64("inventory.id", id),
		attribute.String("inventory.name", ctx.PostForm("Name")),
		attribute.String("inventory.category", ctx.PostForm("Category")),
		attribute.String("inventory.location", ctx.PostForm("Location")),
		attribute.Int("inventory.quantity", int(quantity)),
	)

	// Start database transaction
	tx, err := h.db.Begin(ctx)
	if err != nil {
		slog.Error("Failed to start transaction", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback(ctx) // This will be ignored if tx.Commit() succeeds

	// Create queries with transaction
	qtx := h.queries.WithTx(tx)

	// Check if inventory exists before updating
	dbStart := time.Now()
	existingInventory, err := qtx.GetInventory(ctx, id)
	dbDuration := time.Since(dbStart)

	// Record database operation duration for existence check (Prometheus)
	if h.prometheusMetrics != nil {
		h.prometheusMetrics.RecordDBOperation("get_for_update", "inventory", dbDuration, err)
	}

	if err != nil {
		slog.Error("Inventory not found", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Inventory not found",
		})
		return
	}

	// Update inventory within transaction
	param := models.UpdateInventoryParams{
		ID:       id,
		Name:     ctx.PostForm("Name"),
		Unit:     ctx.PostForm("Unit"),
		Quantity: int32(quantity),
		Measure:  ctx.PostForm("Measure"),
		Category: ctx.PostForm("Category"),
		Location: ctx.PostForm("Location"),
	}

	dbStart = time.Now()
	inventory, err := qtx.UpdateInventory(ctx, param)
	dbDuration = time.Since(dbStart)

	// Record database operation duration for update (Prometheus)
	if h.prometheusMetrics != nil {
		h.prometheusMetrics.RecordDBOperation("update", "inventory", dbDuration, err)
	}

	if err != nil {
		slog.Error("Could not update inventory", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update inventory",
		})
		return
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		slog.Error("Failed to commit transaction", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	// Record successful update (Prometheus)
	if h.prometheusMetrics != nil {
		h.prometheusMetrics.RecordInventoryOperation("update", inventory.Category, inventory.Location)

		// Track location changes if different
		if existingInventory.Location != inventory.Location {
			h.prometheusMetrics.RecordInventoryOperation("location_change", inventory.Category, inventory.Location)
		}

		// Track category changes if different
		if existingInventory.Category != inventory.Category {
			h.prometheusMetrics.RecordInventoryOperation("category_change", inventory.Category, inventory.Location)
		}
	}

	// Record successful operation
	span.SetAttributes(
		attribute.String("operation.status", "success"),
	)

	ctx.JSON(200, gin.H{
		"message": "Update Inventory Successfully",
		"data":    inventory,
	})
}

func (h *Handlers) DeleteInventory(ctx *gin.Context) {
	_, span := h.tracer.Start(ctx.Request.Context(), "DeleteInventory")
	defer span.End()

	// Get inventory ID from URL parameter
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid inventory ID",
		})
		return
	}

	span.SetAttributes(attribute.Int64("inventory.id", id))

	dbStart := time.Now()
	err = h.queries.DeleteInventory(ctx, id)
	dbDuration := time.Since(dbStart)

	// Record database operation duration (Prometheus)
	if h.prometheusMetrics != nil {
		h.prometheusMetrics.RecordDBOperation("delete", "inventory", dbDuration, err)
	}

	if err != nil {
		slog.Error("Got an error while deleting inventory: ", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete inventory item",
		})
		return
	}

	// Record successful deletion (Prometheus)
	if h.prometheusMetrics != nil {
		h.prometheusMetrics.RecordInventoryOperation("delete", "unknown", "unknown")
	}

	span.SetAttributes(
		attribute.String("operation.status", "success"),
	)

	ctx.JSON(200, gin.H{
		"message": "Delete Inventory Successfully",
	})
}

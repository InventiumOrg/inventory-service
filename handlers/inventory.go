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
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	queries         *models.Queries
	tracer          trace.Tracer
	db              *pgx.Conn
	businessMetrics *observability.BusinessMetrics
}

func NewHandlers(db *pgx.Conn, businessMetrics *observability.BusinessMetrics) *Handlers {
	return &Handlers{
		db:              db,
		queries:         models.New(db),
		tracer:          otel.Tracer("inventory-service/handlers"),
		businessMetrics: businessMetrics,
	}
}

func (h *Handlers) GetInventory(ctx *gin.Context) {
	// Start a new span for this operation
	spanCtx, span := h.tracer.Start(ctx.Request.Context(), "GetInventory")
	defer span.End()

	// Record authentication attempt
	if h.businessMetrics != nil {
		h.businessMetrics.AuthenticationAttempts.Add(spanCtx, 1,
			metric.WithAttributes(attribute.String("operation", "get_inventory")))
	}

	_, existed := ctx.Get("claims")
	if !existed {
		if h.businessMetrics != nil {
			h.businessMetrics.AuthenticationAttempts.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "get_inventory"),
					attribute.String("status", "failed"),
				))
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Claims not found in context",
		})
		return
	}

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

	// Record business operation
	if h.businessMetrics != nil {
		h.businessMetrics.InventoryOperations.Add(spanCtx, 1,
			metric.WithAttributes(
				attribute.String("operation", "get"),
				attribute.Int64("inventory.id", id),
			))
	}

	dbStart := time.Now()
	inventory, err := h.queries.GetInventory(ctx, id)
	dbDuration := time.Since(dbStart).Seconds()

	// Record database operation duration
	if h.businessMetrics != nil {
		h.businessMetrics.DBOperationDuration.Record(spanCtx, dbDuration,
			metric.WithAttributes(
				attribute.String("operation", "get_inventory"),
				attribute.String("table", "inventory"),
			))
	}

	if err != nil {
		// Record database error
		if h.businessMetrics != nil {
			h.businessMetrics.DBOperationErrors.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "get_inventory"),
					attribute.String("error_type", "query_failed"),
				))
		}

		slog.Error("Got an error while getting inventory: ", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get inventory item",
		})
		return
	}

	// Record successful retrieval
	if h.businessMetrics != nil {
		h.businessMetrics.InventoryRetrievals.Add(spanCtx, 1,
			metric.WithAttributes(
				attribute.String("inventory.name", inventory.Name),
				attribute.String("status", "success"),
			))
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
	spanCtx, span := h.tracer.Start(ctx.Request.Context(), "ListInventory")
	defer span.End()

	// Record list request
	if h.businessMetrics != nil {
		h.businessMetrics.InventoryListRequests.Add(spanCtx, 1)
	}

	_, existed := ctx.Get("claims")
	if !existed {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Claims not found in context",
		})
		return
	}

	// Add attributes to the span
	span.SetAttributes(
		attribute.Int("inventory.limit", 10),
		attribute.Int("inventory.offset", 0),
	)

	dbStart := time.Now()
	inventories, err := h.queries.ListInventory(spanCtx, models.ListInventoryParams{
		Limit:  10,
		Offset: 0,
	})
	dbDuration := time.Since(dbStart).Seconds()

	// Record database operation duration
	if h.businessMetrics != nil {
		h.businessMetrics.DBOperationDuration.Record(spanCtx, dbDuration,
			metric.WithAttributes(
				attribute.String("operation", "list_inventory"),
				attribute.String("table", "inventory"),
			))
	}

	if err != nil {
		if h.businessMetrics != nil {
			h.businessMetrics.DBOperationErrors.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "list_inventory"),
					attribute.String("error_type", "query_failed"),
				))
		}

		slog.Error("Got an error while listing inventory: ", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list inventory items",
		})
		return
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
	spanCtx, span := h.tracer.Start(ctx.Request.Context(), "CreateInventory")
	defer span.End()

	// Record authentication attempt
	if h.businessMetrics != nil {
		h.businessMetrics.AuthenticationAttempts.Add(spanCtx, 1,
			metric.WithAttributes(attribute.String("operation", "create_inventory")))
	}

	_, existed := ctx.Get("claims")
	if !existed {
		if h.businessMetrics != nil {
			h.businessMetrics.AuthenticationAttempts.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "create_inventory"),
					attribute.String("status", "failed"),
				))
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Claims not found in context",
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

	// Record business operation
	if h.businessMetrics != nil {
		h.businessMetrics.InventoryOperations.Add(spanCtx, 1,
			metric.WithAttributes(
				attribute.String("operation", "create"),
				attribute.String("inventory.category", param.Category),
				attribute.String("inventory.location", param.Location),
			))
	}

	dbStart := time.Now()
	inventory, err := h.queries.CreateInventory(ctx, param)
	dbDuration := time.Since(dbStart).Seconds()

	// Record database operation duration
	if h.businessMetrics != nil {
		h.businessMetrics.DBOperationDuration.Record(spanCtx, dbDuration,
			metric.WithAttributes(
				attribute.String("operation", "create_inventory"),
				attribute.String("table", "inventory"),
			))
	}

	if err != nil {
		// Record database error
		if h.businessMetrics != nil {
			h.businessMetrics.DBOperationErrors.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "create_inventory"),
					attribute.String("error_type", "insert_failed"),
				))
		}

		slog.Error("Could not create inventory: ", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create inventory",
		})
		return
	}

	// Record successful creation
	if h.businessMetrics != nil {
		h.businessMetrics.InventoryCreated.Add(spanCtx, 1,
			metric.WithAttributes(
				attribute.String("inventory.name", inventory.Name),
				attribute.String("inventory.category", inventory.Category),
				attribute.String("inventory.location", inventory.Location),
			))

		// Track inventory by location
		h.businessMetrics.InventoryByLocation.Add(spanCtx, 1,
			metric.WithAttributes(
				attribute.String("location", inventory.Location),
			))

		// Track inventory by category
		h.businessMetrics.InventoryByCategory.Add(spanCtx, 1,
			metric.WithAttributes(
				attribute.String("category", inventory.Category),
			))

		// Increase active inventory count
		h.businessMetrics.ActiveInventoryItems.Add(spanCtx, 1)
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
	spanCtx, span := h.tracer.Start(ctx.Request.Context(), "UpdateInventory")
	defer span.End()

	// Record authentication attempt
	if h.businessMetrics != nil {
		h.businessMetrics.AuthenticationAttempts.Add(spanCtx, 1,
			metric.WithAttributes(attribute.String("operation", "update_inventory")))
	}

	_, existed := ctx.Get("claims")
	if !existed {
		if h.businessMetrics != nil {
			h.businessMetrics.AuthenticationAttempts.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "update_inventory"),
					attribute.String("status", "failed"),
				))
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Claims not found in context",
		})
		return
	}

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

	// Record business operation
	if h.businessMetrics != nil {
		h.businessMetrics.InventoryOperations.Add(spanCtx, 1,
			metric.WithAttributes(
				attribute.String("operation", "update"),
				attribute.Int64("inventory.id", id),
				attribute.String("inventory.category", ctx.PostForm("Category")),
				attribute.String("inventory.location", ctx.PostForm("Location")),
			))
	}

	// Start database transaction
	tx, err := h.db.Begin(ctx)
	if err != nil {
		if h.businessMetrics != nil {
			h.businessMetrics.DBOperationErrors.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "update_inventory"),
					attribute.String("error_type", "transaction_start_failed"),
				))
		}
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
	dbDuration := time.Since(dbStart).Seconds()

	// Record database operation duration for existence check
	if h.businessMetrics != nil {
		h.businessMetrics.DBOperationDuration.Record(spanCtx, dbDuration,
			metric.WithAttributes(
				attribute.String("operation", "get_inventory_for_update"),
				attribute.String("table", "inventory"),
			))
	}

	if err != nil {
		if h.businessMetrics != nil {
			h.businessMetrics.DBOperationErrors.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "get_inventory_for_update"),
					attribute.String("error_type", "not_found"),
				))
		}
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
	dbDuration = time.Since(dbStart).Seconds()

	// Record database operation duration for update
	if h.businessMetrics != nil {
		h.businessMetrics.DBOperationDuration.Record(spanCtx, dbDuration,
			metric.WithAttributes(
				attribute.String("operation", "update_inventory"),
				attribute.String("table", "inventory"),
			))
	}

	if err != nil {
		if h.businessMetrics != nil {
			h.businessMetrics.DBOperationErrors.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "update_inventory"),
					attribute.String("error_type", "update_failed"),
				))
		}
		slog.Error("Could not update inventory", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update inventory",
		})
		return
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		if h.businessMetrics != nil {
			h.businessMetrics.DBOperationErrors.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "update_inventory"),
					attribute.String("error_type", "transaction_commit_failed"),
				))
		}
		slog.Error("Failed to commit transaction", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	// Record successful update
	if h.businessMetrics != nil {
		h.businessMetrics.InventoryUpdates.Add(spanCtx, 1,
			metric.WithAttributes(
				attribute.String("inventory.name", inventory.Name),
				attribute.String("inventory.category", inventory.Category),
				attribute.String("inventory.location", inventory.Location),
				attribute.Int64("inventory.id", inventory.ID),
			))

		// Track location changes if different
		if existingInventory.Location != inventory.Location {
			// Decrease count for old location
			h.businessMetrics.InventoryByLocation.Add(spanCtx, -1,
				metric.WithAttributes(
					attribute.String("location", existingInventory.Location),
				))
			// Increase count for new location
			h.businessMetrics.InventoryByLocation.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("location", inventory.Location),
				))
		}

		// Track category changes if different
		if existingInventory.Category != inventory.Category {
			// Decrease count for old category
			h.businessMetrics.InventoryByCategory.Add(spanCtx, -1,
				metric.WithAttributes(
					attribute.String("category", existingInventory.Category),
				))
			// Increase count for new category
			h.businessMetrics.InventoryByCategory.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("category", inventory.Category),
				))
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
	spanCtx, span := h.tracer.Start(ctx.Request.Context(), "DeleteInventory")
	defer span.End()

	_, existed := ctx.Get("claims")
	if !existed {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Claims not found in context",
		})
		return
	}

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
	dbDuration := time.Since(dbStart).Seconds()

	// Record database operation duration
	if h.businessMetrics != nil {
		h.businessMetrics.DBOperationDuration.Record(spanCtx, dbDuration,
			metric.WithAttributes(
				attribute.String("operation", "delete_inventory"),
				attribute.String("table", "inventory"),
			))
	}

	if err != nil {
		if h.businessMetrics != nil {
			h.businessMetrics.DBOperationErrors.Add(spanCtx, 1,
				metric.WithAttributes(
					attribute.String("operation", "delete_inventory"),
					attribute.String("error_type", "delete_failed"),
				))
		}

		slog.Error("Got an error while deleting inventory: ", slog.Any("err", err.Error()))
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete inventory item",
		})
		return
	}

	// Record successful deletion
	if h.businessMetrics != nil {
		h.businessMetrics.InventoryDeletes.Add(spanCtx, 1)
		// Decrease active inventory count
		h.businessMetrics.ActiveInventoryItems.Add(spanCtx, -1)
	}

	span.SetAttributes(
		attribute.String("operation.status", "success"),
	)

	ctx.JSON(200, gin.H{
		"message": "Delete Inventory Successfully",
	})
}

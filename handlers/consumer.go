package handlers

import (
	"context"
	"fmt"
	"inventory-service/models"
	models_sqlc "inventory-service/models/sqlc"
	"inventory-service/utils"
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type Consumer struct {
	queries *models_sqlc.Queries
	db      *pgx.Conn
}

func NewHandler(db *pgx.Conn, queries *models_sqlc.Queries) *Consumer {
	consumer := &Consumer{
		queries: queries,
		db:      db,
	}
	return consumer
}

// ProcessInventoryUpdateEvent processes a parsed inventory event
func (c *Consumer) ProcessInventoryUpdateEvent(event *models.InventoryImportEvent) error {
	slog.Info("Processing inventory update event",
		"inventoryId", event.InventoryID,
		"quantity", event.Quantity,
		"measure", event.InventoryMeasure,
		"category", event.InventoryCategory,
		"unit", event.InventoryUnit,
		"type", event.Type)

	// Convert inventory ID to int64
	inventoryID, err := strconv.ParseInt(event.InventoryID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse inventory ID: %w", err)
	}

	// Get existing inventory
	inventory, err := c.queries.GetInventory(context.Background(), inventoryID)
	if err != nil {
		slog.Error("Failed to find the inventory item", "inventoryId", inventoryID, "error", err)
		return fmt.Errorf("failed to get inventory: %w", err)
	}
	quantity := utils.ProcessInventory(inventory.Quantity, int32(event.Quantity), event.Type)
	// Update inventory with event data
	_, err = c.queries.UpdateInventory(context.Background(), models_sqlc.UpdateInventoryParams{
		ID:       inventoryID,
		Name:     inventory.Name, // Keep existing name
		Unit:     event.InventoryUnit,
		Quantity: quantity,
		Measure:  event.InventoryMeasure,
		Category: event.InventoryCategory,
		Location: inventory.Location, // Keep existing location
	})

	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	slog.Info("Successfully updated inventory", "inventoryId", inventoryID, "type", event.Type)
	return nil
}

package models

// InventoryImportEvent represents an inventory import event from Kafka
type InventoryImportEvent struct {
	Quantity          int    `json:"quantity"`
	InventoryID       string `json:"inventoryId"`
	InventoryMeasure  string `json:"inventoryMeasure"`
	InventoryCategory string `json:"inventoryCategory"`
	InventoryUnit     string `json:"inventoryUnit"`
	Type              string `json:"type"`
}

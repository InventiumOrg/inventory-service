package processors

import (
	"encoding/binary"
	"fmt"
	"inventory-service/config"
	"inventory-service/models"

	"github.com/linkedin/goavro/v2"
)

type AvroDeserializer struct {
	schemaRegistryURL string
	schemaCache       map[int]*goavro.Codec
}

func NewAvroDeserializer(config config.Config) (*AvroDeserializer, error) {
	return &AvroDeserializer{
		schemaRegistryURL: config.SchemaRegistryURL,
		schemaCache:       make(map[int]*goavro.Codec),
	}, nil
}

func (ad *AvroDeserializer) DeserializeMessage(data []byte) (*models.InventoryImportEvent, error) {
	// Check if message has Schema Registry format (magic byte + schema ID)
	if len(data) < 5 {
		return nil, fmt.Errorf("message too short to contain schema registry header")
	}

	// First byte should be 0 (magic byte)
	if data[0] != 0 {
		return nil, fmt.Errorf("invalid magic byte, expected 0 but got %d", data[0])
	}

	// Next 4 bytes contain the schema ID
	schemaID := int(binary.BigEndian.Uint32(data[1:5]))
	fmt.Printf("Deserializing message with schema ID: %d\n", schemaID)

	// Get or fetch the schema
	codec, err := ad.getSchema(schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema for ID %d: %w", schemaID, err)
	}

	// Deserialize the Avro data (skip the first 5 bytes)
	avroData := data[5:]
	native, _, err := codec.NativeFromBinary(avroData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize avro message: %w", err)
	}

	// Convert to our struct
	event, err := ad.mapToInventoryEvent(native)
	if err != nil {
		return nil, fmt.Errorf("failed to map avro result to struct: %w", err)
	}

	return event, nil
}

func (ad *AvroDeserializer) getSchema(schemaID int) (*goavro.Codec, error) {
	// Check cache first
	if codec, exists := ad.schemaCache[schemaID]; exists {
		return codec, nil
	}

	// Create schema registry client
	client := NewSchemaRegistryClient(ad.schemaRegistryURL, "", "")

	// Fetch schema from registry
	schemaJSON, err := client.GetSchema(schemaID)
	if err != nil {
		// Fallback to hardcoded schema if registry is unavailable
		fmt.Printf("Failed to fetch schema from registry, using fallback: %v\n", err)
		schemaJSON = `{
			"type": "record",
			"name": "inventory_import",
			"namespace": "com.inventium",
			"fields": [
				{"name": "quantity", "type": "int", "default": 5},
				{"name": "inventoryId", "type": "string", "default": "001"},
				{"name": "inventoryMeasure", "type": "string", "default": "kg"},
				{"name": "inventoryCategory", "type": "string", "default": "milk"},
				{"name": "inventoryUnit", "type": "string", "default": "box"},
				{"name": "type", "type": "string", "default": ""}
			]
		}`
	}

	// Create codec from schema
	codec, err := goavro.NewCodec(schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create codec from schema: %w", err)
	}

	// Cache the codec
	ad.schemaCache[schemaID] = codec

	return codec, nil
}

func (ad *AvroDeserializer) mapToInventoryEvent(data interface{}) (*models.InventoryImportEvent, error) {
	// The result from Avro deserializer is typically a map[string]interface{}
	record, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected map[string]interface{}, got %T", data)
	}

	event := &models.InventoryImportEvent{}

	// Extract fields with type checking and defaults
	if quantity, exists := record["quantity"]; exists {
		switch q := quantity.(type) {
		case int32:
			event.Quantity = int(q)
		case int64:
			event.Quantity = int(q)
		case int:
			event.Quantity = q
		default:
			event.Quantity = 5 // default value
		}
	} else {
		event.Quantity = 5
	}

	if inventoryID, exists := record["inventoryId"]; exists {
		if id, ok := inventoryID.(string); ok {
			event.InventoryID = id
		} else {
			event.InventoryID = "001" // default value
		}
	} else {
		event.InventoryID = "001"
	}

	if measure, exists := record["inventoryMeasure"]; exists {
		if m, ok := measure.(string); ok {
			event.InventoryMeasure = m
		} else {
			event.InventoryMeasure = "kg" // default value
		}
	} else {
		event.InventoryMeasure = "kg"
	}

	if category, exists := record["inventoryCategory"]; exists {
		if c, ok := category.(string); ok {
			event.InventoryCategory = c
		} else {
			event.InventoryCategory = "milk" // default value
		}
	} else {
		event.InventoryCategory = "milk"
	}

	if unit, exists := record["inventoryUnit"]; exists {
		if u, ok := unit.(string); ok {
			event.InventoryUnit = u
		} else {
			event.InventoryUnit = "box" // default value
		}
	} else {
		event.InventoryUnit = "box"
	}

	if typ, exists := record["type"]; exists {
		if t, ok := typ.(string); ok {
			event.Type = t
		} else {
			event.Type = "" // default value
		}
	} else {
		event.Type = ""
	}

	return event, nil
}

func (ad *AvroDeserializer) Close() {
	// Clean up resources if needed
	ad.schemaCache = nil
}

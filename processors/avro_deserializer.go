package processors

import (
	"encoding/binary"
	"fmt"
	"inventory-service/config"
	"inventory-service/models"
	"net/url"
	"strings"

	"github.com/linkedin/goavro/v2"
)

// inventoryTransactionFallbackSchema matches Schema Registry subject
// inventory.transaction.status.updated-value (schema id 4). Field order must match
// the writer schema — Avro binary encoding is order-sensitive.
const inventoryTransactionFallbackSchema = `{
	"type": "record",
	"name": "inventory_transaction",
	"doc": "inventory_transaction",
	"fields": [
		{"name": "type", "type": "string"},
		{"name": "quantity", "type": "int", "default": 5},
		{"name": "inventoryId", "type": "string", "default": "1"},
		{"name": "inventoryMeasure", "type": "string", "default": "kg"},
		{"name": "inventoryCategory", "type": "string", "default": "milk"},
		{"name": "inventoryUnit", "type": "string", "default": "box"}
	]
}`

type AvroDeserializer struct {
	schemaRegistryURL      string
	schemaRegistryUsername string
	schemaRegistryPassword string
	schemaCache            map[int]*goavro.Codec
}

func NewAvroDeserializer(cfg config.Config) (*AvroDeserializer, error) {
	baseURL, username, password, err := normalizeSchemaRegistryURL(
		cfg.SchemaRegistryURL,
		cfg.SchemaRegistryUsername,
		cfg.SchemaRegistryPassword,
	)
	if err != nil {
		return nil, err
	}

	return &AvroDeserializer{
		schemaRegistryURL:      baseURL,
		schemaRegistryUsername: username,
		schemaRegistryPassword: password,
		schemaCache:            make(map[int]*goavro.Codec),
	}, nil
}

func normalizeSchemaRegistryURL(rawURL, username, password string) (baseURL, user, pass string, err error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid schema registry URL: %w", err)
	}

	if parsed.User != nil {
		if username == "" {
			username = parsed.User.Username()
		}
		if password == "" {
			password, _ = parsed.User.Password()
		}
		parsed.User = nil
	}

	return strings.TrimRight(parsed.String(), "/"), username, password, nil
}

func (ad *AvroDeserializer) DeserializeMessage(data []byte) (*models.InventoryImportEvent, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("message too short to contain schema registry header")
	}

	if data[0] != 0 {
		return nil, fmt.Errorf("invalid magic byte, expected 0 but got %d", data[0])
	}

	schemaID := int(binary.BigEndian.Uint32(data[1:5]))
	fmt.Printf("Deserializing message with schema ID: %d\n", schemaID)

	codec, err := ad.getSchema(schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema for ID %d: %w", schemaID, err)
	}

	avroData := data[5:]
	native, _, err := codec.NativeFromBinary(avroData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize avro message: %w", err)
	}

	event, err := ad.mapToInventoryEvent(native)
	if err != nil {
		return nil, fmt.Errorf("failed to map avro result to struct: %w", err)
	}

	return event, nil
}

func (ad *AvroDeserializer) getSchema(schemaID int) (*goavro.Codec, error) {
	if codec, exists := ad.schemaCache[schemaID]; exists {
		return codec, nil
	}

	client := NewSchemaRegistryClient(
		ad.schemaRegistryURL,
		ad.schemaRegistryUsername,
		ad.schemaRegistryPassword,
	)

	schemaJSON, fromRegistry, err := ad.fetchSchemaJSON(client, schemaID)
	if err != nil {
		return nil, err
	}

	codec, err := goavro.NewCodec(schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create codec from schema: %w", err)
	}

	// Only cache schemas loaded from the registry so a bad fallback is never
	// pinned after a transient 401 or wrong field order.
	if fromRegistry {
		ad.schemaCache[schemaID] = codec
	}

	return codec, nil
}

func (ad *AvroDeserializer) fetchSchemaJSON(client *SchemaRegistryClient, schemaID int) (string, bool, error) {
	schemaJSON, err := client.GetSchema(schemaID)
	if err == nil {
		return schemaJSON, true, nil
	}

	fmt.Printf("Failed to fetch schema %d from registry (%s): %v\n", schemaID, ad.schemaRegistryURL, err)

	// Auth failures must not fall back to a guessed schema — fix credentials instead.
	if strings.Contains(err.Error(), "status 401") {
		return "", false, fmt.Errorf(
			"schema registry unauthorized for schema ID %d: set SCHEMA_REGISTRY_USERNAME and SCHEMA_REGISTRY_PASSWORD (or embed credentials in SCHEMA_REGISTRY_URL): %w",
			schemaID,
			err,
		)
	}

	fmt.Printf("Using local fallback schema for schema ID %d\n", schemaID)
	return inventoryTransactionFallbackSchema, false, nil
}

func (ad *AvroDeserializer) mapToInventoryEvent(data interface{}) (*models.InventoryImportEvent, error) {
	record, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected map[string]interface{}, got %T", data)
	}

	event := &models.InventoryImportEvent{}

	if quantity, exists := record["quantity"]; exists {
		switch q := quantity.(type) {
		case int32:
			event.Quantity = int(q)
		case int64:
			event.Quantity = int(q)
		case int:
			event.Quantity = q
		default:
			event.Quantity = 5
		}
	} else {
		event.Quantity = 5
	}

	if inventoryID, exists := record["inventoryId"]; exists {
		if id, ok := inventoryID.(string); ok {
			event.InventoryID = id
		} else {
			event.InventoryID = "1"
		}
	} else {
		event.InventoryID = "1"
	}

	if measure, exists := record["inventoryMeasure"]; exists {
		if m, ok := measure.(string); ok {
			event.InventoryMeasure = m
		} else {
			event.InventoryMeasure = "kg"
		}
	} else {
		event.InventoryMeasure = "kg"
	}

	if category, exists := record["inventoryCategory"]; exists {
		if c, ok := category.(string); ok {
			event.InventoryCategory = c
		} else {
			event.InventoryCategory = "milk"
		}
	} else {
		event.InventoryCategory = "milk"
	}

	if unit, exists := record["inventoryUnit"]; exists {
		if u, ok := unit.(string); ok {
			event.InventoryUnit = u
		} else {
			event.InventoryUnit = "box"
		}
	} else {
		event.InventoryUnit = "box"
	}

	if typ, exists := record["type"]; exists {
		if t, ok := typ.(string); ok {
			event.Type = t
		}
	}

	return event, nil
}

func (ad *AvroDeserializer) Close() {
	ad.schemaCache = nil
}

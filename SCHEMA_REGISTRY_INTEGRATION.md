# Schema Registry Integration

This document explains how the inventory service integrates with Confluent Schema Registry for Avro message deserialization.

## Architecture Overview

```
Kafka Message → Processor → Avro Deserializer → Schema Registry → Handler → Database
```

## Components

### 1. Schema Registry Client (`processors/schema_registry_client.go`)
- Fetches schemas from Schema Registry by schema ID
- Supports basic authentication
- Returns raw schema JSON

### 2. Avro Deserializer (`processors/avro_deserializer.go`)
- Parses Schema Registry wire format: `[magic_byte][schema_id][avro_data]`
- Fetches schemas dynamically from Schema Registry
- Caches schemas in memory to avoid repeated registry calls
- Falls back to hardcoded schema if registry is unavailable
- Deserializes Avro binary data to Go structs

### 3. Inventory Processor (`processors/consumer.go`)
- Consumes messages from Kafka
- Uses Avro deserializer to parse messages
- Passes parsed events to handler
- Handles commit/error logic for exactly-once processing

### 4. Consumer Handler (`handlers/consumer.go`)
- Receives parsed `InventoryImportEvent`
- Updates database with event data
- Business logic for inventory updates

### 5. Event Model (`models/events.go`)
- Shared event structure used across all components
- Includes all fields from Avro schema

## Message Flow

### 1. Message Arrives from Kafka
```
[0x00][Schema ID: 4 bytes][Avro Binary Data]
```

### 2. Deserializer Extracts Schema ID
```go
schemaID := binary.BigEndian.Uint32(data[1:5])
// Example: schemaID = 123
```

### 3. Fetch Schema from Registry
```
GET http://schema-registry:8081/schemas/ids/123

Response:
{
  "schema": "{\"type\":\"record\",\"name\":\"inventory_import\",...}"
}
```

### 4. Deserialize Avro Data
```go
codec := goavro.NewCodec(schemaJSON)
native, _, err := codec.NativeFromBinary(avroData)
```

### 5. Map to Go Struct
```go
event := &models.InventoryImportEvent{
    Quantity:          5,
    InventoryID:       "123",
    InventoryMeasure:  "kg",
    InventoryCategory: "food",
    InventoryUnit:     "box",
    Type:              "import",
}
```

### 6. Process Event
```go
handler.ProcessInventoryUpdateEvent(event)
// Updates database
```

### 7. Commit Message
```go
consumer.CommitMessages(ctx, message)
// Marks message as processed
```

## Configuration

### Environment Variables
```env
# Kafka Configuration
KAFKA_URI="kafka-broker:9092"
KAFKA_CONSUMER_GROUP="inventory-service-group"
TOPIC_NAME="inventory.import.status.updated"

# Schema Registry Configuration
SCHEMA_REGISTRY_URL="http://schema-registry:8081"
SCHEMA_REGISTRY_USERNAME=""  # Optional
SCHEMA_REGISTRY_PASSWORD=""  # Optional
```

## Schema Evolution

The Schema Registry integration supports schema evolution:

1. **Backward Compatible**: New optional fields can be added
2. **Forward Compatible**: Old consumers can read new messages
3. **Full Compatible**: Both backward and forward compatible

### Example Schema Evolution

**Version 1:**
```json
{
  "type": "record",
  "name": "inventory_import",
  "fields": [
    {"name": "quantity", "type": "int"},
    {"name": "inventoryId", "type": "string"}
  ]
}
```

**Version 2 (adds type field):**
```json
{
  "type": "record",
  "name": "inventory_import",
  "fields": [
    {"name": "quantity", "type": "int"},
    {"name": "inventoryId", "type": "string"},
    {"name": "type", "type": "string", "default": ""}
  ]
}
```

The deserializer automatically handles both versions!

## Benefits

✅ **No Hardcoded Schemas**: Schemas fetched dynamically from registry
✅ **Schema Evolution**: Automatic handling of schema changes
✅ **Performance**: In-memory schema caching
✅ **Resilience**: Fallback schema if registry is unavailable
✅ **Type Safety**: Strong typing with Go structs
✅ **Exactly-Once Processing**: Proper commit handling prevents duplicates

## Error Handling

### Schema Registry Unavailable
- Falls back to hardcoded schema
- Logs warning message
- Continues processing with fallback

### Invalid Message Format
- Logs error
- Skips message (doesn't commit)
- Message will be reprocessed

### Processing Failure
- Logs error
- Doesn't commit message
- Message will be reprocessed on next poll

## Monitoring

Key metrics to monitor:
- Schema fetch success/failure rate
- Schema cache hit rate
- Message processing latency
- Commit success rate
- Error rate by error type

## Testing

### Test with Sample Message
```bash
# Produce a test message with Schema Registry format
# (requires kafka-avro-console-producer)
kafka-avro-console-producer \
  --broker-list kafka:9092 \
  --topic inventory.import.status.updated \
  --property value.schema='{"type":"record","name":"inventory_import","fields":[...]}' \
  --property schema.registry.url=http://schema-registry:8081
```

### Verify Processing
```bash
# Check logs for deserialization
docker logs inventory-service | grep "Deserializing message with schema ID"

# Check database for updates
psql -h localhost -U root -d inventory-service -c "SELECT * FROM inventory WHERE id = 123;"
```

## Troubleshooting

### Schema Not Found
```
Error: schema registry returned status 404 for schema ID 123
```
**Solution**: Ensure schema is registered in Schema Registry

### Invalid Magic Byte
```
Error: invalid magic byte, expected 0 but got 1
```
**Solution**: Message is not in Schema Registry format. Check producer configuration.

### Deserialization Failed
```
Error: failed to deserialize avro message
```
**Solution**: Schema mismatch. Verify schema ID and schema content match.

## Future Improvements

1. **Schema Registry Authentication**: Add support for API keys
2. **Schema Caching TTL**: Add expiration for cached schemas
3. **Metrics**: Add Prometheus metrics for monitoring
4. **Schema Validation**: Validate messages against expected schema version
5. **Dead Letter Queue**: Send failed messages to DLQ for manual review

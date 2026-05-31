package processors

import (
	"encoding/binary"
	"testing"

	"github.com/linkedin/goavro/v2"
)

func TestDeserializeMessage_roundTrip(t *testing.T) {
	codec, err := goavro.NewCodec(inventoryTransactionFallbackSchema)
	if err != nil {
		t.Fatalf("codec: %v", err)
	}

	native := map[string]interface{}{
		"type":              "import",
		"quantity":          int32(10),
		"inventoryId":       "42",
		"inventoryMeasure":  "kg",
		"inventoryCategory": "milk",
		"inventoryUnit":     "box",
	}

	avroPayload, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	wire := make([]byte, 5+len(avroPayload))
	wire[0] = 0
	binary.BigEndian.PutUint32(wire[1:5], 4)
	copy(wire[5:], avroPayload)

	ad := &AvroDeserializer{
		schemaRegistryURL:      "http://invalid.example",
		schemaRegistryUsername: "",
		schemaRegistryPassword: "",
		schemaCache:            make(map[int]*goavro.Codec),
	}

	event, err := ad.DeserializeMessage(wire)
	if err != nil {
		t.Fatalf("deserialize: %v", err)
	}

	if event.Type != "import" {
		t.Errorf("type: got %q want import", event.Type)
	}
	if event.Quantity != 10 {
		t.Errorf("quantity: got %d want 10", event.Quantity)
	}
	if event.InventoryID != "42" {
		t.Errorf("inventoryId: got %q want 42", event.InventoryID)
	}
}

func TestNormalizeSchemaRegistryURL_stripsEmbeddedCredentials(t *testing.T) {
	base, user, pass, err := normalizeSchemaRegistryURL(
		"https://avnadmin:secret@registry.example.com:16682",
		"",
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	if user != "avnadmin" || pass != "secret" {
		t.Fatalf("credentials: user=%q pass=%q", user, pass)
	}
	if base != "https://registry.example.com:16682" {
		t.Fatalf("base URL: got %q", base)
	}
}

package processors

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"inventory-service/config"
	"inventory-service/handlers"

	models "inventory-service/models/sqlc"
	// "inventory-service/observability"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type InventoryProcessor struct {
	consumer         *kafka.Reader
	dialer           *kafka.Dialer
	con              *pgx.Conn
	handlers         *handlers.Consumer
	avroDeserializer *AvroDeserializer
}

func NewInventoryProcessor(con *pgx.Conn, consumer *kafka.Reader, dialer *kafka.Dialer) *InventoryProcessor {
	inventoryProcessor := &InventoryProcessor{
		consumer: consumer,
		dialer:   dialer,
		con:      con,
		handlers: handlers.NewHandler(con, models.New(con)),
	}
	return inventoryProcessor
}

func (p *InventoryProcessor) Init(config config.Config) error {

	// Initialize Avro deserializer
	avroDeserializer, err := NewAvroDeserializer(config)
	if err != nil {
		return fmt.Errorf("failed to initialize avro deserializer: %w", err)
	}
	p.avroDeserializer = avroDeserializer

	// TOPIC_NAME := "inventory.import.status.updated"

	caCert, err := os.ReadFile(config.KafkaCAFilePath)
	if err != nil {
		log.Fatalf("Failed to read CA certificate file: %s", err)
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		log.Fatalf("Failed to parse CA certificate file: %s", err)
	}

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	scram, err := scram.Mechanism(scram.SHA512, config.KafkaUsername, config.KafkaPassword)
	if err != nil {
		log.Fatalf("Failed to create scram mechanism: %s", err)
	}

	p.dialer = &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		TLS:           tlsConfig,
		SASLMechanism: scram,
	}

	// init consumer
	p.consumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{config.KafkaUri},
		Topic:       config.TopicName,
		GroupID:     config.KafkaConsumerGroup, // Essential for tracking processed messages
		Dialer:      p.dialer,
		StartOffset: kafka.FirstOffset, // Start from latest messages on first run
		MinBytes:    10e3,              // 10KB
		MaxBytes:    10e6,              // 10MB
	})
	return nil
}

func (p *InventoryProcessor) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Consumer context cancelled, shutting down...")
			return
		default:
			// Fetch message with context
			message, err := p.consumer.FetchMessage(ctx)
			if err != nil {
				fmt.Printf("Could not read message: %s\n", err)
				continue
			}

			fmt.Printf("Processing message: %d bytes\n", len(message.Value))

			// Deserialize Avro message using Schema Registry
			event, err := p.avroDeserializer.DeserializeMessage(message.Value)
			if err != nil {
				fmt.Printf("Failed to deserialize Avro message: %s\n", err)
				continue
			}

			fmt.Printf("Deserialized event: ID=%s, Quantity=%d, Category=%s, Type=%s\n",
				event.InventoryID, event.Quantity, event.InventoryCategory, event.Type)

			// Process the parsed event
			err = p.handlers.ProcessInventoryUpdateEvent(event)
			if err != nil {
				fmt.Printf("Failed to process message: %s\n", err)
				// Don't commit on error - message will be reprocessed
				continue
			}

			// Only commit after successful processing
			err = p.consumer.CommitMessages(ctx, message)
			if err != nil {
				fmt.Printf("Failed to commit message: %s\n", err)
				// Message was processed but commit failed - this could cause reprocessing
				// Consider implementing idempotency in your handlers
			} else {
				fmt.Printf("Successfully processed and committed message\n")
			}
		}
	}
}
func (p *InventoryProcessor) Close() error {
	if p.avroDeserializer != nil {
		p.avroDeserializer.Close()
	}
	if p.consumer != nil {
		return p.consumer.Close()
	}
	return nil
}

package producer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
)

// KafkaProducer implements Producer using segmentio/kafka-go.
type KafkaProducer struct {
	writer *kafka.Writer
	topic  string
}

// NewKafkaProducer creates a Kafka producer that writes telemetry events to the given topic.
// brokers must be non-empty. Call Close when shutting down.
func NewKafkaProducer(brokers []string, topic string) (*KafkaProducer, error) {
	if len(brokers) == 0 || topic == "" {
		return nil, nil
	}
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 50 * time.Millisecond,
	}
	return &KafkaProducer{writer: writer, topic: topic}, nil
}

// Emit serializes the event as JSON and writes it to the Kafka topic.
// Uses the request context with a short timeout so slow Kafka does not block callers indefinitely.
func (p *KafkaProducer) Emit(ctx context.Context, event *telemetryv1.TelemetryEvent) error {
	if p == nil || p.writer == nil || event == nil {
		return nil
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = p.writer.WriteMessages(writeCtx, kafka.Message{
		Key:   nil,
		Value: payload,
	})
	if err != nil {
		log.Printf("telemetry: kafka emit failed: %v", err)
		return err
	}
	return nil
}

// Close closes the Kafka writer. Safe to call multiple times.
func (p *KafkaProducer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

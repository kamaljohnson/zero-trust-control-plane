// Worker consumes telemetry events from Kafka and pushes them to Loki.
// Set KAFKA_BROKERS, TELEMETRY_KAFKA_TOPIC, KAFKA_GROUP_ID, and LOKI_URL. GRPC_ADDR is required by config but unused (e.g. set to :0).
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"

	"zero-trust-control-plane/backend/internal/config"
	"zero-trust-control-plane/backend/internal/telemetry/loki"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	brokers := cfg.TelemetryKafkaBrokersList()
	if len(brokers) == 0 {
		log.Fatal("worker: KAFKA_BROKERS is required")
	}
	if cfg.LokiURL == "" {
		log.Fatal("worker: LOKI_URL is required")
	}

	topic := cfg.TelemetryKafkaTopic
	if topic == "" {
		topic = "ztcp-telemetry"
	}
	groupID := cfg.KafkaGroupID
	if groupID == "" {
		groupID = "ztcp-telemetry-worker"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		CommitInterval: time.Second,
	})
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		log.Println("worker: shutting down...")
		cancel()
	}()

	log.Printf("worker: consuming from %s (group %s), pushing to %s", topic, groupID, cfg.LokiURL)

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("worker: stopped")
				return
			}
			log.Printf("worker: kafka read error: %v", err)
			continue
		}

		pushCtx, pushCancel := context.WithTimeout(ctx, 10*time.Second)
		if err := loki.PushEventJSON(pushCtx, cfg.LokiURL, msg.Value); err != nil {
			log.Printf("worker: loki push failed: %v", err)
		}
		pushCancel()
	}
}

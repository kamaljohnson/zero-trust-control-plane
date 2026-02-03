---
title: Telemetry
sidebar_label: Telemetry
---

# Telemetry

Telemetry events (gRPC request metrics and explicit `TelemetryService` emits) flow through Kafka to a worker that pushes logs to Loki for querying in Grafana.

## System architecture

### Pipeline overview

There are **two producers** of telemetry events (the gRPC unary interceptor and the TelemetryService RPCs) and **one consumer** (the telemetry worker) that writes to Loki.

```
gRPC Interceptor + TelemetryService (EmitTelemetryEvent / BatchEmitTelemetry)
         ↓
   Kafka (telemetry topic)
         ↓
   Telemetry Worker
         ↓
   Loki (logs)
         ↓
   Grafana (query via LogQL)
```

### Components

**gRPC server (producer side)**

- When `KAFKA_BROKERS` is set, the server creates a Kafka producer and passes it to both the unary interceptor and the TelemetryService handler. When unset, the producer is nil and telemetry is disabled (interceptor and Emit RPCs no-op).
- **Telemetry unary interceptor**: Runs after each RPC. Captures full method name, gRPC status code, duration, client IP (from context), and org/user/session IDs from context (set by the auth interceptor). Builds a `TelemetryEvent` with `event_type=grpc_request`, `source=grpc_interceptor`, and metadata JSON containing `full_method`, `status_code`, `duration_ms`, `client_ip`. Emits asynchronously in a goroutine with a 5s timeout; failures are logged and do not fail the RPC.
- **TelemetryService**: `EmitTelemetryEvent` and `BatchEmitTelemetry` convert request payloads to `TelemetryEvent` (with `CreatedAt` set server-side), then call the same producer. Best-effort: emit failures are logged; RPCs always return success.
- **Interceptor skip list**: No event is emitted for `HealthService/HealthCheck`; when dev OTP is enabled, `DevService/GetOTP` is also skipped (to avoid flooding Kafka).

**Kafka**

- Single topic (default `ztcp-telemetry`). The producer serializes each `TelemetryEvent` to JSON and writes it as the message value (no key). Batch timeout 50ms (segmentio/kafka-go writer).

**Telemetry worker**

- Separate binary (`cmd/worker`). Requires `KAFKA_BROKERS` and `LOKI_URL`; optional `TELEMETRY_KAFKA_TOPIC`, `KAFKA_GROUP_ID`. Uses a Kafka reader (consumer group). In a loop: read message, call Loki client to push the message value, then next message. Graceful shutdown on SIGTERM/SIGINT. Config validation expects `GRPC_ADDR` (e.g. `:0`) but the worker does not start a gRPC server.

**Loki client**

- Parses the Kafka message value as JSON. Extracts `orgId`, `eventType`, `source`, `createdAt` (camelCase from proto JSON) for stream labels and timestamp. Sanitizes label values (invalid characters replaced). Always sets `job=ztcp`. Pushes one stream per event to Loki's `/loki/api/v1/push` with the full JSON as the log line.

**Loki / Grafana**

- Loki stores log lines (raw event JSON). Queries use stream labels (`job=ztcp`, `org_id`, `event_type`, `source`) and LogQL for filtering. In Grafana, add Loki as a datasource; use LogQL to build panels and dashboards.

### Event model

- **Proto**: `TelemetryEvent` (see [telemetry.proto](../../../backend/proto/telemetry/telemetry.proto)): `org_id`, `user_id`, `device_id`, `session_id`, `event_type`, `source`, `metadata` (bytes, JSON), `created_at`.
- For interceptor-generated events: `event_type=grpc_request`, `source=grpc_interceptor`; `metadata` is JSON with `full_method`, `status_code`, `duration_ms`, `client_ip`.

### When telemetry is enabled or disabled

- **Server**: Telemetry is enabled only when `KAFKA_BROKERS` is non-empty and the Kafka producer is created successfully. If producer creation fails, a warning is logged and the server runs without telemetry.
- **Worker**: Requires `KAFKA_BROKERS` and `LOKI_URL`; exits at startup if either is missing.

## Configuration

| Variable | Used by | Description |
|----------|---------|-------------|
| `KAFKA_BROKERS` | server, worker | Comma-separated Kafka broker addresses (e.g. `localhost:9092`). Required for telemetry and for the worker. |
| `TELEMETRY_KAFKA_TOPIC` | server, worker | Kafka topic (default `ztcp-telemetry`). |
| `KAFKA_GROUP_ID` | worker | Consumer group ID (default `ztcp-telemetry-worker`). |
| `LOKI_URL` | worker | Loki base URL for push (e.g. `http://localhost:3100`). Required to run the worker. |
| `GRPC_ADDR` | server | gRPC listen address. Required by config even for the worker; set to `:0` when running only the worker. |

## Grafana

1. **Add Loki datasource**: In Grafana, add a datasource of type Loki. URL should be the same Loki instance the worker pushes to (e.g. `http://loki:3100`).

2. **Example LogQL queries**:
   - All telemetry: `{job="ztcp"}`
   - By org: `{job="ztcp"} | json | orgId="org1"`
   - gRPC requests only: `{job="ztcp", event_type="grpc_request"}`
   - Errors: `{job="ztcp"} | json | status_code != "OK"`
   - Rate by method: `sum by (full_method) (rate({job="ztcp"} | json | event_type="grpc_request" [5m]))`

## Interceptor skip list

The telemetry interceptor does not emit events for:
- `HealthService/HealthCheck`
- `DevService/GetOTP` (when dev OTP is enabled)

This avoids flooding Kafka with health checks and dev OTP calls.

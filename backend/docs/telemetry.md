# Telemetry

Telemetry events (gRPC request metrics and explicit `TelemetryService` emits) flow through Kafka to a worker that pushes logs to Loki for querying in Grafana.

## Pipeline

```
gRPC Interceptor / TelemetryService
         ↓
   Kafka (telemetry topic)
         ↓
   Telemetry Worker
         ↓
   Loki (logs)
         ↓
   Grafana (dashboards)
```

- **gRPC server**: A unary interceptor records each RPC (method, status, duration, org/user, client IP) and sends a telemetry event to Kafka. The `TelemetryService` RPCs `EmitTelemetryEvent` and `BatchEmitTelemetry` also produce to the same topic. When `KAFKA_BROKERS` is not set, telemetry is disabled (no producer, interceptor and handler no-op).
- **Kafka**: Single topic (default `ztcp-telemetry`). Events are JSON-serialized `TelemetryEvent` (org_id, user_id, event_type, source, metadata, created_at).
- **Worker** (`cmd/worker`): Consumes from the topic and pushes each message to Loki via the Loki push API. Requires `KAFKA_BROKERS`, `LOKI_URL`, and optionally `TELEMETRY_KAFKA_TOPIC`, `KAFKA_GROUP_ID`. Set `GRPC_ADDR` (e.g. `:0`) because config validation expects it.
- **Loki**: Stores log lines (the raw JSON event). Stream labels include `job=ztcp`, `org_id`, `event_type`, `source` when present.
- **Grafana**: Add Loki as a datasource pointing at the same Loki URL. Use LogQL to query and build dashboards.

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

3. **Dashboard**: An optional pre-built dashboard is in [docs/grafana/ztcp-telemetry-dashboard.json](grafana/ztcp-telemetry-dashboard.json). In Grafana: Create → Import → upload the JSON or paste its contents.

## Interceptor skip list

The telemetry interceptor does not emit events for:
- `HealthService/HealthCheck`
- `DevService/GetOTP` (when dev OTP is enabled)

This avoids flooding Kafka with health checks and dev OTP calls.

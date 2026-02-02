# Health checks

The backend exposes a **readiness** health check via the gRPC `HealthService.HealthCheck` RPC. It is used by Kubernetes, load balancers, and CI to determine whether the server is ready to accept traffic.

## Behavior

- **No database configured** (server started without `DATABASE_URL` and JWT keys): the handler has no pinger or policy checker. `HealthCheck` always returns `SERVING`.
- **Database and auth configured**: the handler runs two checks on each call: (1) ping the database via `PingContext`; (2) verify the in-process OPA policy engine by compiling the default Rego policy and running one trivial query (no database or policy-repo call inside the OPA check). It returns `SERVING` only when both checks succeed; otherwise it returns `NOT_SERVING`. On any failure the RPC still succeeds (no gRPC error), so probes receive a successful response with status `NOT_SERVING`.

The same RPC can be used for both liveness and readiness in Kubernetes, or for readiness only (with a separate liveness probe if desired).

## Calling the health RPC

The service is `ztcp.health.v1.HealthService`, method `HealthCheck`. The request is empty; the response contains `status` (`SERVING` or `NOT_SERVING`).

**Kubernetes**: Use a gRPC probe or a tool such as [grpc_health_probe](https://github.com/grpc-ecosystem/grpc-health-probe). The backend uses a custom health proto (not the standard `grpc.health.v1` package), so the probe must call `ztcp.health.v1.HealthService/HealthCheck`. Example with grpc_health_probe and a custom service:

```bash
grpc_health_probe -addr=:8080 -service=ztcp.health.v1.HealthService
```

If your tool only supports the standard `grpc.health.v1.Health` service name, you may need to call the RPC explicitly (e.g. via a small client or a probe that supports custom service/method).

**From Go**: Use the generated client `healthv1.NewHealthServiceClient(cc)` and call `HealthCheck(ctx, &healthv1.HealthCheckRequest{})`.

## Configuration

No extra configuration is required. When auth (and thus the database) is enabled, the server wires the database as the health pinger and the OPA policy evaluator as the policy checker automatically. When auth is disabled, health always returns `SERVING`.

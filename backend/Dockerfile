# Multi-stage build for ZTCP backend
# Builder stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGO disabled for static linking
# -p 1 builds one package at a time to reduce peak memory on small droplets (e.g. 4GB)
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -p 1 -ldflags="-w -s" -o ztcp-server ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS and timezone data
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 ztcp && \
    adduser -D -u 1000 -G ztcp ztcp

# Copy binary from builder
COPY --from=builder /build/ztcp-server /app/ztcp-server

# Copy migrations
COPY --from=builder /build/internal/db/migrations /app/migrations

# Install wget (frontend health) and netcat for TCP health check (gRPC port; must run as root; do before USER)
RUN apk --no-cache add wget netcat-openbsd

# Set ownership
RUN chown -R ztcp:ztcp /app

# Switch to non-root user
USER ztcp

# Expose gRPC port
EXPOSE 8080

# Health check: backend exposes gRPC only on 8080; use TCP connect
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD nc -z localhost 8080 || exit 1

# Run the server
CMD ["/app/ztcp-server"]

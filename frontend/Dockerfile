# Multi-stage build for ZTCP frontend
# Builder stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Ensure public exists (Next.js may not create it; COPY in runner stage requires it)
RUN mkdir -p /app/public

# Set production environment (NEXT_PUBLIC_* baked at build time; pass via build-arg for dev OTP etc.)
ARG NEXT_PUBLIC_DEV_OTP_ENABLED=false
ARG NEXT_PUBLIC_DOCS_URL
ARG NEXT_PUBLIC_GRAFANA_URL
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_PUBLIC_DEV_OTP_ENABLED=${NEXT_PUBLIC_DEV_OTP_ENABLED}
ENV NEXT_PUBLIC_DOCS_URL=${NEXT_PUBLIC_DOCS_URL}
ENV NEXT_PUBLIC_GRAFANA_URL=${NEXT_PUBLIC_GRAFANA_URL}

# Build the Next.js application
RUN npm run build

# Production stage
FROM node:20-alpine AS runner

WORKDIR /app

# Install dumb-init for proper signal handling
RUN apk add --no-cache dumb-init

# Create non-root user (avoid -g/-u 1000; node image may already use it)
RUN addgroup nodejs && \
    adduser -D -G nodejs nodejs

# Set production environment
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Copy necessary files from builder
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/package-lock.json* ./package-lock.json

# Install production dependencies (Next.js standalone includes minimal node_modules;
# install any additional production dependencies that may be needed at runtime)
RUN npm ci --omit=dev && npm cache clean --force

# Install wget for health checks
RUN apk --no-cache add wget

# Set ownership
RUN chown -R nodejs:nodejs /app

# Switch to non-root user
USER nodejs

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --spider --quiet http://localhost:3000 || exit 1

# Use dumb-init to handle signals properly
ENTRYPOINT ["dumb-init", "--"]

# Start the Next.js server
CMD ["node", "server.js"]

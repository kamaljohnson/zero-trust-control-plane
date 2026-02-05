# Multi-stage build for ZTCP docs (Docusaurus)
# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# baseUrl /docs/ so the site is served under /docs/ behind nginx
ENV BASE_URL=/docs/
ENV DOCS_URL=https://ztcp.projects.xborglabs.com

COPY package.json package-lock.json* ./
RUN npm ci

COPY . .
RUN npm run build

# Run stage: serve static build
FROM node:20-alpine AS runner

WORKDIR /app

RUN apk add --no-cache wget && \
    addgroup -g 1001 nodejs && \
    adduser -D -u 1001 -G nodejs nodejs

COPY package.json package-lock.json* ./
RUN npm ci --omit=dev && npm cache clean --force

COPY --from=builder --chown=nodejs:nodejs /app/build ./build

USER nodejs
EXPOSE 3001

# Docusaurus with BASE_URL=/docs/ builds into build/docs/ directory
# We serve from build/docs/ so that / paths map correctly
# serve -s enables SPA mode (fallback to index.html), -l sets port
CMD ["npx", "serve", "-s", "build/docs", "-l", "3001"]

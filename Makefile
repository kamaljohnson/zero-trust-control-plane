# Local deployment for ZTCP.
# Run `make setup` from repo root to start Docker, configure env, migrate, and seed in one go.
# Then start backend and frontend in separate terminals: `make run-backend`, `make run-frontend`.
# See deploy/README.md for details.

.PHONY: setup up down env ensure-env wait-postgres migrate seed run-backend run-frontend run-docs install-frontend install-docs

BACKEND_DIR  := backend
DEPLOY_DIR   := deploy
DOCS_DIR     := docs-site
FRONTEND_DIR := frontend

# Default target: full local setup in one go (env, Docker up, wait for Postgres, migrate, optional seed).
setup: ensure-env up wait-postgres migrate
	@[ "$(SKIP_SEED)" = "1" ] || $(MAKE) seed
	@echo ""
	@echo "--- Local setup complete ---"
	@echo "Start the backend in one terminal:  make run-backend"
	@echo "Start the frontend in another:      make run-frontend"
	@echo "Then open http://localhost:3000"
	@echo "---"

# Copy deploy/.env.example to backend/.env and frontend/.env if missing (no overwrite).
env:
	@if [ ! -f $(BACKEND_DIR)/.env ]; then cp $(DEPLOY_DIR)/.env.example $(BACKEND_DIR)/.env; echo "Created $(BACKEND_DIR)/.env from deploy/.env.example"; fi
	@if [ ! -f $(FRONTEND_DIR)/.env ]; then cp $(DEPLOY_DIR)/.env.example $(FRONTEND_DIR)/.env; echo "Created $(FRONTEND_DIR)/.env from deploy/.env.example"; fi

# Ensure backend/.env and frontend/.env exist (from deploy/.env.example). No mutation of existing files.
# If backend/.env exists but DATABASE_URL uses user "root", replace it with local deploy default (ztcp/ztcp).
ensure-env:
	@if [ ! -f $(BACKEND_DIR)/.env ]; then cp $(DEPLOY_DIR)/.env.example $(BACKEND_DIR)/.env; echo "Created $(BACKEND_DIR)/.env from deploy/.env.example"; fi
	@if [ -f $(BACKEND_DIR)/.env ] && grep -q '^DATABASE_URL=.*//root@' $(BACKEND_DIR)/.env 2>/dev/null; then \
		awk 'BEGIN { d="DATABASE_URL=postgres://ztcp:ztcp@localhost:5432/ztcp?sslmode=disable" } /^DATABASE_URL=/ { print d; next } { print }' $(BACKEND_DIR)/.env > $(BACKEND_DIR)/.env.tmp && mv $(BACKEND_DIR)/.env.tmp $(BACKEND_DIR)/.env; \
		echo "Updated DATABASE_URL in $(BACKEND_DIR)/.env to local deploy default (ztcp/ztcp)."; \
	fi
	@if [ ! -f $(FRONTEND_DIR)/.env ]; then cp $(DEPLOY_DIR)/.env.example $(FRONTEND_DIR)/.env; echo "Created $(FRONTEND_DIR)/.env from deploy/.env.example"; fi

# Start PostgreSQL (Docker Compose).
up:
	cd $(DEPLOY_DIR) && docker compose up -d
	@echo "Waiting for Postgres..."
	@$(MAKE) wait-postgres

# Stop Docker stack.
down:
	cd $(DEPLOY_DIR) && docker compose down

# Wait for Postgres to be ready (used after up).
wait-postgres:
	@max=30; i=0; until cd $(DEPLOY_DIR) && docker compose exec -T postgres pg_isready -U ztcp 2>/dev/null; do \
		i=$$((i+1)); [ $$i -ge $$max ] && { echo "Postgres not ready in time" >&2; exit 1; }; \
		sleep 2; \
	done
	@echo "Postgres is ready"

# Run database migrations. Requires backend/.env with DATABASE_URL.
migrate: ensure-env
	cd $(BACKEND_DIR) && ./scripts/migrate.sh up

# Seed development data (e.g. dev@example.com). Use SKIP_SEED=1 with make setup to skip.
seed: ensure-env
	cd $(BACKEND_DIR) && ./scripts/seed.sh

# Run backend gRPC server (foreground). Use in a dedicated terminal after setup.
run-backend:
	cd $(BACKEND_DIR) && go run ./cmd/server

# Install frontend deps and run Next.js dev server (foreground). Use in a dedicated terminal after setup.
run-frontend: install-frontend
	cd $(FRONTEND_DIR) && npm run dev

# Install frontend dependencies (npm install).
install-frontend:
	cd $(FRONTEND_DIR) && npm install

# Install docs-site dependencies (npm install).
install-docs:
	cd $(DOCS_DIR) && npm install

# Run docs-site dev server (foreground). Use in a dedicated terminal. Serves at http://localhost:3001.
run-docs: install-docs
	cd $(DOCS_DIR) && npm run start -- --port 3001

# GitHub Actions Workflows

This directory contains CI/CD workflows for the Zero Trust Control Plane (ZTCP).

## Workflows

### Backend Tests (`backend-tests.yml`)

- **Trigger**: Push or pull request to `main` when `backend/**` or this workflow file changes; also manual.
- **Purpose**: Run backend tests with race detector and coverage; upload coverage to Codecov.
- **Secrets**: `CODECOV_TOKEN` (optional, for coverage upload).

### Deploy to Digital Ocean (`deploy.yml`)

- **Trigger**: Push to `main`; also manual via **Actions → Deploy to Digital Ocean → Run workflow**.
- **Purpose**: Test backend and frontend, build Docker images, push to a container registry, then SSH to the droplet to pull images and update services.
- **Jobs**:
  1. **test-backend** – `go test` in `backend/`
  2. **test-frontend** – `npm ci`, `npm run lint`, `npm run build` in `frontend/`
  3. **build-and-push** – Build backend, frontend, and nginx images; tag with `latest` and commit SHA; push to registry
  4. **deploy** – SSH to droplet, `git pull`, `docker compose pull`, run migrations, `docker compose up -d`, verify

## Required secrets (Deploy workflow)

| Secret | Required | Description |
|--------|----------|-------------|
| `DROPLET_HOST` | Yes | Droplet IP or hostname |
| `DROPLET_USER` | Yes | SSH user (e.g. `root`, `deploy`) |
| `DROPLET_SSH_KEY` | Yes | Private SSH key (full PEM) |
| `DROPLET_PORT` | No | SSH port (default 22) |
| `DOCKER_USERNAME` | No* | Registry username (Docker Hub) |
| `DOCKER_PASSWORD` | No* | Registry password/token (Docker Hub) |

\* If `DOCKER_USERNAME` is not set, the workflow uses **GitHub Container Registry** (ghcr.io) with `GITHUB_TOKEN`.

## Optional variables (Deploy workflow)

| Variable | Description |
|----------|-------------|
| `DEPLOY_REPO_DIR` | Path to the repo on the droplet (default: `/opt/zero-trust-control-plane`) |

## Deployment process summary

1. Push to `main` (or run the workflow manually).
2. Backend and frontend tests must pass.
3. Images are built and pushed to the registry as:
   - `$REGISTRY/ztcp-backend:latest` and `$REGISTRY/ztcp-backend:$SHA`
   - `$REGISTRY/ztcp-frontend:latest` and `$REGISTRY/ztcp-frontend:$SHA`
   - `$REGISTRY/ztcp-nginx:latest` and `$REGISTRY/ztcp-nginx:$SHA`
4. The deploy job SSHs to the droplet and:
   - Updates the repo (`git pull`),
   - Sets `IMAGE_TAG` and `DOCKER_REGISTRY`,
   - Runs `docker compose pull`,
   - Runs migrations (host-side, with `DATABASE_URL` using `localhost:5432`),
   - Runs `docker compose up -d`,
   - Prunes unused images,
   - Verifies with `docker compose ps`.

Full setup (secrets, droplet prep, registry login, troubleshooting) is in [deploy/PRODUCTION.md](../../deploy/PRODUCTION.md#cicd-github-actions).

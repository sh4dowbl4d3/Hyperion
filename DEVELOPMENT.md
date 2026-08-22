# Development Guide

## Prerequisites

- Go >= 1.26
- Node.js >= 22 (with npm)
- Docker + Docker Compose
- `psql` client (optional, for inspecting the database)

## Full Stack (Docker)

```bash
cp .env.example .env
docker compose up --build
```

Services:

| Service | Port | Notes |
| --- | --- | --- |
| frontend | 5173 | Vite dev server (API proxied to backend) |
| backend | 8080 | REST API under `/api/v1` |
| postgres | internal only | reachable on the compose network as `postgres` |

Reset everything including data:

```bash
docker compose down -v
```

## Backend Only

```bash
cd backend

# start just the database
docker compose up -d postgres

export $(grep -v '^#' ../.env.example | xargs)   # or set vars from your own .env
go run ./cmd/server
```

The server loads configuration from environment variables (see `.env.example`)
and exits with a clear error when required values are missing.

## Frontend Only

```bash
cd frontend
npm install
npm run dev      # expects the backend on http://localhost:8080
```

## Testing

Backend:

```bash
cd backend
go test ./...                 # unit tests
go build ./...                # compile check
go vet ./...                  # static analysis
```

Integration tests that need PostgreSQL are guarded by a build tag and run when
the database is available:

```bash
docker compose up -d postgres
TEST_DATABASE_DSN="postgres://moderndvwa:devpassword@localhost:5432/moderndvwa?sslmode=disable" \
  go test -tags=integration ./tests/...
```

Frontend:

```bash
cd frontend
npm run build     # type-checks via tsc and produces dist/
npm test          # vitest unit tests
```

## Migrations

Migrations are plain SQL files in `backend/migrations/`, named
`NNNN_description.up.sql` / `NNNN_description.down.sql`. The runner applies all
unapplied migrations in order inside transactions at startup and records them
in `schema_migrations`. Down migrations exist so rollback behavior can be
tested.

## Git Workflow

- Conventional Commits (`feat:`, `fix:`, `test:`, `docs:`, `chore:`)
- Atomic commits; one logical change per commit
- Never rewrite published history; never commit `.env` or secrets

## Debugging Tips

- Backend logs are structured (key/value). Set `LOG_LEVEL=debug` for verbose output.
- `docker compose logs -f backend` follows API logs in Docker mode.
- If migrations fail on startup, check `SELECT * FROM schema_migrations;` —
  partially applied batches roll back transactionally, but a failed *down*
  migration can leave state behind.

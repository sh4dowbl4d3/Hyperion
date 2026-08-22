# Architecture

## System Overview

```
+---------------------+        +------------------------+
|  React + TS + Vite  |  HTTP  |   Go API (Gin)         |
|  frontend :5173     | -----> |   backend :8080        |
|  (browser)          |        |   /api/v1/*            |
+---------------------+        +-----------+------------+
                                           |
                              +------------+------------+
                              |                         |
                    +---------v---------+    +----------v----------+
                    | PostgreSQL :5432  |    | lab services        |
                    | users, labs,      |    | (in-process,        |
                    | progress          |    | isolated modules)   |
                    +-------------------+    +---------------------+
```

All traffic flows through the versioned REST API (`/api/v1`). The frontend is a
static SPA; it never talks to the database directly.

## Technology Choices

| Concern | Choice | Rationale |
| --- | --- | --- |
| Backend language | Go 1.26 | Single static binary, strong stdlib, fast tests |
| HTTP framework | Gin | De-facto standard, middleware ecosystem, good test story |
| Auth | JWT (HS256, short-lived) | Stateless auth matching modern SPA patterns |
| Database | PostgreSQL | Relational fit for users/labs/progress; transactions needed for race-condition lab |
| Migrations | SQL files + custom runner | Explicit, reviewable schema changes; no ORM magic |
| Frontend | React 19 + TypeScript + Vite | Fast DX, typed API layer, protected routing |
| Infrastructure | Docker Compose | Reproducible local environment with network isolation |

Dependencies are kept deliberately minimal: only libraries whose absence would
force us to build and maintain risky primitives ourselves (JWT signing,
password hashing, routing).

## Backend Layout

```
backend/
├── cmd/server/          # entrypoint: config load, server start, graceful shutdown
├── internal/
│   ├── api/             # router construction, route registration
│   ├── auth/            # password hashing, JWT service, registration/login services
│   ├── config/          # env-based configuration
│   ├── database/        # pgx pool + migration runner
│   ├── httpx/           # HTTP server wrapper, centralized error envelope
│   ├── labs/            # lab engine (see below)
│   ├── logging/         # structured logger
│   └── middleware/      # request id, logging, recovery, authn, authz
└── migrations/          # numbered .sql files
```

## Lab Engine

Every vulnerability is implemented as a **lab** registered in an in-memory
registry at startup. A lab is a self-contained module that provides:

```go
type Lab interface {
    Meta() Meta                       // id, slug, name, category, difficulty, xp, hints...
    RegisterRoutes(rg *gin.RouterGroup) // mounts vulnerable + secure endpoints
}
```

Design rules:

1. **Isolation** — lab endpoints live under `/api/v1/labs/<slug>/...`. Labs use
   their own synthetic data stores (DB tables prefixed per-lab or in-memory
   fixtures). They never touch platform tables beyond progress tracking.
2. **Vulnerable + secure pair** — each lab exposes both behaviors side by side
   (e.g. `GET /labs/sqli/users/search` vs `GET /labs/sqli/users/search-safe`)
   so the learner can diff the implementations.
3. **Completion detection** — labs report completion events to the progress
   service; the engine validates the event against the lab's completion
   condition before awarding XP. XP is awarded once per user per lab.
4. **No global weakening** — the platform's own authentication/authorization is
   production-grade. The JWT lab, for example, implements a *separate* token
   flow scoped entirely inside the lab.

## Data Model

- `roles` — seeded roles (`user`, `admin`)
- `users` — credentials (bcrypt), role reference
- `labs` — lab catalog mirrored from the registry (seeded on boot)
- `progress` — per-user lab state: status, xp awarded, timestamps

Lab-specific synthetic data lives in separate tables created by lab-owned
migrations (e.g. `sqli_users` for the SQL injection lab).

## Security Boundaries

Three concentric boundaries:

1. **Host boundary** — Docker Compose runs everything inside a dedicated
   bridge network. Only `frontend` (5173) and `backend` (8080) publish ports;
   Postgres is reachable only on the internal network.
2. **Platform boundary** — real authentication (bcrypt + signed JWT with
   expiry validation), role-based authorization middleware, centralized error
   handling. This code must never contain intentional weaknesses.
3. **Lab boundary** — intentional vulnerabilities exist *only* inside lab
   modules, operating exclusively on synthetic data. Labs cannot escalate to
   the host, the database superuser context, or other users' real data.

Details in [SECURITY.md](SECURITY.md).

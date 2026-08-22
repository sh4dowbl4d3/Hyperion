# ModernDVWA

ModernDVWA is a modern, intentionally vulnerable cybersecurity training platform
inspired by [DVWA](https://github.com/digininja/DVWA), redesigned around the way
real applications are built today: REST APIs, token-based authentication,
client-side rendering, and containerized infrastructure.

It is designed to be run **locally only**, against **synthetic data**, inside an
isolated Docker environment.

## Highlights

- Go (Gin) backend with a versioned REST API and real JWT authentication
- React + TypeScript + Vite frontend with protected routing
- PostgreSQL persistence with SQL migrations
- Reusable lab engine: every lab ships vulnerable behavior *and* a secure
  reference implementation, plus completion detection, hints and XP tracking
- Fully containerized with Docker Compose; no host filesystem access required

## MVP Labs

| Lab | Category | Difficulty |
| --- | --- | --- |
| SQL Injection | Injection | Medium |
| Stored XSS | Injection | Easy |
| IDOR / BOLA | Broken Access Control | Medium |
| JWT Vulnerability | Authentication | Hard |
| SSRF | Server-Side Request Forgery | Hard |
| Race Condition / Business Logic | Business Logic | Hard |

## Quick Start

```bash
cp .env.example .env          # generate local dev secrets
docker compose up --build     # start postgres, backend, frontend
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080/api/v1/healthz

Register an account in the UI, then pick a lab from the dashboard.

For local (non-Docker) development see [DEVELOPMENT.md](DEVELOPMENT.md).

## Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) — system design and lab engine
- [DEVELOPMENT.md](DEVELOPMENT.md) — development workflow and testing
- [SECURITY.md](SECURITY.md) — security boundaries and responsible use
- `docs/labs/` — per-lab guides and solutions
- `docs/architecture/` — component deep-dives

## Legal & Intended Use

ModernDVWA is an educational platform. All vulnerabilities are intentional,
isolated, and exist only inside lab endpoints backed by synthetic data.
Never expose this application to untrusted networks. See
[SECURITY.md](SECURITY.md).

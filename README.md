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
- In-app lab playground: drive every vulnerable endpoint (and its safe twin)
  from the lab detail page — no curl required
- Fully containerized with Docker Compose; no host filesystem access required

## MVP Labs

| Lab | Category | Difficulty | XP |
| --- | --- | --- | --- |
| SQL Injection | Injection | Easy | 100 |
| Stored XSS | Injection | Easy | 100 |
| IDOR / BOLA | Authorization | Easy | 100 |
| JWT Vulnerability | Authentication | Medium | 150 |
| SSRF | Request Forgery | Medium | 150 |
| Race Condition | Business Logic | Hard | 200 |

Per-lab guides with objectives, walkthroughs and mitigations live in
[`docs/labs/`](docs/labs/) ([sqli](docs/labs/sqli.md),
[xss](docs/labs/xss.md), [idor](docs/labs/idor.md), [jwt](docs/labs/jwt.md),
[ssrf](docs/labs/ssrf.md), [race-condition](docs/labs/race-condition.md)).

## Quick Start

```bash
cp .env.example .env          # generate local dev secrets
docker compose up --build     # start postgres, backend, frontend
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080/api/v1/healthz

Register an account in the UI, then pick a lab from the dashboard. Each lab
page includes a **Playground** panel for interacting with its endpoints; the
Race Condition lab adds one-click burst firing to win the race without any
scripting.

For local (non-Docker) development see [DEVELOPMENT.md](DEVELOPMENT.md).

## Documentation

- [LICENSE](LICENSE) — GNU General Public License v3
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

# Hyperion

> Modern, intentionally vulnerable web application security training platform.

![Hyperion Interface](screenshot/screenshot1.png)

Hyperion is an intentionally vulnerable web application security training platform inspired by classic tools like [DVWA](https://github.com/digininja/DVWA), rebuilt from the ground up for modern software architectures. Instead of legacy monolithic server-side rendering, Hyperion reflects how modern production systems are built: decoupled single-page applications, REST APIs, token-based authentication, and containerized microservices.

Hyperion is designed to be run **locally only**, against **synthetic data**, inside an isolated Docker environment.

---

## Core Features

- **Realistic Modern Stack**: Built with a Go (Gin) REST API backend, a React + TypeScript single-page application frontend, and a PostgreSQL database.
- **Twin-Endpoint Lab Design**: Every lab module implements both an exploitable endpoint and its secure reference implementation (`-safe` twin), enabling side-by-side comparison of attack vectors and effective remediations.
- **In-Browser Interactive Playground**: Test payloads, modify HTTP parameters, and inspect structured server responses directly within the application—no external proxies or curl commands required.
- **Automated Validation & Progress Tracking**: Real-time detection of successful exploits, flag capture validation, progressive hint unlocking, and XP accounting.
- **Isolated Sandbox**: Runs entirely on local synthetic data in an isolated container network with zero host filesystem access.
- **Frictionless Initialization**: Automatically initializes an operator session upon launch, allowing immediate exploration of the lab catalog without manual database seeding or registration hurdles.

---

## Available Labs

Hyperion includes 8 core vulnerability modules covering critical OWASP categories:

| Lab | Category | Difficulty | XP | Documentation |
| :--- | :--- | :--- | :--- | :--- |
| **SQL Injection** | Injection | Easy | 100 | [Guide](docs/labs/sqli.md) |
| **Stored XSS** | Injection | Easy | 100 | [Guide](docs/labs/xss.md) |
| **IDOR / BOLA** | Authorization | Easy | 100 | [Guide](docs/labs/idor.md) |
| **JWT Vulnerabilities** | Authentication | Medium | 150 | [Guide](docs/labs/jwt.md) |
| **SSRF** | Request Forgery | Medium | 150 | [Guide](docs/labs/ssrf.md) |
| **Race Condition** | Business Logic | Hard | 200 | [Guide](docs/labs/race-condition.md) |
| **Open Redirect** | Request Forgery | Easy | 100 | [Guide](docs/labs/open-redirect.md) |
| **Command Injection** | Injection | Medium | 150 | [Guide](docs/labs/command-injection.md) |

Detailed documentation for each vulnerability, including exploitation objectives, walkthroughs, and code-level mitigations, is available in [`docs/labs/`](docs/labs/).

---

## System Architecture

| Component | Technology | Role |
| :--- | :--- | :--- |
| **Frontend** | React 19, TypeScript, Vite | Single-page application, lab playground, telemetry dashboard |
| **Backend API** | Go 1.26, Gin | RESTful API (`/api/v1`), JWT authentication, lab target routing |
| **Database** | PostgreSQL 17 | User accounts, lab metadata, progress persistence |
| **Runtime** | Docker Compose | Network-isolated multi-container environment |

All client interaction routes through the versioned API (`/api/v1`). Lab targets execute as isolated in-process modules that record user progress upon successful exploitation.

---

## Installation & Getting Started

### Prerequisites

Ensure the following dependencies are installed on your host system:

- [Docker](https://docs.docker.com/get-docker/) (v24.0+)
- [Docker Compose](https://docs.docker.com/compose/) (v2.20+)
- [Git](https://git-scm.com/)

### 1. Clone the Repository

```bash
git clone https://github.com/sh4dowbl4d3/Hyperion.git
cd Hyperion
```

### 2. Configure Environment

Copy the example environment configuration file:

```bash
cp .env.example .env
```

The default values in `.env.example` are pre-configured for local containerized execution.

### 3. Build and Start Services

Launch the environment using Docker Compose:

```bash
docker compose up --build -d
```

This command will:
1. Initialize the PostgreSQL database and execute schema migrations.
2. Build and start the Go API server on port `8090` (mapped internally to `8080`).
3. Build and serve the React frontend on port `5173`.

### 4. Verify Installation

Check service health status:

```bash
docker compose ps
```

Verify backend API readiness:

```bash
curl http://localhost:5173/api/v1/healthz
```

Expected response:
```json
{"service":"hyperion-api","status":"ok","version":"0.1.0"}
```

Access the web interface at **`http://localhost:5173`**.

---

## Authentication & Credentials

Hyperion automatically creates and connects an active operator session when accessing the web interface. 

For direct API access or manual authentication testing, a default demo account is seeded automatically:

| Field | Value |
| :--- | :--- |
| **Email** | `demo@hyperion.test` |
| **Password** | `Password123!` |

API endpoints require authentication via Bearer token:

```bash
# Obtain token
TOKEN=$(curl -s -X POST http://localhost:5173/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@hyperion.test","password":"Password123!"}' | jq -r .token)

# Access protected catalog
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:5173/api/v1/labs
```

---

## Stopping the Services

To stop running containers:

```bash
docker compose down
```

To stop containers and reset all database progress and volumes:

```bash
docker compose down -v
```

---

## Local Development & Testing

For contributors wishing to run the backend and frontend outside Docker:

- **Backend Development**: Requires Go >= 1.26. See [DEVELOPMENT.md](DEVELOPMENT.md#backend-only).
- **Frontend Development**: Requires Node.js >= 22. See [DEVELOPMENT.md](DEVELOPMENT.md#frontend-only).

### Running Tests

**Backend unit tests:**
```bash
cd backend
go test ./...
```

**Frontend test suite:**
```bash
cd frontend
npm test
npm run typecheck
```

---

## Documentation Index

- [ARCHITECTURE.md](ARCHITECTURE.md) — System architecture, lab registry, and data model
- [DEVELOPMENT.md](DEVELOPMENT.md) — Local development guide, environment variables, and test instructions
- [SECURITY.md](SECURITY.md) — Security policy, threat model, and isolation boundaries
- [docs/labs/](docs/labs/) — Lab-specific walkthroughs, solutions, and secure remediation guides
- [docs/architecture/](docs/architecture/) — Component-level architectural documentation

---

## Legal & Responsible Use

Hyperion is designed strictly for educational and defensive cybersecurity training. All vulnerabilities are intentional and operate exclusively within self-contained lab endpoints using synthetic data.

**Do not deploy Hyperion to production or expose it to public, untrusted networks.** See [SECURITY.md](SECURITY.md) for further guidance.

---

## License

This project is licensed under the [GNU General Public License v3.0](LICENSE).

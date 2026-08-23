# Security

## Purpose and Intended Use

Hyperion is a **deliberately vulnerable training platform** for learning
about web application security in a safe, local environment.

- Run it **only on your own machine or an isolated lab network**.
- Never expose it to the public internet, shared networks, or untrusted users.
- Never point lab tooling at systems you do not own.

## Security Model: Platform vs Labs

The application has two strictly separated zones:

### 1. The platform (must stay secure)

Registration, login, JWT issuance/validation, role authorization, progress
tracking, and all infrastructure code. These components use real security
controls:

- bcrypt password hashing with per-user salts
- HS256 JWTs with `exp` validation enforced on every protected route
- Role-based authorization middleware
- Parameterized SQL for all platform queries
- Centralized error handling that never leaks internals (stack traces, SQL)

**Rule:** no intentional weakness may ever be introduced into platform code.
Vulnerability demonstrations belong exclusively in labs.

### 2. The labs (intentionally vulnerable, isolated)

Each vulnerability lives inside its own lab module under
`/api/v1/labs/<slug>/`:

- Operates only on **synthetic seed data** created for that lab.
- Cannot modify platform tables except through the progress service API.
- Has no host access: no filesystem paths outside container-scoped storage, no
  shell execution, no outbound requests beyond the internal compose network
  (the SSRF lab's target is a dedicated synthetic service).
- Ships a secure reference implementation so learners see the correct control
  side-by-side with the broken one.

## Specific Isolation Decisions

| Risk | Control |
| --- | --- |
| SSRF reaching cloud metadata / LAN | Lab fetches are restricted to the internal `labs-net` compose network; there is no metadata service to reach. Secure mode validates scheme + allowlisted hosts. |
| JWT lab weakening real auth | The vulnerable token flow is a separate service with its own endpoints and its own signing material; platform auth is untouched. |
| XSS persisting across sessions | Payloads render only inside the lab page using React rendering modes chosen by the lab; no cookies are exposed to JS (token kept out of `document.cookie`). |
| SQL injection reaching platform data | Vulnerable query runs against a dedicated `sqli_*` table containing fake customer data only. |
| Container escape | No privileged containers, no Docker socket mounts, no host path mounts; containers run as non-root users where practical. |

## Reporting Issues

If you find an *unintentional* vulnerability — one that breaks the isolation
rules above — please open an issue describing reproduction steps. Unintentional
vulnerabilities are treated as high-priority bugs.

## Security Review Checklist (MVP release)

Reviewed at the end of the 3-day sprint:

- [x] No secrets in git: `.env` gitignored and untracked; only `.env.example`
      with placeholders is committed.
- [x] Platform JWT secret requires ≥32 bytes at startup (`config.Load`).
- [x] Weak lab secret (`"secret"`) exists only inside `internal/labs/jwtlab`;
      no platform code references it.
- [x] JWT lab flow fully isolated: own subject namespace, own secrets, own
      `X-Lab-Badge` header; platform middleware untouched (tampered platform
      token verified to still return 401).
- [x] All lab target routes require authentication
      (`middleware.Authenticate` on the `/targets` group).
- [x] SSRF internal service binds to `127.0.0.1:<ephemeral>` only; serves
      hard-coded synthetic data; no cloud metadata or LAN reachability.
- [x] Docker: no privileged containers, no socket or host-path mounts,
      backend runs as non-root `app` user; both published ports bound to
      `127.0.0.1`.
- [x] No file-upload, file-read or path-traversal endpoints exist anywhere in
      the backend.
- [x] Lab SQL errors never echo driver messages (regression-tested).
- [x] Unknown vs forbidden object access collapsed to 404 (no existence
      oracle) in the IDOR lab.
- [x] `npm audit --omit=dev`: 0 vulnerabilities; govulncheck recommended as a
      follow-up CI step.

## Data Handling

All data is synthetic. Do not enter real credentials, personal information, or
production secrets into this application.

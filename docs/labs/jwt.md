# Lab: JWT Vulnerability — Guest Badge Forge

- **Slug:** `jwt`
- **Category:** authentication
- **Difficulty:** medium
- **XP:** 150

## Objective

Take the guest token from `/api/v1/targets/jwt/guest-token`, craft an
admin-privileged variant that `/targets/jwt/admin-panel` accepts (unsigned
`alg:none`, or self-signed with the weak secret), and retrieve the flag. The
hardened verifier behind `/admin-panel-safe` rejects every forgery.

## Learning goal

Understand why a JWT's header must never be trusted to describe its own
security, and how algorithm confusion plus guessable secrets let attackers
mint arbitrary claims.

## Scenario

A legacy badge printer issues JWTs signed with the well-known weak secret
`secret` and honours whatever `alg` the token declares. The admin panel only
checks the decoded `role` claim — it never questions where the token came from.

**Isolation guarantee:** this lab's flow is entirely separate from platform
authentication. Lab tokens use their own subject namespace (`badge-<id>`),
their own secrets, and are sent via a dedicated `X-Lab-Badge` header. The real
platform JWT service, middleware and secret are untouched; a tampered platform
token still fails with 401 (covered by E2E verification).

## Walkthrough (spoilers)

Two independent forgeries work against the vulnerable panel:

1. **alg:none.** Decode the guest token's header, keep it but set
   `{"alg":"none"}`, replace the payload's role with `"admin"`, and use any
   (or an empty) signature segment. The decoder sees `alg: none` and skips
   signature verification entirely.
2. **Weak secret.** Re-sign `{"role":"admin",...}` yourself with
   HMAC-SHA256 and the key `secret`. The naive verifier compares MACs using
   that same key, so your forgery verifies perfectly.

Both forgeries return 401 from `/admin-panel-safe`: the secure verifier pins
the algorithm server-side (`HS256` or nothing), uses a separate strong secret,
and enforces issuer and expiry claims.

## Impact

Complete authentication bypass: privilege escalation, impersonation of any
user, and durable backdoors via long-lived forged tokens.

## Secure implementation

`internal/labs/jwtlab/token.go` holds both verifiers side by side:

- `DecodeVulnerable` — trusts header-declared algorithms, weak secret,
  ignores expiry/issuer.
- `SecureVerifier.Verify` — pinned algorithm, strong secret, issuer and
  expiry enforced.

## Mitigations

1. Pin the accepted algorithm in code; reject everything else before parsing.
2. Reject `alg:none` outright unless a documented, authenticated use exists.
3. Use high-entropy secrets (the platform enforces ≥32 bytes at config load).
4. Validate every security claim (`iss`, `exp`, `aud`) after signature checks.
5. Keep test/legacy signing keys out of production configuration.

## Testing

Unit tests in `internal/labs/jwtlab/lab_test.go` forge both variants and prove
acceptance by the vulnerable decoder, rejection by the secure verifier, expiry
and issuer enforcement, guest issuance staying low-privilege, completion
awarded exactly once per genuine forgery, and no award from the safe twin.

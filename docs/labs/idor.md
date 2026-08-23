# Lab: IDOR/BOLA — Shared Drive Documents

- **Slug:** `idor`
- **Category:** authorization
- **Difficulty:** easy
- **XP:** 100

## Objective

Enumerate document IDs on `/api/v1/targets/idor/documents/:id` and read a
document you do not own (classification `confidential`). Then confirm
`/targets/idor/documents-safe/:id` refuses cross-tenant reads.

## Learning goal

Understand broken object-level authorization (OWASP API #1): authentication
proves *who* you are, but each object access must independently verify *whose*
the object is. Sequential IDs make enumeration trivial.

## Scenario

A shared-drive document service resolves files by numeric ID. Every registered
user is seeded with their own demo documents; a synthetic service account
(`idor-victim@example.test`, unusable password hash) owns a confidential
quarterly review containing the flag. The store's `GetByID` intentionally has
**no** ownership filter — checking ownership is the handler's job, which is
exactly what the vulnerable variant fails to do.

## Walkthrough (spoilers)

1. GET `/documents` — you receive only your own files, with small sequential IDs.
2. Request IDs just outside your range, e.g. `/documents/3`.
3. The vulnerable handler fetches whatever ID you name and returns it:
   another user's confidential document, including the flag.
4. The same request to `/documents-safe/3` returns 404 — it compares
   `owner_id` against your authenticated identity before responding.
5. Your own documents still work through the safe twin (200).

Note both twins answer "not found" identically for unknown IDs, avoiding an
existence oracle that would help attackers map other tenants' data.

## Impact

Cross-tenant data exposure, privacy violations, regulatory breach; in real
APIs this is the most commonly found class of critical API vulnerability.

## Secure implementation

`internal/labs/idor/lab.go`:

```go
if enforceOwnership && doc.OwnerID != userID {
    httpx.NotFound(c, "no such document")
    return
}
```

The check happens after fetch but before any response body is built, and the
denial is indistinguishable from non-existence.

## Mitigations

1. Authorize every object access against the authenticated subject — never
   trust the identifier alone.
2. Prefer scoping queries (`WHERE owner_id = $1`) over post-fetch checks.
3. Collapse forbidden and unknown into one response to prevent oracles.
4. Avoid predictable identifiers where feasible (UUIDs raise enumeration cost
   but never replace authorization).

## Testing

Unit tests in `internal/labs/idor/lab_test.go` cover listing scope, cross-tenant
read success/failure per twin, own-document access through the safe twin, 404
uniformity, invalid IDs, and completion awarded only for foreign + confidential
reads on the vulnerable path.

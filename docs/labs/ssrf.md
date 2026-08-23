# Lab: SSRF — Inventory Webhook Preview

- **Slug:** `ssrf`
- **Category:** request-forgery
- **Difficulty:** medium
- **XP:** 150

## Objective

Use `/api/v1/targets/ssrf/preview?url=` to make the server fetch its own
internal inventory-sync service and read the file under `/secret/`. The
hardened twin validates every URL before fetching.

## Learning goal

Understand server-side request forgery: when a server fetches a client-supplied
URL, the client inherits the server's network position — including loopback and
private segments that should be unreachable from outside.

## Scenario

The inventory service offers a URL-preview feature that GETs whatever absolute
URL it is given. An internal stock-sync service runs on loopback (ephemeral
port) and serves synthetic training data, including a "credentials" file
containing the flag. Recon begins at `/targets/ssrf/config`, which reveals the
internal base URL.

**Isolation guarantee:** the internal service binds strictly to
`127.0.0.1:<ephemeral>`, serves only three hard-coded synthetic responses,
touches no cloud metadata services, holds no credentials, and shuts down with
the backend process. Nothing outside its own loopback socket is reachable
through this lab.

## Walkthrough (spoilers)

1. GET `/config` → learn the internal base URL, e.g. `http://127.0.0.1:34213`.
2. Ask the vulnerable previewer to fetch the internal secret:
   `GET /preview?url=http://127.0.0.1:34213/secret/credentials.txt`
   The server performs the request from *inside* and returns the body:
   `FLAG-SSRF-a17d55: …`.
3. Send the identical URL to `/preview-safe` → 400. Validation rejects
   loopback addresses before any network activity occurs.
4. Confirm the safe twin still fetches genuinely public URLs (e.g.
   `https://example.com/`) so the fix doesn't over-block.

## Impact

Reading internal-only services, cloud metadata credential theft (in real
deployments), pivoting into private networks, and port-scanning from the
server's vantage point.

## Secure implementation

`internal/labs/ssrf/lab.go`:

```go
func ValidatePublicURL(raw string) error // scheme allow-list + IP range deny-list
```

Applied on the safe twin before fetching; fetch errors are collapsed into a
generic 502 so dial failures never leak internal topology.

## Mitigations

1. Allow-list schemes (`http`/`https` only).
2. Resolve hostnames and validate the resulting IPs against loopback,
   link-local (169.254.0.0/16), private (RFC1918) and unspecified ranges.
3. Deny obvious internal hostnames (`.internal`, `.local`,
   `metadata.google.internal`).
4. Never echo transport-layer errors to clients.
5. Run outbound fetchers from an isolated network namespace where feasible.

## Testing

Unit tests in `internal/labs/ssrf/lab_test.go` include a full round trip
against the real loopback service: the safe twin refuses the loopback URL, the
vulnerable endpoint completes the SSRF, and completion is awarded once.
A blocked-URL matrix covers IPv6 loopback, RFC1918 ranges, link-local metadata
addresses, non-http schemes and internal hostnames.

# Lab: Open Redirect — Campaign Link Forwarder

- **Slug:** `open-redirect`
- **Category:** request-forgery
- **Difficulty:** easy
- **XP:** 100

## Objective

Abuse the forwarder at `/api/v1/targets/open-redirect/redirect?to=` to bounce a
victim toward the internal admin console (any host whose path is
`/admin/secret-token`) and read what it exposes. Then verify the hardened twin
only lets internal campaign slugs through.

## Learning goal

Understand how a redirect endpoint that trusts its destination parameter lets
attackers launder malicious links through your own domain — and why an
allow-list beats blacklisting for redirect validation.

## Scenario

A marketing tracker redirects users through `/redirect` to count campaign
clicks before sending them on their way. The destination arrives as a plain
query parameter and is forwarded without question. Legitimate campaigns use
root-relative paths (`/promo`, `/newsletter`) — nothing stops other values.

Because the lab runs headless, the "redirect" is reported rather than followed:
the response names where the victim would be sent, and completion triggers when
that destination points at the internal admin console.

## Walkthrough (spoilers)

1. `GET /campaigns` lists the legitimate slugs.
2. Point the forwarder anywhere:
   `GET /redirect?to=https://evil.example.com/phish` — accepted, 200. In a real
   app this would 302 victims to attacker infrastructure with your domain in
   the address bar.
3. Chain it to an internal-only host:
   `GET /redirect?to=http://10.0.0.9/admin/secret-token` — completion awarded.
4. The safe twin rejects absolute URLs, protocol-relative (`//evil.com`),
   and non-http schemes; only exactly `/promo` or `/newsletter` pass.

## Impact

Phishing that inherits your domain's trust, OAuth token theft via redirect_uri
manipulation, and bypassing SSRF-style filters by bouncing through trusted
hosts.

## Secure implementation

`internal/labs/redirect/lab.go`:

```go
func isInternalSlug(to string) bool // scheme/host/opaque must be empty, path ∈ allow-list
```

The safe twin parses the destination and requires: no scheme, no host, no
opaque part, no protocol-relative prefix, and a path matching the campaign
allow-list with no query or fragment.

## Mitigations

1. Prefer internal identifiers (campaign slugs) over raw URLs — map to
   destinations server-side.
2. If URLs must be accepted, enforce an exact-match allow-list of destinations.
3. Reject scheme-relative (`//host`) forms explicitly; parsers disagree on them.
4. Never rely on blocklists of known-bad hosts — they cannot enumerate the good.

## Testing

Unit tests cover external-target acceptance on the vulnerable path, completion
for the admin-console destination, safe-twin rejection of absolute/protocol-
relative/javascript URLs, internal slug acceptance, missing-parameter handling,
and meta validity.

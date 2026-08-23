# Lab: Stored XSS — Feedback Wall

- **Slug:** `xss`
- **Category:** injection
- **Difficulty:** easy
- **XP:** 100

## Objective

Post a comment containing `<script>FLAG-XSS-77b1e4</script>` through
`/api/v1/targets/xss/comments`, then load `/targets/xss/comments/feed` where
the vulnerable feed serves it back as raw markup. The sanitized twin at
`/targets/xss/comments-safe` neutralises the payload.

## Learning goal

Understand how stored (Type II) XSS differs from reflected XSS — the payload
persists and executes for *every* subsequent reader — and how sanitization at
write time plus escaping at read time form defense in depth.

## Scenario

A product feedback wall accepts comments and renders them inside an HTML page.
The vulnerable rendering path writes stored bodies straight into markup; the
author field is always escaped, only the body diverges. This models real-world
comment widgets, profile bios and support tickets.

## Walkthrough (spoilers)

1. POST a comment whose body contains the script payload to `/comments`.
2. GET `/comments/feed` — an HTML page. The raw `<script>` tag appears
   verbatim; in a browser it would execute for every visitor.
3. Repeat the same submission against `/comments-safe`: tags are stripped
   before storage, keeping inner text but removing markup.
4. Even pre-existing malicious rows are escaped by the safe feed's renderer
   (`&lt;script&gt;…`), demonstrating output encoding as a second layer.

## Impact

Session hijacking, credential theft, defacement, silent keylogging, and
worm-style propagation when payloads target other users (including admins).

## Secure implementation

`internal/labs/xss/store.go` and `render.go` contain both layers:

- Write-time: `SanitizeBody` strips markup before storage (safe twin only).
- Read-time: `bodyForFeed` escapes everything except explicitly-sanitized
  content — the single point where the two rendering paths diverge.

## Mitigations

1. Escape on output (`html.EscapeString` or template auto-escaping) as the
   default for all user-derived strings.
2. Sanitize rich input server-side with an allow-list parser if HTML must be
   preserved.
3. Add `Content-Security-Policy` headers to reduce exploit impact.
4. Prefer frameworks that auto-encode (React does) over raw innerHTML sinks.

## Testing

Unit tests in `internal/labs/xss/lab_test.go` prove:

- the vulnerable path stores and re-serves executable markup,
- the sanitizer strips tags while preserving inner text,
- the safe feed escapes even pre-existing payloads,
- completion is awarded only when the marker is genuinely stored.

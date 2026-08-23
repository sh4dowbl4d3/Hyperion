# Lab: SQL Injection — Contact Directory Search

- **Slug:** `sqli`
- **Category:** injection
- **Difficulty:** easy
- **XP:** 100

## Objective

Use the `/api/v1/targets/sqli/search?q=` endpoint to reveal the hidden
"Archive Vault" contact and recover its flag note. Then compare with
`/targets/sqli/search-safe`, which is immune.

## Learning goal

Understand how string-concatenated SQL allows an attacker to change the
structure of a query, and why parameterized statements eliminate the flaw
without losing functionality.

## Scenario

A staff directory search builds its SQL by pasting the user's `q` parameter
directly into a quoted `ILIKE` literal:

```sql
SELECT name, email, department, note FROM sqli_contacts
WHERE hidden = false AND name ILIKE '%<q>%' ORDER BY id;
```

One contact row is marked `hidden = true`. The application intends the WHERE
clause to keep it unreachable forever.

## Walkthrough (spoilers)

1. Search normally: `?q=ali` returns matching visible contacts.
2. Submit a quote to break out of the string literal: `q='` — the query
   becomes syntactically invalid and the endpoint returns a generic error.
3. Inject a tautology: `q=' OR true --`
   The clause becomes `name ILIKE '%' OR true --%'`, which is always true, so
   every row — including the hidden archive record — is returned.
4. Read the flag from the hidden record's note field.
5. Send the identical payload to `/search-safe`: zero results. The safe twin
   passes `q` as a bound parameter (`$1`) so its value can never alter the
   statement's structure.

## Impact

Authentication bypass, data exfiltration across tenant boundaries, and in
real systems: modification or destruction of data and, depending on DB
privileges, remote code execution.

## Secure implementation

`internal/labs/sqli/store.go` contains both paths side by side:

- Vulnerable: `BuildVulnerableQuery(q)` concatenates input with `fmt.Sprintf`.
- Safe: a fixed SQL text with `%s || $1 || %` placeholders bound via pgx.

## Mitigations

1. **Parameterized queries** for every statement, no exceptions.
2. Never interpolate user input into SQL, even "trusted" internal values.
3. Return generic errors on SQL failure — driver messages leak schema detail.
4. Least-privilege database accounts limit blast radius when injection occurs.

## Testing

Unit tests in `internal/labs/sqli/lab_test.go` prove:

- the vulnerable path stores/executes attacker-controlled structure,
- the safe twin passes input strictly as a parameter,
- malformed SQL produces a generic client-facing error (no driver leakage),
- completion is awarded only when the flag actually leaks via the vulnerable path.

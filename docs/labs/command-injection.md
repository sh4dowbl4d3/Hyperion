# Lab: Command Injection — Network Tool DNS Probe

- **Slug:** `command-injection`
- **Category:** injection
- **Difficulty:** medium
- **XP:** 150

## Objective

Inject a second command into `/api/v1/targets/command-injection/lookup?host=`
so the tool executes something beyond the DNS query (try appending
`; cat /etc/passwd`-style payloads) and recover the leaked output. The hardened
twin rejects any host that is not a plain hostname.

## Learning goal

Understand how user input concatenated into a shell command line lets attackers
run arbitrary operating-system commands, and why allow-list validation plus
avoiding shells entirely are the correct defenses.

## Scenario

A diagnostics page lets staff run a DNS lookup against a host they type in.
The lookup command is assembled by pasting the host straight onto the end of
it — so a semicolon starts a second command, exactly as it would in a shell.

**Isolation guarantee:** the lab does not spawn a real process. A simulator
reproduces the shell's metacharacter semantics (`;` splits commands, output
appends) against synthetic data only. No binaries run, no filesystem outside
lab memory is touched, no network calls are made.

## Walkthrough (spoilers)

1. Normal lookup works: `?host=example.org`.
2. Inject a second command:
   `?host=example.org;+cat+/etc/passwd`
   The output contains the tool's normal result **plus** the second command's
   output — including `FLAG-CMDINJ-6e4b18: …`.
3. Other separators work identically in principle (`|`, `&&`, backticks) — try
   them to see the pattern generalize.
4. The safe twin allow-lists plain hostnames (letters, digits, dots, dashes)
   and rejects every metacharacter: all injection attempts return 400.

## Impact

Full server compromise in real deployments: arbitrary command execution under
the application's account, leading to data theft, lateral movement, ransomware,
or persistence.

## Secure implementation

`internal/labs/cmdinj/store.go`:

```go
var hostPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9.-]{0,251}[a-zA-Z0-9])?$`)
```

The safe twin validates against this allow-list *and* explicitly refuses
shell metacharacters (`; | & $ \` newline) before any processing. The best real
fix is stronger still: never invoke a shell — call resolver libraries directly.

## Mitigations

1. Avoid shelling out; use language-native libraries (DNS resolver APIs).
2. If a command is unavoidable, pass user input as separate argv entries —
   never build command strings by concatenation.
3. Allow-list strict formats (hostnames, IDs) before use.
4. Run command-executing services under a dedicated least-privilege account.

## Testing

Unit tests prove: injected second-command output appears (and awards
completion), normal lookups stay clean, the safe twin blocks six
metacharacter families (`; | && $()` backticks, newlines) while accepting valid
hostnames, missing parameters return 400, and meta validity.

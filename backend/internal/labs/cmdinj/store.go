package cmdinj

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"hyperion/backend/internal/labs"
)

const LabSlug = "command-injection"

// FlagOutput is what the injected extra command "prints". The vulnerable
// executor simulates a shell: the `;` metacharacter splits a second command,
// exactly like a real injection, but nothing ever touches a real shell.
const FlagOutput = "FLAG-CMDINJ-6e4b18: second command output leaked"

var ErrEmptyHost = errors.New("host must not be empty")
var ErrInvalidHost = errors.New("host contains invalid characters")

// hostPattern is the safe twin's allow-list.
var hostPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9.-]{0,251}[a-zA-Z0-9])?$`)

type Completer interface {
	CompleteLab(ctx context.Context, userID, labSlug string) (*labs.CompletionResult, error)
}

// Simulator answers lookups without touching a real network or shell.
type Simulator struct {
	pool *pgxpool.Pool // reserved for future per-user lookup history
}

func NewSimulator() *Simulator {
	return &Simulator{}
}

type LookupResult struct {
	Host    string `json:"host"`
	Output  string `json:"output"`
	Flagged bool   `json:"-"`
}

// LookupVulnerable simulates `nslookup <host>` by naive string assembly. A
// `;` in the host starts a "second command"; if that command mentions the
// secret file, its output leaks — mirroring classic command injection while
// remaining fully synthetic and side-effect free.
func (s *Simulator) LookupVulnerable(_ context.Context, host string) (*LookupResult, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, ErrEmptyHost
	}
	output := fmt.Sprintf("Server:  lab-dns.internal\nAddress: 10.20.30.40\n\nName:    %s\n", host)

	flagged := false
	if idx := strings.Index(host, ";"); idx >= 0 {
		first := strings.TrimSpace(host[:idx])
		second := strings.TrimSpace(host[idx+1:])
		if first != "" {
			output = fmt.Sprintf("Server:  lab-dns.internal\n\nName:    %s\n", first)
		}
		if second != "" {
			output += fmt.Sprintf("\n[%s]\n", second)
			if strings.Contains(strings.ToLower(second), "secret") ||
				strings.Contains(second, "/etc/") {
				output += FlagOutput + "\n"
				flagged = true
			}
		}
	}
	return &LookupResult{Host: host, Output: output, Flagged: flagged}, nil
}

// LookupSafe validates the host against a strict allow-list before use.
func (s *Simulator) LookupSafe(_ context.Context, host string) (*LookupResult, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, ErrEmptyHost
	}
	if !hostPattern.MatchString(host) || strings.ContainsAny(host, ";|&$`\n\r") {
		return nil, ErrInvalidHost
	}
	return &LookupResult{
		Host:   host,
		Output: fmt.Sprintf("Server:  lab-dns.internal\nAddress: 10.20.30.40\n\nName:    %s\n", host),
	}, nil
}

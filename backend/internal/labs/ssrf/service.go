package ssrf

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// The internal lab service is a tiny in-process HTTP server bound to a random
// loopback port. It serves ONLY the synthetic training documents below — no
// credentials, no cloud metadata, no access to anything outside itself.

const (
	MetadataPath = "/lab-metadata.json"
	SecretPath   = "/secret/credentials.txt"

	FlagContent = "FLAG-SSRF-a17d55: internal service reached via SSRF"
)

// StartInternalService runs the synthetic internal service and returns its
// base URL (http://127.0.0.1:<port>).
func StartInternalService(ctx context.Context) (string, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"service":"inventory-sync","status":"healthy","docs":["/lab-metadata.json"]}`)
	})

	mux.HandleFunc(MetadataPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"region":"lab-local","replicas":2,"note":"synthetic training data only"}`)
	})

	mux.HandleFunc(SecretPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "%s\n", FlagContent)
	})

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	ln := newLoopbackListener()
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", portOf(ln))
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			// Service dies with the process; nothing to recover.
			_ = err
		}
	}()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := contextWithTimeout(5 * time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	return baseURL, nil
}

func IsInternalFlag(body string) bool {
	return strings.Contains(body, FlagContent)
}

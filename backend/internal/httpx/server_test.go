package httpx

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServeStopsOnContextCancel(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := NewServer("", mux, slog.New(slog.NewTextHandler(&nullWriter{}, nil)))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ctx, ln) }()

	waitForServer(t, ln.Addr().String())

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve() after cancel = %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return within 5s of context cancellation")
	}

	conn, dialErr := net.DialTimeout("tcp", ln.Addr().String(), 500*time.Millisecond)
	if dialErr == nil {
		conn.Close()
		t.Fatal("connection succeeded after shutdown; server still accepting")
	}
}

func waitForServer(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequest(http.MethodGet, "http://"+addr+"/ping", nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("server did not become ready within 3s")
}

type nullWriter struct{}

func (nullWriter) Write(p []byte) (int, error) { return len(p), nil }

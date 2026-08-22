package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const shutdownTimeout = 10 * time.Second

type Server struct {
	addr string
	srv  *http.Server
	log  *slog.Logger
}

func NewServer(addr string, handler http.Handler, log *slog.Logger) *Server {
	return &Server{
		addr: addr,
		srv: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		log: log,
	}
}

func (s *Server) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.log.Info("http server listening", slog.String("addr", ln.Addr().String()))
	return s.Serve(ctx, ln)
}

func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	serveErr := make(chan error, 1)
	go func() {
		if err := s.srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			return errors.Join(ctx.Err(), err)
		}
		s.log.Info("http server stopped gracefully")
		return nil
	}
}

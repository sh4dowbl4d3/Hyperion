package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"moderndvwa/backend/internal/api"
	"moderndvwa/backend/internal/config"
	"moderndvwa/backend/internal/httpx"
	"moderndvwa/backend/internal/logging"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	log, err := logging.New(cfg.LogLevel)
	if err != nil {
		return err
	}
	log.Info("configuration loaded",
		slog.String("env", cfg.Env),
		slog.String("addr", cfg.Addr),
		slog.String("log_level", cfg.LogLevel),
	)

	router := api.NewRouter(log)
	server := httpx.NewServer(cfg.Addr, router, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("starting moderndvwa api",
		slog.String("service", api.ServiceName),
		slog.String("version", api.ServiceVersion),
	)
	if err := server.Run(ctx); err != nil {
		return err
	}
	return nil
}

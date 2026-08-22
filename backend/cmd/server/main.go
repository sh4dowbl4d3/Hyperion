package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"moderndvwa/backend/internal/api"
	"moderndvwa/backend/internal/config"
	"moderndvwa/backend/internal/database"
	"moderndvwa/backend/internal/httpx"
	"moderndvwa/backend/internal/logging"
	"moderndvwa/backend/migrations"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer pool.Close()
	log.Info("database connection established")

	applied, err := database.MigrateUp(ctx, pool, migrations.FS())
	if err != nil {
		return err
	}
	if len(applied) > 0 {
		log.Info("migrations applied", slog.Any("applied", applied))
	}

	router := api.NewRouter(log, pool)
	server := httpx.NewServer(cfg.Addr, router, log)

	log.Info("starting moderndvwa api",
		slog.String("service", api.ServiceName),
		slog.String("version", api.ServiceVersion),
	)
	if err := server.Run(ctx); err != nil {
		return err
	}
	return nil
}

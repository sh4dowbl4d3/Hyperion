package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"moderndvwa/backend/internal/api"
	"moderndvwa/backend/internal/auth"
	"moderndvwa/backend/internal/config"
	"moderndvwa/backend/internal/database"
	"moderndvwa/backend/internal/httpx"
	"moderndvwa/backend/internal/labs"
	"moderndvwa/backend/internal/labs/idor"
	"moderndvwa/backend/internal/labs/jwtlab"
	"moderndvwa/backend/internal/labs/sqli"
	"moderndvwa/backend/internal/labs/xss"
	"moderndvwa/backend/internal/logging"
	"moderndvwa/backend/internal/users"
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

	userRepo := users.NewRepository(pool)
	authService := auth.NewService(userRepo, log)
	tokenService := auth.NewTokenService(cfg.JWTSecret, cfg.JWTTTL)

	registry := labs.NewRegistry()
	catalogRepo := labs.NewCatalogRepository(pool)
	progressStore := labs.NewProgressStore(pool)

	sqliStore := sqli.NewPgStore(pool)
	registry.MustRegister(sqli.NewLab(sqliStore, progressStore))

	xssStore := xss.NewPgStore(pool)
	registry.MustRegister(xss.NewLab(xssStore, progressStore))

	idorStore := idor.NewPgStore(pool)
	registry.MustRegister(idor.NewLab(idorStore, progressStore))

	registry.MustRegister(jwtlab.NewLab(progressStore))

	if err := catalogRepo.Seed(ctx, registryMetas(registry)); err != nil {
		return err
	}

	router := api.NewRouter(api.Deps{
		Log:      log,
		DB:       pool,
		Auth:     authService,
		Tokens:   tokenService,
		Users:    userRepo,
		Registry: registry,
		Catalog:  catalogRepo,
		Progress: progressStore,
	})
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

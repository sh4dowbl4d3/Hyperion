package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"hyperion/backend/internal/api"
	"hyperion/backend/internal/auth"
	"hyperion/backend/internal/config"
	"hyperion/backend/internal/database"
	"hyperion/backend/internal/httpx"
	"hyperion/backend/internal/labs"
	"hyperion/backend/internal/labs/cmdinj"
	"hyperion/backend/internal/labs/idor"
	"hyperion/backend/internal/labs/jwtlab"
	"hyperion/backend/internal/labs/redirect"
	"hyperion/backend/internal/labs/race"
	"hyperion/backend/internal/labs/ssrf"
	"hyperion/backend/internal/labs/sqli"
	"hyperion/backend/internal/labs/xss"
	"hyperion/backend/internal/logging"
	"hyperion/backend/internal/users"
	"hyperion/backend/migrations"
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

	ssrfCtx, ssrfCancel := context.WithCancel(ctx)
	defer ssrfCancel()
	internalBaseURL, err := ssrf.StartInternalService(ssrfCtx)
	if err != nil {
		return fmt.Errorf("start internal lab service: %w", err)
	}
	log.Info("internal lab service listening", slog.String("base_url", internalBaseURL))
	registry.MustRegister(ssrf.NewLab(ssrf.NewHTTPFetcher(), progressStore, internalBaseURL))

	raceStore := race.NewStore(pool)
	registry.MustRegister(race.NewLab(raceStore, progressStore))

	registry.MustRegister(redirect.NewLab(progressStore))
	registry.MustRegister(cmdinj.NewLab(cmdinj.NewSimulator(), progressStore))

	if err := catalogRepo.Seed(ctx, registryMetas(registry)); err != nil {
		return err
	}
	// Prune catalog rows left behind by removed labs or test runs so the
	// listing always reflects the compiled-in registry.
	if err := catalogRepo.PruneMissing(ctx, registryMetas(registry)); err != nil {
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

	log.Info("starting hyperion api",
		slog.String("service", api.ServiceName),
		slog.String("version", api.ServiceVersion),
	)
	if err := server.Run(ctx); err != nil {
		return err
	}
	return nil
}

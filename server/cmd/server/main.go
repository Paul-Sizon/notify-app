package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/paulsizon/notify/server/internal/agent"
	"github.com/paulsizon/notify/server/internal/api"
	"github.com/paulsizon/notify/server/internal/config"
	"github.com/paulsizon/notify/server/internal/db"
	"github.com/paulsizon/notify/server/internal/push"
	"github.com/paulsizon/notify/server/internal/scheduler"
)

func main() {
	once := flag.String("once", "", "run a single subscription end-to-end and exit (UUID)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	d := db.New(pool)

	searcher := agent.NewBraveClient(cfg.BraveKey)
	extractor := agent.NewOpenAIExtractor(cfg.OpenAIKey)

	var pusher push.Pusher
	if cfg.APNs.Configured() {
		ap, err := push.NewAPNsPusher(push.APNsConfig{
			KeyPath:    cfg.APNs.KeyPath,
			KeyID:      cfg.APNs.KeyID,
			TeamID:     cfg.APNs.TeamID,
			BundleID:   cfg.APNs.BundleID,
			Production: cfg.APNs.Production,
		})
		if err != nil {
			logger.Error("apns init", "err", err)
			os.Exit(1)
		}
		pusher = ap
		logger.Info("apns configured", "production", cfg.APNs.Production)
	} else {
		pusher = &push.LogPusher{Logger: logger}
		logger.Warn("apns not configured — using log-only stub pusher")
	}

	deps := agent.Deps{
		DB:        d,
		Searcher:  searcher,
		Planner:   extractor,
		Extractor: extractor,
		Pusher:    pusher,
	}
	runner := func(ctx context.Context, subID uuid.UUID) ([]uuid.UUID, error) {
		return agent.RunSubscription(ctx, deps, subID)
	}

	if *once != "" {
		id, err := uuid.Parse(*once)
		if err != nil {
			logger.Error("parse uuid", "err", err)
			os.Exit(1)
		}
		ids, err := runner(ctx, id)
		if err != nil {
			logger.Error("run failed", "err", err)
			os.Exit(1)
		}
		logger.Info("run done", "new_signals", len(ids), "ids", ids)
		return
	}

	sched := &scheduler.Scheduler{
		DB:          d,
		Run:         runner,
		Interval:    10 * time.Second,
		BatchLimit:  20,
		Concurrency: 4,
		Logger:      logger,
	}
	go sched.Loop(ctx)

	h := api.NewHandler(d, runner).
		WithSuggester(agent.NewOpenAISuggester(cfg.OpenAIKey)).
		WithAdminToken(cfg.AdminToken)
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		logger.Info("http listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

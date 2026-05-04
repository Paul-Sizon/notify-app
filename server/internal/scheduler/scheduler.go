package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/paulsizon/notify/server/internal/db"
)

// Runner runs the agent pipeline for one subscription. RunSubscription matches.
type Runner func(ctx context.Context, subID uuid.UUID) ([]uuid.UUID, error)

type Scheduler struct {
	DB          *db.DB
	Run         Runner
	Now         func() time.Time
	Interval    time.Duration
	BatchLimit  int
	Concurrency int
	Logger      *slog.Logger

	inflight sync.Map // subID -> struct{}
}

func (s *Scheduler) defaults() {
	if s.Now == nil {
		s.Now = time.Now
	}
	if s.Interval == 0 {
		s.Interval = 10 * time.Second
	}
	if s.BatchLimit == 0 {
		s.BatchLimit = 20
	}
	if s.Concurrency == 0 {
		s.Concurrency = 4
	}
	if s.Logger == nil {
		s.Logger = slog.Default()
	}
}

// Tick picks due subscriptions and runs them in parallel up to Concurrency.
// Blocks until all spawned runs complete (so tests are deterministic).
func (s *Scheduler) Tick(ctx context.Context) error {
	s.defaults()
	due, err := s.DB.ListDueSubscriptions(ctx, s.Now(), s.BatchLimit)
	if err != nil {
		return err
	}
	sem := make(chan struct{}, s.Concurrency)
	var wg sync.WaitGroup
	for _, sub := range due {
		if _, loaded := s.inflight.LoadOrStore(sub.ID, struct{}{}); loaded {
			s.Logger.Debug("skip in-flight", "sub", sub.ID)
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(id uuid.UUID) {
			defer wg.Done()
			defer func() { <-sem }()
			defer s.inflight.Delete(id)
			if _, err := s.Run(ctx, id); err != nil {
				s.Logger.Error("run failed", "sub", id, "err", err)
			}
		}(sub.ID)
	}
	wg.Wait()
	return nil
}

// Loop runs Tick on Interval until ctx is done.
func (s *Scheduler) Loop(ctx context.Context) {
	s.defaults()
	t := time.NewTicker(s.Interval)
	defer t.Stop()
	if err := s.Tick(ctx); err != nil {
		s.Logger.Error("tick", "err", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := s.Tick(ctx); err != nil {
				s.Logger.Error("tick", "err", err)
			}
		}
	}
}

//go:build integration

package scheduler_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/paulsizon/notify/server/internal/db"
	"github.com/paulsizon/notify/server/internal/scheduler"
	"github.com/paulsizon/notify/server/internal/testhelpers"
)

func TestScheduler_PicksUpDueAndSkipsFuture(t *testing.T) {
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	ctx := context.Background()

	devID, _ := d.UpsertDevice(ctx, "tok-1")
	subDue, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devID, Query: "due", Type: "event", CadenceSeconds: 3600,
	})
	subFuture, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devID, Query: "later", Type: "event", CadenceSeconds: 3600,
	})

	// push subFuture into the future
	_, err := pool.Exec(ctx, `UPDATE subscriptions SET next_run_at = NOW() + INTERVAL '10 minutes' WHERE id = $1`, subFuture.ID)
	require.NoError(t, err)

	var (
		mu   sync.Mutex
		runs []uuid.UUID
	)
	runner := func(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
		mu.Lock()
		runs = append(runs, id)
		mu.Unlock()
		// simulate the runner rescheduling so it isn't picked up again immediately
		_ = d.RescheduleSubscription(ctx, id, time.Now(), "")
		return nil, nil
	}
	s := &scheduler.Scheduler{DB: d, Run: runner, BatchLimit: 10, Concurrency: 4}

	require.NoError(t, s.Tick(ctx))
	require.Len(t, runs, 1)
	require.Equal(t, subDue.ID, runs[0])
}

func TestScheduler_DoesNotDoubleRunInflight(t *testing.T) {
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	ctx := context.Background()

	devID, _ := d.UpsertDevice(ctx, "tok-1")
	sub, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devID, Query: "x", Type: "event", CadenceSeconds: 3600,
	})

	gate := make(chan struct{})
	var calls int
	var mu sync.Mutex
	runner := func(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		<-gate
		return nil, nil
	}
	s := &scheduler.Scheduler{DB: d, Run: runner, BatchLimit: 10, Concurrency: 4}

	// kick first tick in background — runner blocks on gate
	done := make(chan struct{})
	go func() { _ = s.Tick(ctx); close(done) }()

	// wait until runner has been invoked
	require.Eventually(t, func() bool {
		mu.Lock(); defer mu.Unlock(); return calls == 1
	}, time.Second, 5*time.Millisecond)

	// sub still due (runner hasn't rescheduled yet because it's blocked).
	// second Tick must skip it because of inflight.
	require.NoError(t, s.Tick(ctx))
	mu.Lock()
	require.Equal(t, 1, calls, "second tick should not invoke runner while first is in flight")
	mu.Unlock()

	close(gate)
	<-done

	_ = sub
}

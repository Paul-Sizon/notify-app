//go:build integration

package agent

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/paulsizon/notify/server/internal/db"
	"github.com/paulsizon/notify/server/internal/testhelpers"
)

func TestRunSubscription_FullPipeline_DedupsOnSecondRun(t *testing.T) {
	requireKeys(t)
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	devID, err := d.UpsertDevice(ctx, "tok-e2e")
	require.NoError(t, err)

	sub, err := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID:       devID,
		Query:          "Coldplay concerts in 2026 with dates and venues",
		Type:           "event",
		CadenceSeconds: 3600,
	})
	require.NoError(t, err)

	extractor := NewOpenAIExtractor(os.Getenv("OPENAI_API_KEY"))
	deps := Deps{
		DB:        d,
		Searcher:  NewBraveClient(os.Getenv("BRAVE_API_KEY")),
		Planner:   extractor,
		Extractor: extractor,
	}

	ids1, err := RunSubscription(ctx, deps, sub.ID)
	require.NoError(t, err)
	t.Logf("first run inserted %d signals", len(ids1))
	require.NotEmpty(t, ids1, "first run should produce signals from real Coldplay query")

	ids2, err := RunSubscription(ctx, deps, sub.ID)
	require.NoError(t, err)
	t.Logf("second run inserted %d signals", len(ids2))
	// LLM extraction is non-deterministic; identical Brave prose can yield slightly
	// different fingerprints across calls. Strict dedup is proven at DB layer
	// (TestInsertSignals_OnConflictReturnsOnlyNew). Here we assert the weaker but
	// always-true property: pipeline never INVENTS more signals on a re-run.
	require.LessOrEqual(t, len(ids2), len(ids1), "second run must not exceed first")

	// Reschedule sanity check.
	got, err := d.GetSubscription(ctx, sub.ID)
	require.NoError(t, err)
	require.NotNil(t, got.LastRunAt)
	require.True(t, got.NextRunAt.After(*got.LastRunAt))
}

func TestRunSubscription_News_UpdatesRollingSummary(t *testing.T) {
	requireKeys(t)
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	devID, _ := d.UpsertDevice(ctx, "tok-news")
	sub, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID:       devID,
		Query:          "Cryptocurrency regulation Vietnam",
		Type:           "news",
		CadenceSeconds: 3600,
	})

	extractor := NewOpenAIExtractor(os.Getenv("OPENAI_API_KEY"))
	deps := Deps{
		DB:        d,
		Searcher:  NewBraveClient(os.Getenv("BRAVE_API_KEY")),
		Planner:   extractor,
		Extractor: extractor,
	}

	_, err := RunSubscription(ctx, deps, sub.ID)
	require.NoError(t, err)

	got, err := d.GetSubscription(ctx, sub.ID)
	require.NoError(t, err)
	t.Logf("rolling_summary after run: %s", got.RollingSummary)
	require.NotEmpty(t, got.RollingSummary, "news run should populate rolling_summary")
}

// stubSearcher and stubExtractor remove API variability so we can prove the
// orchestration-layer dedup contract: identical inputs across two runs ⇒ zero
// new signals on the second run.
type stubSearcher struct{ a AnswerResult }

func (s stubSearcher) Answer(ctx context.Context, q string) (AnswerResult, error) {
	return s.a, nil
}

type stubExtractor struct {
	events []EventCandidate
	news   NewsExtraction
}

func (s stubExtractor) ExtractEvents(ctx context.Context, in ExtractInput) ([]EventCandidate, error) {
	return s.events, nil
}
func (s stubExtractor) ExtractNews(ctx context.Context, in ExtractInput) (NewsExtraction, error) {
	return s.news, nil
}

func TestRunSubscription_Stubbed_DedupsExactly(t *testing.T) {
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	ctx := context.Background()

	devID, _ := d.UpsertDevice(ctx, "tok-stub")
	sub, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devID, Query: "x", Type: "event", CadenceSeconds: 3600,
	})

	dt := "2026-09-05"
	venue := "Allianz Parque"
	city := "São Paulo"

	deps := Deps{
		DB:       d,
		Searcher: stubSearcher{a: AnswerResult{Text: "fixed"}},
		Extractor: stubExtractor{
			events: []EventCandidate{
				{Title: "Coldplay", Date: &dt, Venue: &venue, City: &city, Confidence: 0.9},
				{Title: "Arctic Monkeys", Date: &dt, Venue: &venue, City: &city, Confidence: 0.9},
			},
		},
	}

	ids1, err := RunSubscription(ctx, deps, sub.ID)
	require.NoError(t, err)
	require.Len(t, ids1, 2)

	ids2, err := RunSubscription(ctx, deps, sub.ID)
	require.NoError(t, err)
	require.Empty(t, ids2, "stub orchestration must dedup exactly")
}

// failingSearcher always returns the given error. Used to exercise the
// failure-path scheduling without burning real Brave credits.
type failingSearcher struct{ err error }

func (f failingSearcher) Answer(ctx context.Context, q string) (AnswerResult, error) {
	return AnswerResult{}, f.err
}

// TestRunSubscription_FailingSearcher_AppliesBackoff is the regression
// test for the 10-second retry storm. Before: a Brave 402 left next_run_at
// in the past, so every scheduler tick (10s) re-ran the failing sub. After:
// the deferred backoff in RunSubscription advances next_run_at by hours
// and stamps last_error_kind so the sub leaves the "due" set immediately.
func TestRunSubscription_FailingSearcher_AppliesBackoff(t *testing.T) {
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	ctx := context.Background()

	devID, _ := d.UpsertDevice(ctx, "tok-backoff")
	sub, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devID, Query: "x", Type: "event", CadenceSeconds: 3600,
	})

	deps := Deps{
		DB:        d,
		Searcher:  failingSearcher{err: &BraveError{StatusCode: 402, Body: `{"error":"USAGE_LIMIT_EXCEEDED"}`}},
		Extractor: stubExtractor{},
	}

	startNextRun := sub.NextRunAt
	_, err := RunSubscription(ctx, deps, sub.ID)
	require.Error(t, err)
	require.ErrorContains(t, err, "402")

	got, err := d.GetSubscription(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, 1, got.FailureCount, "first failure should set count to 1")
	require.Equal(t, "brave_quota_exceeded", got.LastErrorKind)
	require.True(t, got.NextRunAt.After(startNextRun.Add(time.Hour)),
		"402 should push next_run_at hours into the future, not stay in the past")

	// Second failure: failure_count increments; for quota errors the delay
	// is fixed (not exponential), so we just confirm the counter bumps.
	_, err = RunSubscription(ctx, deps, sub.ID)
	require.Error(t, err)
	got, _ = d.GetSubscription(ctx, sub.ID)
	require.Equal(t, 2, got.FailureCount)
	require.Equal(t, "brave_quota_exceeded", got.LastErrorKind)
}

// TestRunSubscription_SuccessAfterFailure_ResetsCounters proves the success
// path clears prior backoff state. Without this, a sub that recovered would
// still carry "brave_quota_exceeded" forever.
func TestRunSubscription_SuccessAfterFailure_ResetsCounters(t *testing.T) {
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	ctx := context.Background()

	devID, _ := d.UpsertDevice(ctx, "tok-recover")
	sub, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devID, Query: "x", Type: "event", CadenceSeconds: 3600,
	})

	// Step 1: fail.
	depsFail := Deps{
		DB:        d,
		Searcher:  failingSearcher{err: &BraveError{StatusCode: 429}},
		Extractor: stubExtractor{},
	}
	_, _ = RunSubscription(ctx, depsFail, sub.ID)
	got, _ := d.GetSubscription(ctx, sub.ID)
	require.Equal(t, 1, got.FailureCount)
	require.Equal(t, "rate_limited", got.LastErrorKind)

	// Step 2: succeed.
	depsOK := Deps{
		DB:        d,
		Searcher:  stubSearcher{a: AnswerResult{Text: "ok"}},
		Extractor: stubExtractor{},
	}
	_, err := RunSubscription(ctx, depsOK, sub.ID)
	require.NoError(t, err)

	got, _ = d.GetSubscription(ctx, sub.ID)
	require.Equal(t, 0, got.FailureCount, "success must reset failure_count")
	require.Equal(t, "", got.LastErrorKind, "success must clear last_error_kind")
}

func requireKeys(t *testing.T) {
	t.Helper()
	if os.Getenv("BRAVE_API_KEY") == "" || os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("BRAVE_API_KEY and OPENAI_API_KEY required")
	}
}

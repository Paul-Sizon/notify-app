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

	deps := Deps{
		DB:        d,
		Searcher:  NewBraveClient(os.Getenv("BRAVE_API_KEY")),
		Extractor: NewOpenAIExtractor(os.Getenv("OPENAI_API_KEY")),
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

	deps := Deps{
		DB:        d,
		Searcher:  NewBraveClient(os.Getenv("BRAVE_API_KEY")),
		Extractor: NewOpenAIExtractor(os.Getenv("OPENAI_API_KEY")),
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

func requireKeys(t *testing.T) {
	t.Helper()
	if os.Getenv("BRAVE_API_KEY") == "" || os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("BRAVE_API_KEY and OPENAI_API_KEY required")
	}
}

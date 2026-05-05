//go:build e2e

package e2e_test

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/paulsizon/notify/server/e2e"
	"github.com/paulsizon/notify/server/internal/agent"
	"github.com/paulsizon/notify/server/internal/api"
	"github.com/paulsizon/notify/server/internal/db"
	"github.com/paulsizon/notify/server/internal/push"
	"github.com/paulsizon/notify/server/internal/testhelpers"
)

// TestE2E_FullStack_RealAPIs hits real Brave + OpenAI through the HTTP API.
// Costs: ~1 Brave call, ~2 OpenAI calls per run. ~$0.05 total.
// Run with: make test-e2e
func TestE2E_FullStack_RealAPIs(t *testing.T) {
	if os.Getenv("BRAVE_SEARCH_API_KEY") == "" || os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("BRAVE_SEARCH_API_KEY and OPENAI_API_KEY required")
	}
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)

	extractor := agent.NewOpenAIExtractor(os.Getenv("OPENAI_API_KEY"))
	deps := agent.Deps{
		DB:        d,
		Searcher:  agent.NewBraveClient(os.Getenv("BRAVE_SEARCH_API_KEY")),
		Planner:   extractor,
		Extractor: extractor,
		Pusher:    &push.LogPusher{},
	}
	runner := func(ctx context.Context, subID uuid.UUID) ([]uuid.UUID, error) {
		return agent.RunSubscription(ctx, deps, subID)
	}

	h := api.NewHandler(d, runner)
	srv := httptest.NewServer(h.Routes())
	defer srv.Close()

	c := e2e.NewClient(srv.URL)

	_, err := c.RegisterDevice("real-apns-token")
	require.NoError(t, err)

	sub, err := c.CreateSubscription("Upcoming Coldplay concerts in 2026 with dates and venues", "event", 3600)
	require.NoError(t, err)

	run, err := c.RunSubscription(sub.ID)
	require.NoError(t, err)
	t.Logf("real-API run produced %d new signals", run.NewSignals)

	sigs, err := c.ListSignals(sub.ID)
	require.NoError(t, err)
	t.Logf("listed %d signals", len(sigs))
	for i, s := range sigs {
		t.Logf("  [%d] %s — conf=%.2f", i, s.Title, s.Confidence)
	}
	// Brave's prose can occasionally lack any concrete events for a query.
	// We assert the pipeline executed cleanly; signal volume is best-effort.
	require.Equal(t, run.NewSignals, len(sigs))
	if run.NewSignals == 0 {
		t.Log("Brave returned no extractable events this run; pipeline still healthy")
	}
}

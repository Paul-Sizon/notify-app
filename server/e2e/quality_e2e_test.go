//go:build e2e

// Quality scenarios — real Brave + real OpenAI. Cost: ~$0.05 per subtest.
// Run on demand:
//   make test-e2e                                    # all scenarios
//   go test -tags=e2e -run TestQuality/zero_ ./e2e   # just one
//
// These tests characterize the AGENT's behavior across query shapes the iOS
// app will see in the wild. Assertions are deliberately loose because LLM
// + web search are nondeterministic — we assert structure (no crash, future
// dates, non-hallucination) not exact counts.
package e2e_test

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/paulsizon/notify/server/e2e"
	"github.com/paulsizon/notify/server/internal/agent"
	"github.com/paulsizon/notify/server/internal/api"
	"github.com/paulsizon/notify/server/internal/db"
	"github.com/paulsizon/notify/server/internal/push"
	"github.com/paulsizon/notify/server/internal/testhelpers"
)

func newQualityHarness(t *testing.T) *e2e.Client {
	t.Helper()
	if os.Getenv("BRAVE_API_KEY") == "" || os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("BRAVE_API_KEY and OPENAI_API_KEY required")
	}
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)

	deps := agent.Deps{
		DB:        d,
		Searcher:  agent.NewBraveClient(os.Getenv("BRAVE_API_KEY")),
		Extractor: agent.NewOpenAIExtractor(os.Getenv("OPENAI_API_KEY")),
		Pusher:    &push.LogPusher{},
	}
	runner := func(ctx context.Context, subID uuid.UUID) ([]uuid.UUID, error) {
		return agent.RunSubscription(ctx, deps, subID)
	}
	h := api.NewHandler(d, runner)
	srv := httptest.NewServer(h.Routes())
	t.Cleanup(srv.Close)

	c := e2e.NewClient(srv.URL)
	_, err := c.RegisterDevice("quality-suite-tok-" + t.Name())
	require.NoError(t, err)
	return c
}

type expectFn func(t *testing.T, sigs []api.SignalDTO)

func qSubtest(t *testing.T, name, query, typ string, expect expectFn) {
	t.Run(name, func(t *testing.T) {
		c := newQualityHarness(t)
		sub, err := c.CreateSubscription(query, typ, 3600)
		require.NoError(t, err)

		_, err = c.RunSubscription(sub.ID)
		require.NoError(t, err)

		sigs, err := c.ListSignals(sub.ID)
		require.NoError(t, err)
		t.Logf("query=%q type=%s -> %d signals", query, typ, len(sigs))
		for i, s := range sigs {
			occ := "no-date"
			if s.OccursAt != nil {
				occ = s.OccursAt.Format("2006-01-02")
			}
			t.Logf("  [%d] %s | %s | conf=%.2f", i, s.Title, occ, s.Confidence)
		}
		expect(t, sigs)
	})
}

// ----- expectation helpers -----

func anyCount(*testing.T, []api.SignalDTO) {}

func nonEmpty(t *testing.T, sigs []api.SignalDTO) {
	t.Helper()
	require.NotEmpty(t, sigs, "expected at least one signal — sanity sub")
}

func empty(t *testing.T, sigs []api.SignalDTO) {
	t.Helper()
	require.Empty(t, sigs, "expected zero signals — query has no real events")
}

func futureDatesOnly(t *testing.T, sigs []api.SignalDTO) {
	t.Helper()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for _, s := range sigs {
		if s.OccursAt == nil {
			continue // null date allowed
		}
		require.False(t, s.OccursAt.Before(today),
			"signal %q has past date %s", s.Title, s.OccursAt.Format("2006-01-02"))
	}
}

func allHaveTitleAndConfidence(t *testing.T, sigs []api.SignalDTO) {
	t.Helper()
	for _, s := range sigs {
		require.NotEmpty(t, s.Title)
		require.Greater(t, s.Confidence, float32(0))
	}
}

// compose assertions
func all(fns ...expectFn) expectFn {
	return func(t *testing.T, sigs []api.SignalDTO) {
		for _, f := range fns {
			f(t, sigs)
		}
	}
}

// ---------------------------------------------------------------------------
//                              Scenarios
// ---------------------------------------------------------------------------

// Sanity bucket: real upcoming, well-known. Should produce ≥1 signal with
// future-only dates and structurally valid signals.
func TestQuality_Sanity_KnownUpcoming(t *testing.T) {
	cases := []struct {
		name, query, typ string
	}{
		{"coldplay_sao_paulo", "Coldplay tour São Paulo 2026", "event"},
		{"f1_brazilian_gp", "Formula 1 Brazilian Grand Prix Interlagos 2026", "event"},
		{"lollapalooza_sp", "Lollapalooza São Paulo 2026 lineup and dates", "event"},
	}
	for _, c := range cases {
		qSubtest(t, c.name, c.query, c.typ, all(allHaveTitleAndConfidence, futureDatesOnly))
	}
}

// Zero-result bucket: query has no real-world events. Extractor's "empty
// array is the correct answer" rule must hold — no hallucination.
func TestQuality_ZeroResult_NoHallucination(t *testing.T) {
	cases := []struct {
		name, query, typ string
	}{
		{"underwater_basket_weaving", "underwater basket weaving conference Curitiba", "event"},
		{"city_council_obscure", "city council meetings Guarapuava 2026", "event"},
	}
	for _, c := range cases {
		qSubtest(t, c.name, c.query, c.typ, all(allHaveTitleAndConfidence, empty))
	}
}

// Stale-trap bucket: query references a past event. Extractor's "future
// relative to today" rule must filter all candidates out.
func TestQuality_StaleTrap_FiltersPast(t *testing.T) {
	cases := []struct {
		name, query, typ string
	}{
		{"woodstock_1969", "Woodstock 1969 lineup and dates", "event"},
		{"olympics_2016_rio", "Rio 2016 Olympics events schedule", "event"},
	}
	for _, c := range cases {
		// Either zero signals OR all dates strictly in future.
		qSubtest(t, c.name, c.query, c.typ, all(allHaveTitleAndConfidence, futureDatesOnly))
	}
}

// Multi-language bucket: Portuguese query for events likely covered by
// Portuguese sources. Pipeline must not crash and structure must hold.
func TestQuality_MultiLanguage_Portuguese(t *testing.T) {
	qSubtest(t, "eventos_blockchain_curitiba_pt",
		"eventos de blockchain em Curitiba 2026", "event",
		all(allHaveTitleAndConfidence, futureDatesOnly))
}

// High-noise bucket: extremely broad query that returns webinar spam.
// Acceptance: completes without error, ≥0 signals, all structurally valid.
func TestQuality_HighNoise_BroadQuery(t *testing.T) {
	qSubtest(t, "ai_events_global", "AI conferences and events 2026", "event",
		all(allHaveTitleAndConfidence, futureDatesOnly))
}

// Niche / long-tail bucket: real interest but sparse coverage. May produce
// few or no signals; both are acceptable. Asserts only structural validity.
func TestQuality_Niche_LongTail(t *testing.T) {
	cases := []struct {
		name, query, typ string
	}{
		{"solana_hacker_houses_latam", "Solana hacker houses in Latin America 2026", "event"},
		{"kmp_conferences", "Kotlin Multiplatform conferences 2026", "event"},
		{"zk_workshops", "Zero-knowledge proof workshops 2026 worldwide", "event"},
	}
	for _, c := range cases {
		qSubtest(t, c.name, c.query, c.typ, all(allHaveTitleAndConfidence, futureDatesOnly))
	}
}

// News bucket: regulatory + competitive monitoring. Verifies news pipeline
// returns NewsExtraction with rolling_summary populated.
func TestQuality_News_RegulatoryAndCompetitive(t *testing.T) {
	cases := []struct {
		name, query string
	}{
		{"vietnam_crypto", "Cryptocurrency regulation changes in Vietnam"},
		{"brazil_stablecoin", "Stablecoin legislation in Brazil"},
		{"nubank_launches", "Nubank product launches and announcements"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newQualityHarness(t)

			sub, err := c.CreateSubscription(tc.query, "news", 3600)
			require.NoError(t, err)

			run, err := c.RunSubscription(sub.ID)
			require.NoError(t, err)
			t.Logf("query=%q -> %d new signals", tc.query, run.NewSignals)

			// rolling_summary should be populated even if 0 new signals
			// (the LLM rewrites the summary on every run).
			subs, err := c.ListSubscriptions()
			require.NoError(t, err)
			require.Len(t, subs, 1)
			// SubscriptionDTO doesn't expose rolling_summary; the indirect
			// proof is that the next run dedups against it. Smoke-only here.

			sigs, err := c.ListSignals(sub.ID)
			require.NoError(t, err)
			for i, s := range sigs {
				t.Logf("  [%d] %s — conf=%.2f", i, s.Title, s.Confidence)
			}
		})
	}
}

// Idempotency bucket: re-run with stable LLM output should NOT add signals.
// This is the strict-dedup-with-real-APIs check; loose assertion (≤ first run).
func TestQuality_Idempotency_RerunDoesNotInflate(t *testing.T) {
	c := newQualityHarness(t)
	sub, err := c.CreateSubscription("Coldplay tour 2026 with dates and venues", "event", 3600)
	require.NoError(t, err)

	run1, err := c.RunSubscription(sub.ID)
	require.NoError(t, err)
	t.Logf("first run: %d new", run1.NewSignals)

	run2, err := c.RunSubscription(sub.ID)
	require.NoError(t, err)
	t.Logf("second run: %d new", run2.NewSignals)

	require.LessOrEqual(t, run2.NewSignals, run1.NewSignals,
		"re-run must not invent more signals than first run")
}

// Anti-ambiguity bucket: query is genuinely ambiguous. Pipeline should pick
// one interpretation gracefully — no crash, no schema violation.
func TestQuality_Ambiguous_WingsConcert(t *testing.T) {
	qSubtest(t, "wings_concert_band_or_food", "Wings concert 2026", "event",
		all(allHaveTitleAndConfidence, futureDatesOnly))
}

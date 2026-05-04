//go:build integration

package e2e_test

import (
	"context"
	"net/http/httptest"
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

type stubSearcher struct{}

func (stubSearcher) Answer(ctx context.Context, q string) (agent.AnswerResult, error) {
	return agent.AnswerResult{Text: "Coldplay 2026-09-05 at Allianz Parque, São Paulo. Coldplay 2026-09-08 at Estádio Nilton Santos, Rio de Janeiro."}, nil
}

type stubExtractor struct{}

func sP(s string) *string { return &s }

func (stubExtractor) ExtractEvents(ctx context.Context, in agent.ExtractInput) ([]agent.EventCandidate, error) {
	return []agent.EventCandidate{
		{Title: "Coldplay", Date: sP("2026-09-05"), Venue: sP("Allianz Parque"), City: sP("São Paulo"), Confidence: 0.95},
		{Title: "Coldplay", Date: sP("2026-09-08"), Venue: sP("Estádio Nilton Santos"), City: sP("Rio de Janeiro"), Confidence: 0.95},
	}, nil
}
func (stubExtractor) ExtractNews(ctx context.Context, in agent.ExtractInput) (agent.NewsExtraction, error) {
	return agent.NewsExtraction{}, nil
}

// TestE2E_Stubbed exercises the full HTTP surface against the in-process server
// using stub Searcher/Extractor. No external API costs.
func TestE2E_Stubbed_FullFlow(t *testing.T) {
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)

	deps := agent.Deps{
		DB:        d,
		Searcher:  stubSearcher{},
		Extractor: stubExtractor{},
		Pusher:    &push.LogPusher{},
	}
	runner := func(ctx context.Context, subID uuid.UUID) ([]uuid.UUID, error) {
		return agent.RunSubscription(ctx, deps, subID)
	}

	h := api.NewHandler(d, runner)
	srv := httptest.NewServer(h.Routes())
	defer srv.Close()

	c := e2e.NewClient(srv.URL)

	// 1. register device
	devID, err := c.RegisterDevice("apns-token-e2e")
	require.NoError(t, err)
	require.NotEmpty(t, devID)

	// 2. idempotent re-register
	devID2, err := c.RegisterDevice("apns-token-e2e")
	require.NoError(t, err)
	require.Equal(t, devID, devID2)

	// 3. create subscription
	sub, err := c.CreateSubscription("Coldplay concerts 2026", "event", 3600)
	require.NoError(t, err)
	require.NotEmpty(t, sub.ID)

	// 4. list — one sub
	list, err := c.ListSubscriptions()
	require.NoError(t, err)
	require.Len(t, list, 1)

	// 5. run sub manually — expect 2 signals
	run, err := c.RunSubscription(sub.ID)
	require.NoError(t, err)
	require.Equal(t, 2, run.NewSignals)

	// 6. signals visible
	sigs, err := c.ListSignals(sub.ID)
	require.NoError(t, err)
	require.Len(t, sigs, 2)
	for _, s := range sigs {
		require.NotEmpty(t, s.Title)
		require.NotZero(t, s.Confidence)
	}

	// 7. re-run — dedup, 0 new
	run2, err := c.RunSubscription(sub.ID)
	require.NoError(t, err)
	require.Equal(t, 0, run2.NewSignals)

	// 8. delete sub — cascades signals
	require.NoError(t, c.DeleteSubscription(sub.ID))
	list, err = c.ListSubscriptions()
	require.NoError(t, err)
	require.Empty(t, list)
}

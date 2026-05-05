//go:build integration

package agent

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestDebug_FullPipelineTrace runs plan → Brave → extract for one query and
// logs each stage's output. Diagnoses "why are we getting 0 signals?".
//
//	go test -tags=integration -count=1 -run TestDebug_FullPipelineTrace -v ./internal/agent
func TestDebug_FullPipelineTrace(t *testing.T) {
	if os.Getenv("BRAVE_API_KEY") == "" || os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("BRAVE_API_KEY and OPENAI_API_KEY required")
	}
	queries := []string{
		"Formula 1 Brazilian Grand Prix Interlagos 2026",
		"Lollapalooza São Paulo 2026 lineup and dates",
	}
	e := NewOpenAIExtractor(os.Getenv("OPENAI_API_KEY"))
	b := NewBraveClient(os.Getenv("BRAVE_API_KEY"))
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	for _, q := range queries {
		t.Logf("\n========== query: %s ==========", q)
		plan, err := e.PlanQuery(ctx, q, "event", "2026-05-04")
		if err != nil {
			t.Errorf("plan: %v", err)
			continue
		}
		buf, _ := json.MarshalIndent(plan, "", "  ")
		t.Logf("plan:\n%s", string(buf))

		ans, err := b.Answer(ctx, plan.WebQuestion)
		if err != nil {
			t.Errorf("brave: %v", err)
			continue
		}
		t.Logf("brave (%d chars):\n%s", len(ans.Text), ans.Text)

		cands, err := e.ExtractEvents(ctx, ExtractInput{
			Query:    q,
			TodayISO: "2026-05-04",
			Answer:   ans.Text,
			Plan:     plan,
		})
		if err != nil {
			t.Errorf("extract: %v", err)
			continue
		}
		t.Logf("extracted %d candidates:", len(cands))
		for i, c := range cands {
			t.Logf("  [%d] title=%q date=%v venue=%v city=%v gatePass=%v",
				i, c.Title, c.Date, c.Venue, c.City, plan.PassesGate(eventBlob(c)))
		}
	}
}

// TestDebug_PlanQuery_Outputs is a one-off introspection helper. Run it to
// see the exact plan the LLM generates for canonical queries — tunes the
// system prompt without burning a full Brave roundtrip.
//
//	go test -tags=integration -run TestDebug_PlanQuery_Outputs -v ./internal/agent
func TestDebug_PlanQuery_Outputs(t *testing.T) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		t.Skip("OPENAI_API_KEY missing")
	}
	e := NewOpenAIExtractor(key)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cases := []struct {
		query, kind string
	}{
		{"Coldplay tour São Paulo 2026", "event"},
		{"Formula 1 Brazilian Grand Prix Interlagos 2026", "event"},
		{"Lollapalooza São Paulo 2026 lineup and dates", "event"},
		{"Stablecoin legislation in Brazil", "news"},
	}
	for _, c := range cases {
		plan, err := e.PlanQuery(ctx, c.query, c.kind, "2026-05-04")
		if err != nil {
			t.Errorf("[%s] plan err: %v", c.query, err)
			continue
		}
		buf, _ := json.MarshalIndent(plan, "", "  ")
		t.Logf("\nquery=%q\n%s", c.query, string(buf))
	}
}

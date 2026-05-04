//go:build integration

package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func loadBraveFixture(t *testing.T) AnswerResult {
	t.Helper()
	path := filepath.Join("testdata", "brave_sample.json")
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("no brave fixture at %s — run brave integration test first", path)
	}
	var f struct {
		Parsed AnswerResult `json:"parsed"`
	}
	if err := json.Unmarshal(buf, &f); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if f.Parsed.Text == "" {
		t.Fatalf("fixture has empty text")
	}
	return f.Parsed
}

func TestExtractEvents_FromBraveFixture(t *testing.T) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		t.Skip("OPENAI_API_KEY missing")
	}

	answer := loadBraveFixture(t)

	e := NewOpenAIExtractor(key)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cands, err := e.ExtractEvents(ctx, ExtractInput{
		Query:    "Coldplay concerts 2026",
		TodayISO: "2026-05-04",
		Answer:   answer.Text,
	})
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	t.Logf("extracted %d candidates", len(cands))
	if len(cands) == 0 {
		t.Fatal("expected at least one candidate from a fixture full of concerts")
	}
	for i, c := range cands {
		date := ""
		if c.Date != nil {
			date = *c.Date
		}
		venue := ""
		if c.Venue != nil {
			venue = *c.Venue
		}
		t.Logf("  [%d] %s | %s | %s | conf=%.2f", i, c.Title, date, venue, c.Confidence)
		if c.Title == "" {
			t.Errorf("candidate %d missing title", i)
		}
	}
}

func TestExtractNews_Live(t *testing.T) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		t.Skip("OPENAI_API_KEY missing")
	}

	sampleAnswer := `On May 1, 2026, Vietnam's State Bank announced new draft regulations for cryptocurrency exchanges, requiring all platforms to register with the Ministry of Finance by Q3 2026. Industry response has been mixed: domestic exchanges welcomed clarity while critics worry about high compliance costs.`

	e := NewOpenAIExtractor(key)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	res, err := e.ExtractNews(ctx, ExtractInput{
		Query:          "Cryptocurrency regulation Vietnam",
		TodayISO:       "2026-05-04",
		Answer:         sampleAnswer,
		RollingSummary: "",
	})
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	t.Logf("items: %d", len(res.Items))
	t.Logf("updated_summary: %s", res.UpdatedSummary)
	if len(res.Items) == 0 {
		t.Fatal("expected at least one news item from this seeded answer")
	}
	if res.UpdatedSummary == "" {
		t.Fatal("expected updated_summary to be set")
	}
	for i, it := range res.Items {
		t.Logf("  [%d] new=%v | %s", i, it.IsNewDevelopment, it.Headline)
	}
}

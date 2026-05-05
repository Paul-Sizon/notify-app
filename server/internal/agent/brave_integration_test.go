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

func TestBraveSearch_Live(t *testing.T) {
	key := os.Getenv("BRAVE_SEARCH_API_KEY")
	if key == "" {
		t.Skip("BRAVE_SEARCH_API_KEY missing")
	}
	c := NewBraveClient(key)

	// capture raw response for fixture creation + debugging.
	var raw []byte
	c.SetRawLogger(func(_ string, body []byte) { raw = body })

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	res, err := c.Answer(ctx,
		"List upcoming Coldplay concerts in 2026. Include each concert's date, city, and venue.")
	if err != nil {
		t.Fatalf("answer: %v\nraw: %s", err, string(raw))
	}
	if res.Text == "" {
		t.Fatalf("empty text. raw: %s", string(raw))
	}
	if len(res.Text) < 50 {
		t.Fatalf("suspiciously short text: %q", res.Text)
	}
	if len(res.Citations) == 0 {
		t.Fatalf("expected at least one citation from Brave Search results")
	}
	t.Logf("text (%d chars): %s", len(res.Text), res.Text)
	t.Logf("citations: %d", len(res.Citations))

	// Save fixture for unit tests + extractor tests.
	if dir := os.Getenv("FIXTURE_DIR"); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
		path := filepath.Join(dir, "brave_sample.json")
		buf, _ := json.MarshalIndent(map[string]any{
			"raw":    json.RawMessage(raw),
			"parsed": res,
		}, "", "  ")
		_ = os.WriteFile(path, buf, 0o644)
		t.Logf("wrote fixture %s", path)
	}
}

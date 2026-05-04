package agent

import "testing"

func TestDedupeCitations_DropsEmptyAndDupes(t *testing.T) {
	in := []Citation{
		{URL: "https://a.com", Title: "A"},
		{URL: ""},
		{URL: "https://a.com", Title: "A2"},
		{URL: "https://b.com"},
	}
	out := dedupeCitations(in)
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d: %+v", len(out), out)
	}
	if out[0].URL != "https://a.com" || out[1].URL != "https://b.com" {
		t.Fatalf("wrong order: %+v", out)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("abcdef", 3); got != "abc...(truncated)" {
		t.Fatalf("got %q", got)
	}
	if got := truncate("ab", 3); got != "ab" {
		t.Fatalf("got %q", got)
	}
}

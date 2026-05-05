package agent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestBraveAnswer_ParsesSearchResponse(t *testing.T) {
	body := `{
		"web": {
			"results": [
				{"title": "Coldplay 2026", "url": "https://coldplay.com/tour", "description": "Tour dates.", "extra_snippets": ["Wembley July 10."], "page_age": "2026-04-12"},
				{"title": "Ticketmaster", "url": "https://ticketmaster.com/x", "description": "Buy tickets."}
			]
		},
		"news": {"results": [
			{"title": "News item", "url": "https://news.example/a", "description": "News snippet.", "age": "2 days ago"}
		]}
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Subscription-Token"); got != "test-key" {
			t.Errorf("missing/wrong subscription token: %q", got)
		}
		if r.URL.Query().Get("q") == "" {
			t.Errorf("missing q param")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := NewBraveClient("test-key")
	c.http = srv.Client()
	// Override URL by hitting test server through transport hack: easiest is
	// to swap braveSearchURL via a small indirection. Instead, just point the
	// client at the test server via a custom RoundTripper.
	c.http.Transport = rewriteHostTransport{target: srv.URL, base: srv.Client().Transport}

	res, err := c.Answer(context.Background(), "coldplay 2026")
	if err != nil {
		t.Fatalf("answer: %v", err)
	}
	if !strings.Contains(res.Text, "Coldplay 2026") {
		t.Errorf("text missing first result title: %q", res.Text)
	}
	if !strings.Contains(res.Text, "Wembley July 10") {
		t.Errorf("text missing extra_snippet: %q", res.Text)
	}
	if !strings.Contains(res.Text, "Published: 2026-04-12") {
		t.Errorf("text missing page_age: %q", res.Text)
	}
	if !strings.Contains(res.Text, "Age: 2 days ago") {
		t.Errorf("news result age missing: %q", res.Text)
	}
	if len(res.Citations) != 3 {
		t.Errorf("expected 3 citations (web + news), got %d", len(res.Citations))
	}
}

func TestBraveAnswer_HTTPErrorWrappedAsBraveError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer srv.Close()

	c := NewBraveClient("test-key")
	c.http.Transport = rewriteHostTransport{target: srv.URL, base: srv.Client().Transport}

	_, err := c.Answer(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error")
	}
	var bErr *BraveError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected *BraveError, got %T: %v", err, err)
	}
	if bErr.StatusCode != 429 {
		t.Errorf("status: got %d", bErr.StatusCode)
	}
}

// rewriteHostTransport rewrites every outbound request's URL to point at
// the httptest server while preserving path + query, so we don't have to
// expose braveSearchURL as a mutable var.
type rewriteHostTransport struct {
	target string
	base   http.RoundTripper
}

func (rt rewriteHostTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.URL.Scheme = "http"
	r.URL.Host = strings.TrimPrefix(strings.TrimPrefix(rt.target, "http://"), "https://")
	if rt.base == nil {
		return http.DefaultTransport.RoundTrip(r)
	}
	return rt.base.RoundTrip(r)
}


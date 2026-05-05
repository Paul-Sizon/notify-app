package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	braveSearchURL    = "https://api.search.brave.com/res/v1/web/search"
	braveDefaultCount = 20
	braveDefaultCountry = "US"
)

// BraveError carries the HTTP status from a Brave API failure so callers
// can classify it (402 = monthly cap, 429 = rate limit, 5xx = transient).
// Implements the error interface; pair with errors.As to unwrap.
type BraveError struct {
	StatusCode int
	Body       string
}

func (e *BraveError) Error() string {
	return fmt.Sprintf("brave http %d: %s", e.StatusCode, truncate(e.Body, 500))
}

type BraveClient struct {
	key     string
	count   int
	country string
	http    *http.Client
	rawLog  func(string, []byte)
}

func NewBraveClient(key string) *BraveClient {
	return &BraveClient{
		key:     key,
		count:   braveDefaultCount,
		country: braveDefaultCountry,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

// SetRawLogger lets callers inspect the raw response body (for tests).
func (b *BraveClient) SetRawLogger(f func(label string, body []byte)) { b.rawLog = f }

// braveSearchResp is the subset of the Web Search API response we consume.
// Full schema: https://api-dashboard.search.brave.com/app/documentation/web-search
type braveSearchResp struct {
	Web struct {
		Results []braveResult `json:"results"`
	} `json:"web"`
	News struct {
		Results []braveResult `json:"results"`
	} `json:"news"`
}

type braveResult struct {
	Title         string   `json:"title"`
	URL           string   `json:"url"`
	Description   string   `json:"description"`
	ExtraSnippets []string `json:"extra_snippets"`
	Age           string   `json:"age"`
	PageAge       string   `json:"page_age"`
}

func (b *BraveClient) Answer(ctx context.Context, query string) (AnswerResult, error) {
	q := url.Values{}
	q.Set("q", query)
	q.Set("count", strconv.Itoa(b.count))
	q.Set("country", b.country)
	q.Set("extra_snippets", "true")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, braveSearchURL+"?"+q.Encode(), nil)
	if err != nil {
		return AnswerResult{}, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", b.key)

	resp, err := b.http.Do(req)
	if err != nil {
		return AnswerResult{}, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return AnswerResult{}, fmt.Errorf("read: %w", err)
	}
	if b.rawLog != nil {
		b.rawLog("brave_response", raw)
	}
	if resp.StatusCode >= 400 {
		return AnswerResult{}, &BraveError{StatusCode: resp.StatusCode, Body: string(raw)}
	}

	var br braveSearchResp
	if err := json.Unmarshal(raw, &br); err != nil {
		return AnswerResult{}, fmt.Errorf("decode: %w (body: %s)", err, truncate(string(raw), 500))
	}

	results := append([]braveResult{}, br.Web.Results...)
	results = append(results, br.News.Results...)

	out := AnswerResult{
		Text:      formatResultsAsText(results),
		Citations: resultsToCitations(results),
	}
	return out, nil
}

// formatResultsAsText turns a list of search results into a single block of
// text the downstream extractor can read. Format is a numbered list where
// each entry has title, URL, age (when available), and the description plus
// any extra snippets joined together. The extractor is an LLM that consumes
// this text — keep it dense but parseable.
func formatResultsAsText(rs []braveResult) string {
	if len(rs) == 0 {
		return ""
	}
	var b strings.Builder
	for i, r := range rs {
		fmt.Fprintf(&b, "[%d] %s\n", i+1, strings.TrimSpace(r.Title))
		if r.URL != "" {
			fmt.Fprintf(&b, "URL: %s\n", r.URL)
		}
		if r.PageAge != "" {
			fmt.Fprintf(&b, "Published: %s\n", r.PageAge)
		} else if r.Age != "" {
			fmt.Fprintf(&b, "Age: %s\n", r.Age)
		}
		body := strings.TrimSpace(r.Description)
		for _, s := range r.ExtraSnippets {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if body != "" {
				body += " "
			}
			body += s
		}
		if body != "" {
			b.WriteString(body)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func resultsToCitations(rs []braveResult) []Citation {
	out := make([]Citation, 0, len(rs))
	for _, r := range rs {
		if r.URL == "" {
			continue
		}
		out = append(out, Citation{URL: r.URL, Title: r.Title})
	}
	return dedupeCitations(out)
}

func dedupeCitations(cs []Citation) []Citation {
	seen := make(map[string]bool, len(cs))
	out := make([]Citation, 0, len(cs))
	for _, c := range cs {
		if c.URL == "" || seen[c.URL] {
			continue
		}
		seen[c.URL] = true
		out = append(out, c)
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}

package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const braveChatURL = "https://api.search.brave.com/res/v1/chat/completions"

type BraveClient struct {
	key    string
	model  string
	http   *http.Client
	rawLog func(string, []byte)
}

func NewBraveClient(key string) *BraveClient {
	return &BraveClient{
		key:   key,
		model: "brave",
		http:  &http.Client{Timeout: 60 * time.Second},
	}
}

// SetRawLogger lets callers inspect the raw response body (for tests).
func (b *BraveClient) SetRawLogger(f func(label string, body []byte)) { b.rawLog = f }

type braveMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type braveRequest struct {
	Model            string             `json:"model"`
	Messages         []braveMessage     `json:"messages"`
	Stream           bool               `json:"stream"`
	WebSearchOptions *braveSearchOpts   `json:"web_search_options,omitempty"`
}

type braveSearchOpts struct {
	EnableCitations bool   `json:"enable_citations,omitempty"`
	EnableEntities  bool   `json:"enable_entities,omitempty"`
	EnableResearch  bool   `json:"enable_research,omitempty"`
	Country         string `json:"country,omitempty"`
}

// Best-effort response shape. Brave is OpenAI-compatible at the top level.
// Citation field name may vary; we capture both common variants and dedupe.
type braveResponse struct {
	Choices []struct {
		Message struct {
			Content   string     `json:"content"`
			Citations []Citation `json:"citations"`
		} `json:"message"`
	} `json:"choices"`
	Citations []Citation `json:"citations"`
}

func (b *BraveClient) Answer(ctx context.Context, query string) (AnswerResult, error) {
	body, err := json.Marshal(braveRequest{
		Model: b.model,
		Messages: []braveMessage{
			{Role: "user", Content: query},
		},
		Stream: false,
		WebSearchOptions: &braveSearchOpts{
			EnableCitations: true,
		},
	})
	if err != nil {
		return AnswerResult{}, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, braveChatURL, bytes.NewReader(body))
	if err != nil {
		return AnswerResult{}, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-subscription-token", b.key)

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
		return AnswerResult{}, fmt.Errorf("brave http %d: %s", resp.StatusCode, truncate(string(raw), 500))
	}

	var br braveResponse
	if err := json.Unmarshal(raw, &br); err != nil {
		return AnswerResult{}, fmt.Errorf("decode: %w (body: %s)", err, truncate(string(raw), 500))
	}

	out := AnswerResult{}
	if len(br.Choices) > 0 {
		out.Text = br.Choices[0].Message.Content
		out.Citations = append(out.Citations, br.Choices[0].Message.Citations...)
	}
	out.Citations = append(out.Citations, br.Citations...)
	out.Citations = dedupeCitations(out.Citations)
	return out, nil
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

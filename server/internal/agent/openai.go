package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type OpenAIExtractor struct {
	c     *openai.Client
	model string
}

func NewOpenAIExtractor(key string) *OpenAIExtractor {
	return &OpenAIExtractor{
		c:     openai.NewClient(key),
		model: openai.GPT4oMini,
	}
}

func (e *OpenAIExtractor) ExtractEvents(ctx context.Context, in ExtractInput) ([]EventCandidate, error) {
	user := fmt.Sprintf(
		"Query: %s\nToday: %s\n%s\nWeb-grounded answer to extract events from:\n\"\"\"\n%s\n\"\"\"\n\nReturn JSON. Hard rules:\n- Only events with a concrete date (or first day of date range) in the future relative to today.\n- Only events that match the query intent. Apply the plan rules above strictly.\n- Reject tribute bands, cover acts, fan events, and lookalike acts unless the query explicitly asks for them.\n- Reject events whose primary act is not the subject of the query (e.g. random venue listings that share a keyword).\n- If the same event appears twice, emit once.\n- Empty array is the correct answer when nothing qualifies — do not pad results to look useful.\n- URL is optional; emit a URL only if the answer text actually contains one for that event.",
		in.Query, in.TodayISO, formatPlanForPrompt(in.Plan), in.Answer)

	schema := &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"events": {
				Type: jsonschema.Array,
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"title":      {Type: jsonschema.String},
						"date":       {Type: jsonschema.String, Description: "ISO 8601 date or empty string"},
						"venue":      {Type: jsonschema.String, Description: "venue name or empty string"},
						"city":       {Type: jsonschema.String, Description: "city or empty string"},
						"url":        {Type: jsonschema.String, Description: "URL or empty string"},
						"confidence": {Type: jsonschema.Number},
					},
					Required:             []string{"title", "date", "venue", "city", "url", "confidence"},
					AdditionalProperties: false,
				},
			},
		},
		Required:             []string{"events"},
		AdditionalProperties: false,
	}

	resp, err := e.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: e.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: eventSystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "events_response",
				Schema: schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai: no choices")
	}

	var out struct {
		Events []struct {
			Title      string  `json:"title"`
			Date       string  `json:"date"`
			Venue      string  `json:"venue"`
			City       string  `json:"city"`
			URL        string  `json:"url"`
			Confidence float64 `json:"confidence"`
		} `json:"events"`
	}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &out); err != nil {
		return nil, fmt.Errorf("decode events: %w (body: %s)", err, resp.Choices[0].Message.Content)
	}
	cands := make([]EventCandidate, 0, len(out.Events))
	for _, e := range out.Events {
		cands = append(cands, EventCandidate{
			Title:      e.Title,
			Date:       strPtr(e.Date),
			Venue:      strPtr(e.Venue),
			City:       strPtr(e.City),
			URL:        strPtr(e.URL),
			Confidence: e.Confidence,
		})
	}
	return cands, nil
}

func (e *OpenAIExtractor) ExtractNews(ctx context.Context, in ExtractInput) (NewsExtraction, error) {
	user := fmt.Sprintf(
		"Topic: %s\nToday: %s\n%s\nWhat has already been reported (do not repeat substantively similar items):\n\"\"\"\n%s\n\"\"\"\n\nWeb-grounded answer to extract news from:\n\"\"\"\n%s\n\"\"\"\n\nReturn JSON. Hard rules:\n- is_new_development must be false for any item already covered above.\n- Discard opinion pieces, speculation, listicles, and SEO content.\n- Apply the plan rules above strictly — drop items that don't match the query subject or that hit a reject term.\n- updated_summary must be 3-5 sentences capturing all material developments now known. Rewrite, do not append.",
		in.Query, in.TodayISO, formatPlanForPrompt(in.Plan), in.RollingSummary, in.Answer)

	schema := &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"items": {
				Type: jsonschema.Array,
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"headline":           {Type: jsonschema.String},
						"canonical_headline": {Type: jsonschema.String},
						"summary":            {Type: jsonschema.String},
						"url":                {Type: jsonschema.String},
						"published_at":       {Type: jsonschema.String, Description: "ISO 8601 or empty string"},
						"confidence":         {Type: jsonschema.Number},
						"is_new_development": {Type: jsonschema.Boolean},
					},
					Required:             []string{"headline", "canonical_headline", "summary", "url", "published_at", "confidence", "is_new_development"},
					AdditionalProperties: false,
				},
			},
			"updated_summary": {Type: jsonschema.String},
		},
		Required:             []string{"items", "updated_summary"},
		AdditionalProperties: false,
	}

	resp, err := e.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: e.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: newsSystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "news_response",
				Schema: schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		return NewsExtraction{}, fmt.Errorf("openai: %w", err)
	}
	if len(resp.Choices) == 0 {
		return NewsExtraction{}, fmt.Errorf("openai: no choices")
	}

	var out struct {
		Items []struct {
			Headline          string  `json:"headline"`
			CanonicalHeadline string  `json:"canonical_headline"`
			Summary           string  `json:"summary"`
			URL               string  `json:"url"`
			PublishedAt       string  `json:"published_at"`
			Confidence        float64 `json:"confidence"`
			IsNewDevelopment  bool    `json:"is_new_development"`
		} `json:"items"`
		UpdatedSummary string `json:"updated_summary"`
	}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &out); err != nil {
		return NewsExtraction{}, fmt.Errorf("decode news: %w (body: %s)", err, resp.Choices[0].Message.Content)
	}
	items := make([]NewsCandidate, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, NewsCandidate{
			Headline:          it.Headline,
			CanonicalHeadline: it.CanonicalHeadline,
			Summary:           it.Summary,
			URL:               strPtr(it.URL),
			PublishedAt:       strPtr(it.PublishedAt),
			Confidence:        it.Confidence,
			IsNewDevelopment:  it.IsNewDevelopment,
		})
	}
	return NewsExtraction{Items: items, UpdatedSummary: out.UpdatedSummary}, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

const eventSystemPrompt = `You extract upcoming events from web-grounded answer text. You are precise. You discard speculation, retrospectives, tribute/cover acts, and irrelevant content. You never invent details. If the text says no events are known, return an empty array.

Granularity rules:
- One emitted event per real-world event. For multi-day or multi-session events (race weekends, festivals, conferences, concert tours of a single show), emit ONE event with the headline date (race day / opening day / main set), not one per session/practice/heat.
- The title MUST reference the primary subject of the query (artist, race series, festival, regulator). Use the canonical full name, not a generic sub-label. Example: "Formula 1 Brazilian Grand Prix", not "Grand Prix" or "Race".
- A tour with multiple dates IS multiple events — one per show date.`

const newsSystemPrompt = `You monitor news for material new developments on a topic. You distinguish genuinely new facts from commentary, retrospectives, and rephrased coverage. You prefer primary sources. You never invent details.`

const plannerSystemPrompt = `You are a query planner for a precision-focused web watcher. The user types a short, often ambiguous query. You convert it into:

1. web_question — a precise, source-aware question for the web grounder. Explicitly request primary/official sources (artist channels, ticketing platforms, regulators, named outlets, governing bodies) and call out the user's intent.

2. must_match — at MOST two entries. Each entry must be a substring (case-insensitive) that the downstream extractor would naturally include in a candidate's title, venue, city, or url. Every entry is required (logical AND). Inside one entry, use "|" to list spelling alternatives (logical OR within the entry).

   STRICT RULES for must_match:
   - Cap the list at 1–2 entries total. More entries = pipeline rejects valid signals.
   - Entry 1: the named entity (the artist, festival, race series, regulator, company). Include common spelling variants. Example: "coldplay" or "f1|formula 1|formula one".
   - Entry 2 (optional): the region IF the user specified one. Example: "são paulo|sao paulo|brasil|brazil". If the user didn't specify a region, omit this.
   - NEVER include: years (already filtered by date), generic verbs ("tour", "concert", "lineup", "dates", "schedule", "news"), the user's exact query repeated.
   - Empty list is acceptable if the query is too generic to constrain (e.g. "AI conferences").

3. reject — case-insensitive substrings that disqualify a candidate. Common values for entertainment queries: "tribute", "cover band", "experience", "lookalike", "fan event". Leave empty for non-entertainment queries unless a specific noise pattern applies.

4. notes — 1–2 sentences of extraction guidance (what counts as on-target vs noise).

You bias toward precision over recall, but a too-strict must_match silently kills the pipeline. When in doubt, drop the second entry.`

// formatPlanForPrompt renders the plan as a short instructions block. Empty
// plans render as the empty string so prompts stay clean for stub/test use.
func formatPlanForPrompt(p QueryPlan) string {
	if len(p.MustMatch) == 0 && len(p.Reject) == 0 && p.Notes == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("Plan rules (apply strictly):\n")
	if len(p.MustMatch) > 0 {
		b.WriteString("- Must match (each entry required; alternatives separated by '|'): ")
		b.WriteString(strings.Join(p.MustMatch, "; "))
		b.WriteString("\n")
	}
	if len(p.Reject) > 0 {
		b.WriteString("- Reject if candidate contains any of: ")
		b.WriteString(strings.Join(p.Reject, ", "))
		b.WriteString("\n")
	}
	if p.Notes != "" {
		b.WriteString("- Notes: ")
		b.WriteString(p.Notes)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

func (e *OpenAIExtractor) PlanQuery(ctx context.Context, query, kind, todayISO string) (QueryPlan, error) {
	user := fmt.Sprintf("User query: %s\nWatcher type: %s\nToday: %s\n\nReturn the plan as JSON.", query, kind, todayISO)

	schema := &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"web_question": {Type: jsonschema.String, Description: "precise web question for the grounder; cite primary-source preference"},
			"must_match": {
				Type:        jsonschema.Array,
				Description: "each entry required (case-insensitive substring); use '|' inside an entry for spelling alternatives",
				Items:       &jsonschema.Definition{Type: jsonschema.String},
			},
			"reject": {
				Type:        jsonschema.Array,
				Description: "case-insensitive substrings that disqualify a candidate",
				Items:       &jsonschema.Definition{Type: jsonschema.String},
			},
			"notes": {Type: jsonschema.String, Description: "1-2 sentences of extraction guidance"},
		},
		Required:             []string{"web_question", "must_match", "reject", "notes"},
		AdditionalProperties: false,
	}

	resp, err := e.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: e.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: plannerSystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "query_plan",
				Schema: schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		return QueryPlan{}, fmt.Errorf("openai plan: %w", err)
	}
	if len(resp.Choices) == 0 {
		return QueryPlan{}, fmt.Errorf("openai plan: no choices")
	}

	var out QueryPlan
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &out); err != nil {
		return QueryPlan{}, fmt.Errorf("decode plan: %w (body: %s)", err, resp.Choices[0].Message.Content)
	}
	if strings.TrimSpace(out.WebQuestion) == "" {
		// Defense in depth: never let the planner leave Brave with an empty prompt.
		out.WebQuestion = query
	}
	return out, nil
}

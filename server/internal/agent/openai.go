package agent

import (
	"context"
	"encoding/json"
	"fmt"

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
		"Query: %s\nToday: %s\n\nWeb-grounded answer to extract events from:\n\"\"\"\n%s\n\"\"\"\n\nReturn JSON. Hard rules:\n- Only events with a concrete date (or first day of date range) in the future relative to today.\n- Only events that match the query intent.\n- If the same event appears twice, emit once.\n- Empty array is the correct answer when nothing qualifies.\n- URL is optional; emit a URL only if the answer text actually contains one for that event.",
		in.Query, in.TodayISO, in.Answer)

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
		"Topic: %s\nToday: %s\n\nWhat has already been reported (do not repeat substantively similar items):\n\"\"\"\n%s\n\"\"\"\n\nWeb-grounded answer to extract news from:\n\"\"\"\n%s\n\"\"\"\n\nReturn JSON. Hard rules:\n- is_new_development must be false for any item already covered above.\n- Discard opinion pieces and speculation.\n- updated_summary must be 3-5 sentences capturing all material developments now known. Rewrite, do not append.",
		in.Query, in.TodayISO, in.RollingSummary, in.Answer)

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

const eventSystemPrompt = `You extract upcoming events from web-grounded answer text. You are precise. You discard speculation, retrospectives, and irrelevant content. You never invent details. If the text says no events are known, return an empty array.`

const newsSystemPrompt = `You monitor news for material new developments on a topic. You distinguish genuinely new facts from commentary, retrospectives, and rephrased coverage. You prefer primary sources. You never invent details.`

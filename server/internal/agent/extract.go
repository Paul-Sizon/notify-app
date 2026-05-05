package agent

import "context"

type ExtractInput struct {
	Query          string
	TodayISO       string
	Answer         string
	RollingSummary string // news only
	Plan           QueryPlan
}

type EventCandidate struct {
	Title      string  `json:"title"`
	Date       *string `json:"date"`
	Venue      *string `json:"venue"`
	City       *string `json:"city"`
	URL        *string `json:"url"`
	Confidence float64 `json:"confidence"`
	// Summary is a 2-3 sentence GPT-written prose description of the event:
	// what is happening, when, where, why notable. Surfaced in the signal
	// detail UI so the user gets a readable explanation instead of a bare
	// "date · venue · city" line.
	Summary string `json:"summary"`
}

type NewsCandidate struct {
	Headline          string  `json:"headline"`
	CanonicalHeadline string  `json:"canonical_headline"`
	Summary           string  `json:"summary"`
	URL               *string `json:"url"`
	PublishedAt       *string `json:"published_at"`
	Confidence        float64 `json:"confidence"`
	IsNewDevelopment  bool    `json:"is_new_development"`
}

type NewsExtraction struct {
	Items          []NewsCandidate `json:"items"`
	UpdatedSummary string          `json:"updated_summary"`
}

type Extractor interface {
	ExtractEvents(ctx context.Context, in ExtractInput) ([]EventCandidate, error)
	ExtractNews(ctx context.Context, in ExtractInput) (NewsExtraction, error)
}

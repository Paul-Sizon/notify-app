package agent

import "context"

type Citation struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

type AnswerResult struct {
	Text      string     `json:"text"`
	Citations []Citation `json:"citations"`
}

type Searcher interface {
	Answer(ctx context.Context, query string) (AnswerResult, error)
}

package agent

import (
	"context"
	"strings"
)

// QueryPlan is the output of stage A. The agent pipeline runs:
//
//	planner → web search (Brave) → extractor (OpenAI) → Go-side gate
//
// The plan is the contract between stages. WebQuestion is what we actually
// send to Brave (not the raw user query, which tends to be terse and
// ambiguous). MustMatch and Reject are enforced both in the extractor's
// prompt AND deterministically in Go (PassesGate) — belt and suspenders,
// because LLMs occasionally drop a rule.
type QueryPlan struct {
	// WebQuestion is the precise, source-aware question for the web grounder.
	WebQuestion string `json:"web_question"`

	// MustMatch: every entry must appear (case-insensitive substring) in at
	// least one of {title, body/venue/city, url}. An entry may itself be a
	// pipe-separated set of alternatives, e.g. "são paulo|brasil|brazil" —
	// in which case any one alternative satisfies that entry.
	MustMatch []string `json:"must_match"`

	// Reject: any candidate whose blob contains any of these (ci) is dropped.
	Reject []string `json:"reject"`

	// Notes is free-text guidance forwarded into the extractor prompt.
	Notes string `json:"notes"`
}

// Planner is stage A of the pipeline.
type Planner interface {
	PlanQuery(ctx context.Context, query, kind, todayISO string) (QueryPlan, error)
}

// PassesGate runs the deterministic post-LLM filter. blob is the lowercased
// concatenation of every text field on the candidate that we want to gate
// against (title, body, venue, city, url, etc.).
func (p QueryPlan) PassesGate(blob string) bool {
	blob = strings.ToLower(blob)
	for _, r := range p.Reject {
		r = strings.ToLower(strings.TrimSpace(r))
		if r == "" {
			continue
		}
		if strings.Contains(blob, r) {
			return false
		}
	}
	for _, m := range p.MustMatch {
		if !anyAltMatches(blob, m) {
			return false
		}
	}
	return true
}

// anyAltMatches treats `entry` as a `|`-separated set of alternatives;
// returns true if any alternative is a substring of blob.
func anyAltMatches(blob, entry string) bool {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return true
	}
	for alt := range strings.SplitSeq(entry, "|") {
		alt = strings.ToLower(strings.TrimSpace(alt))
		if alt == "" {
			continue
		}
		if strings.Contains(blob, alt) {
			return true
		}
	}
	return false
}

// PassthroughPlan is the no-op plan used when Deps.Planner is nil — the raw
// user query goes to Brave and the gate accepts everything. Used by tests
// that stub out the LLM stack and don't care about the planner stage.
func PassthroughPlan(query string) QueryPlan {
	return QueryPlan{WebQuestion: query}
}

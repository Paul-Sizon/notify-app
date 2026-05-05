package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/paulsizon/notify/server/internal/db"
	"github.com/paulsizon/notify/server/internal/push"
)

type Deps struct {
	DB        *db.DB
	Searcher  Searcher
	Planner   Planner // optional; nil falls back to a passthrough plan (raw query, no gate)
	Extractor Extractor
	Pusher    push.Pusher
	Now       func() time.Time
}

// RunSubscription executes the full agent pipeline for a single subscription.
// Returns IDs of newly inserted signals (for downstream push delivery).
//
// Scheduling contract: this function ALWAYS advances the subscription's
// next_run_at, on both success and failure. Success: by cadence. Failure:
// by BackoffDelay based on the error kind. This is what stops the
// 10-second retry storm — a failing sub gets pushed out of the "due" set
// for at least minBackoff, instead of being instantly re-due.
func RunSubscription(ctx context.Context, deps Deps, subID uuid.UUID) (newIDs []uuid.UUID, retErr error) {
	if deps.Now == nil {
		deps.Now = time.Now
	}

	sub, err := deps.DB.GetSubscription(ctx, subID)
	if err != nil {
		return nil, fmt.Errorf("load subscription: %w", err)
	}

	now := deps.Now().UTC()
	todayISO := now.Format("2006-01-02")

	// Track the success-path summary so the deferred reschedule can persist
	// it. Empty string = "no summary update" (events runs).
	var updatedSummary string

	// One reschedule path for both outcomes. Use a detached context with a
	// short timeout so we still record state even if the caller's ctx was
	// just canceled (e.g. server shutting down). Without this, a cancellation
	// mid-run leaves next_run_at in the past = retry storm on next start.
	defer func() {
		persistCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if retErr != nil {
			kind := ClassifyError(retErr)
			cadence := time.Duration(sub.CadenceSeconds) * time.Second
			delay := BackoffDelay(kind, cadence, sub.FailureCount)
			_ = deps.DB.BackoffSubscription(persistCtx, sub.ID, now, delay, string(kind))
			return
		}
		_ = deps.DB.RescheduleSubscription(persistCtx, sub.ID, now, updatedSummary)
	}()

	// Stage A — query plan. Without a planner, we send the raw user query
	// to Brave; this matches pre-planner behavior (used by stub tests).
	plan := PassthroughPlan(sub.Query)
	if deps.Planner != nil {
		p, err := deps.Planner.PlanQuery(ctx, sub.Query, sub.Type, todayISO)
		if err != nil {
			return nil, fmt.Errorf("plan: %w", err)
		}
		plan = p
	}

	// Stage B — web search using the planned question, not the raw query.
	answer, err := deps.Searcher.Answer(ctx, plan.WebQuestion)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	in := ExtractInput{
		Query:          sub.Query,
		TodayISO:       todayISO,
		Answer:         answer.Text,
		RollingSummary: sub.RollingSummary,
		Plan:           plan,
	}

	var toInsert []db.SignalInsert
	switch sub.Type {
	case "event":
		cands, err := deps.Extractor.ExtractEvents(ctx, in)
		if err != nil {
			return nil, fmt.Errorf("extract events: %w", err)
		}
		toInsert = eventsToSignals(sub.ID, cands, now, plan)
	case "news":
		res, err := deps.Extractor.ExtractNews(ctx, in)
		if err != nil {
			return nil, fmt.Errorf("extract news: %w", err)
		}
		toInsert = newsToSignals(sub.ID, res.Items, plan)
		updatedSummary = res.UpdatedSummary
	default:
		return nil, fmt.Errorf("unknown subscription type: %s", sub.Type)
	}

	newIDs, err = deps.DB.InsertSignals(ctx, toInsert)
	if err != nil {
		return nil, fmt.Errorf("insert signals: %w", err)
	}
	// Reschedule is handled by the deferred function above (success path
	// resets failure_count and bumps next_run_at by cadence).

	if deps.Pusher != nil && len(newIDs) > 0 {
		dev, err := deps.DB.GetDevice(ctx, sub.DeviceID)
		if err == nil && dev.APNsToken != "" {
			pushNewSignals(ctx, deps, sub, dev.APNsToken, newIDs, toInsert)
		}
	}

	return newIDs, nil
}

func pushNewSignals(ctx context.Context, deps Deps, sub db.Subscription, token string, newIDs []uuid.UUID, inserted []db.SignalInsert) {
	// Build a fingerprint→insert map so we can look up titles for the IDs we got back.
	byFP := make(map[string]db.SignalInsert, len(inserted))
	for _, s := range inserted {
		byFP[s.Fingerprint] = s
	}
	for _, id := range newIDs {
		// We don't carry FP back from InsertSignals, so re-load from DB.
		// Fast: indexed PK lookup. Use ListSignalsBySubscription with a tight window if needed.
		// For MVP: load via raw query.
		var fp, title, body string
		var url *string
		err := deps.DB.Pool.QueryRow(ctx, `SELECT fingerprint, title, COALESCE(body,''), url FROM signals WHERE id = $1`, id).Scan(&fp, &title, &body, &url)
		if err != nil {
			continue
		}
		urlStr := ""
		if url != nil {
			urlStr = *url
		}
		n := push.Notification{
			APNsToken:      token,
			SubscriptionID: sub.ID,
			SignalID:       id,
			Title:          sub.Query,
			Body:           title,
			URL:            urlStr,
		}
		if err := deps.Pusher.Send(ctx, n); err != nil {
			continue
		}
		_ = deps.DB.MarkSignalNotified(ctx, id, deps.Now())
		_ = body // reserved for richer payload later
	}
}

func eventsToSignals(subID uuid.UUID, cands []EventCandidate, now time.Time, plan QueryPlan) []db.SignalInsert {
	today := now.UTC().Truncate(24 * time.Hour)
	out := make([]db.SignalInsert, 0, len(cands))
	for _, c := range cands {
		date := ""
		if c.Date != nil {
			date = *c.Date
		}
		venue := ""
		if c.Venue != nil {
			venue = *c.Venue
		}
		fp := EventFingerprint(c.Title, date, venue)

		var occurs *time.Time
		if c.Date != nil {
			if t, err := time.Parse("2006-01-02", *c.Date); err == nil {
				occurs = &t
			}
		}
		// Defense in depth: drop past events even if extractor missed the rule.
		if occurs != nil && occurs.Before(today) {
			continue
		}
		// Plan gate (deterministic; runs after the LLM filter).
		if !plan.PassesGate(eventBlob(c)) {
			continue
		}

		// GPT-written prose summary is the preferred body — explains what,
		// when, where, why notable. Fall back to "date · venue · city" for
		// stub/test paths and any extractor that doesn't fill summary.
		body := strings.TrimSpace(c.Summary)
		if body == "" {
			bodyParts := []string{}
			if c.Date != nil {
				bodyParts = append(bodyParts, *c.Date)
			}
			if c.Venue != nil {
				bodyParts = append(bodyParts, *c.Venue)
			}
			if c.City != nil {
				bodyParts = append(bodyParts, *c.City)
			}
			body = strings.Join(bodyParts, " · ")
		}

		domains := domainsFromURL(c.URL)
		payload, _ := json.Marshal(c)

		out = append(out, db.SignalInsert{
			SubscriptionID: subID,
			Fingerprint:    fp,
			Title:          c.Title,
			Body:           strPtr(body),
			URL:            c.URL,
			OccursAt:       occurs,
			SourceDomains:  domains,
			Confidence:     float32(c.Confidence),
			Payload:        payload,
		})
	}
	return out
}

func newsToSignals(subID uuid.UUID, items []NewsCandidate, plan QueryPlan) []db.SignalInsert {
	out := make([]db.SignalInsert, 0, len(items))
	for _, it := range items {
		if !it.IsNewDevelopment {
			continue
		}
		if !plan.PassesGate(newsBlob(it)) {
			continue
		}
		fp := NewsFingerprint(it.CanonicalHeadline)

		var pub *time.Time
		if it.PublishedAt != nil {
			if t, err := time.Parse(time.RFC3339, *it.PublishedAt); err == nil {
				pub = &t
			} else if t, err := time.Parse("2006-01-02", *it.PublishedAt); err == nil {
				pub = &t
			}
		}

		domains := domainsFromURL(it.URL)
		payload, _ := json.Marshal(it)
		body := it.Summary

		out = append(out, db.SignalInsert{
			SubscriptionID: subID,
			Fingerprint:    fp,
			Title:          it.Headline,
			Body:           &body,
			URL:            it.URL,
			OccursAt:       pub,
			SourceDomains:  domains,
			Confidence:     float32(it.Confidence),
			Payload:        payload,
		})
	}
	return out
}

// eventBlob assembles every text field that the gate inspects for an event
// candidate, joined by spaces. Pointer dereferences are guarded.
func eventBlob(c EventCandidate) string {
	parts := []string{c.Title}
	for _, p := range []*string{c.Venue, c.City, c.URL} {
		if p != nil {
			parts = append(parts, *p)
		}
	}
	return strings.Join(parts, " ")
}

// newsBlob is the news-side analogue of eventBlob.
func newsBlob(it NewsCandidate) string {
	parts := []string{it.Headline, it.CanonicalHeadline, it.Summary}
	if it.URL != nil {
		parts = append(parts, *it.URL)
	}
	return strings.Join(parts, " ")
}

func domainsFromURL(u *string) []string {
	if u == nil || *u == "" {
		return []string{}
	}
	parsed, err := url.Parse(*u)
	if err != nil || parsed.Host == "" {
		return []string{}
	}
	return []string{strings.TrimPrefix(parsed.Host, "www.")}
}

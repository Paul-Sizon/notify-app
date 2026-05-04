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
	Extractor Extractor
	Pusher    push.Pusher
	Now       func() time.Time
}

// RunSubscription executes the full agent pipeline for a single subscription.
// Returns IDs of newly inserted signals (for downstream push delivery).
func RunSubscription(ctx context.Context, deps Deps, subID uuid.UUID) ([]uuid.UUID, error) {
	if deps.Now == nil {
		deps.Now = time.Now
	}

	sub, err := deps.DB.GetSubscription(ctx, subID)
	if err != nil {
		return nil, fmt.Errorf("load subscription: %w", err)
	}

	answer, err := deps.Searcher.Answer(ctx, sub.Query)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	now := deps.Now().UTC()
	in := ExtractInput{
		Query:          sub.Query,
		TodayISO:       now.Format("2006-01-02"),
		Answer:         answer.Text,
		RollingSummary: sub.RollingSummary,
	}

	var (
		toInsert       []db.SignalInsert
		updatedSummary string
	)
	switch sub.Type {
	case "event":
		cands, err := deps.Extractor.ExtractEvents(ctx, in)
		if err != nil {
			return nil, fmt.Errorf("extract events: %w", err)
		}
		toInsert = eventsToSignals(sub.ID, cands)
	case "news":
		res, err := deps.Extractor.ExtractNews(ctx, in)
		if err != nil {
			return nil, fmt.Errorf("extract news: %w", err)
		}
		toInsert = newsToSignals(sub.ID, res.Items)
		updatedSummary = res.UpdatedSummary
	default:
		return nil, fmt.Errorf("unknown subscription type: %s", sub.Type)
	}

	newIDs, err := deps.DB.InsertSignals(ctx, toInsert)
	if err != nil {
		return nil, fmt.Errorf("insert signals: %w", err)
	}

	if err := deps.DB.RescheduleSubscription(ctx, sub.ID, now, updatedSummary); err != nil {
		return nil, fmt.Errorf("reschedule: %w", err)
	}

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

func eventsToSignals(subID uuid.UUID, cands []EventCandidate) []db.SignalInsert {
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
		body := strings.Join(bodyParts, " · ")

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

func newsToSignals(subID uuid.UUID, items []NewsCandidate) []db.SignalInsert {
	out := make([]db.SignalInsert, 0, len(items))
	for _, it := range items {
		if !it.IsNewDevelopment {
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

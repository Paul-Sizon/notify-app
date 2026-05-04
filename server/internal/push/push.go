package push

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type Notification struct {
	APNsToken      string
	SubscriptionID uuid.UUID
	SignalID       uuid.UUID
	Title          string // shown as alert.title
	Body           string // shown as alert.body
	URL            string
}

type Pusher interface {
	Send(ctx context.Context, n Notification) error
}

// LogPusher prints notifications to slog. Use for local dev / Day 1 demo until
// the real APNs key is wired up.
type LogPusher struct {
	Logger *slog.Logger
}

func (p *LogPusher) Send(ctx context.Context, n Notification) error {
	logger := p.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.Info("push (stub)",
		"sub", n.SubscriptionID,
		"signal", n.SignalID,
		"token", n.APNsToken,
		"title", n.Title,
		"body", n.Body,
		"url", n.URL,
	)
	return nil
}

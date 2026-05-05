package agent

import (
	"context"
	"errors"
	"net"
	"time"
)

// ErrorKind classifies a pipeline failure for the scheduler. Used as the
// last_error_kind column on subscriptions and as the input to BackoffPolicy.
type ErrorKind string

const (
	// ErrKindRateLimited — Brave 429 (or analogous transient throttling).
	// Short backoff, recoverable on its own.
	ErrKindRateLimited ErrorKind = "rate_limited"

	// ErrKindBraveQuotaExceeded — Brave 402, monthly usage cap. Permanent
	// until the billing cycle resets. Long fixed pause; retrying every
	// cadence does nothing but waste OpenAI plan calls.
	ErrKindBraveQuotaExceeded ErrorKind = "brave_quota_exceeded"

	// ErrKindUpstreamUnavailable — 5xx, network errors, timeouts. Transient.
	ErrKindUpstreamUnavailable ErrorKind = "upstream_unavailable"

	// ErrKindUnknown — anything we didn't recognize. Treat as transient
	// but with medium backoff so a misclassified permanent error doesn't
	// thrash forever.
	ErrKindUnknown ErrorKind = "unknown"
)

// ClassifyError maps a Go error to an ErrorKind. Callers use this to choose
// a backoff delay and to record a structured kind on the subscription.
func ClassifyError(err error) ErrorKind {
	if err == nil {
		return ""
	}

	// Brave HTTP errors carry the status code. Check first because they
	// are the most common operational failure.
	var bErr *BraveError
	if errors.As(err, &bErr) {
		switch {
		case bErr.StatusCode == 402:
			return ErrKindBraveQuotaExceeded
		case bErr.StatusCode == 429:
			return ErrKindRateLimited
		case bErr.StatusCode >= 500 && bErr.StatusCode < 600:
			return ErrKindUpstreamUnavailable
		}
	}

	// Network-layer issues — DNS failure, connection refused, TLS, etc.
	var netErr net.Error
	if errors.As(err, &netErr) {
		return ErrKindUpstreamUnavailable
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return ErrKindUpstreamUnavailable
	}

	return ErrKindUnknown
}

// BackoffDelay picks how long to wait before retrying after a failure.
//
// Permanent errors (Brave monthly cap) get a long fixed pause: there's no
// point retrying within the same billing cycle, and each retry burns a
// planner call.
//
// Transient errors use exponential backoff anchored on the subscription's
// own cadence: delay = cadence * 2^failureCount, capped at maxBackoff.
// This means a sub configured for 1h cadence pauses 2h, then 4h, etc.,
// instead of hammering at the 10-second tick rate.
func BackoffDelay(kind ErrorKind, cadence time.Duration, failureCount int) time.Duration {
	const (
		minBackoff      = 1 * time.Minute
		maxBackoff      = 24 * time.Hour
		quotaBackoff    = 6 * time.Hour // Brave billing cycles are monthly; 6h = 4 retries/day
		rateLimitMin    = 30 * time.Second
		rateLimitGrowth = 2
	)

	switch kind {
	case ErrKindBraveQuotaExceeded:
		return quotaBackoff

	case ErrKindRateLimited:
		// Rate limits typically clear in seconds. Don't anchor on cadence —
		// users with hourly subs shouldn't wait an hour for a 1-second
		// throttle to clear. Grow fast to avoid synchronized retries.
		d := rateLimitMin << failureCount // 30s, 60s, 120s, 240s, ...
		if d > maxBackoff || d <= 0 {
			return maxBackoff
		}
		if d < rateLimitMin {
			return rateLimitMin
		}
		_ = rateLimitGrowth
		return d

	case ErrKindUpstreamUnavailable, ErrKindUnknown:
		base := cadence
		if base < minBackoff {
			base = minBackoff
		}
		// d = base * 2^failureCount, with overflow guard.
		d := base
		for i := 0; i < failureCount && d < maxBackoff; i++ {
			d *= 2
		}
		if d > maxBackoff || d <= 0 {
			return maxBackoff
		}
		return d
	}
	return minBackoff
}

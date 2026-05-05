package agent

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClassifyError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want ErrorKind
	}{
		{"nil", nil, ""},
		{"brave 402", &BraveError{StatusCode: 402, Body: "{}"}, ErrKindBraveQuotaExceeded},
		{"brave 429", &BraveError{StatusCode: 429, Body: "{}"}, ErrKindRateLimited},
		{"brave 500", &BraveError{StatusCode: 500, Body: "{}"}, ErrKindUpstreamUnavailable},
		{"brave 400", &BraveError{StatusCode: 400, Body: "{}"}, ErrKindUnknown},
		{"context deadline", context.DeadlineExceeded, ErrKindUpstreamUnavailable},
		{"net op error", &net.OpError{Op: "dial", Err: errors.New("refused")}, ErrKindUpstreamUnavailable},
		{"random", errors.New("plain"), ErrKindUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, ClassifyError(tc.err))
		})
	}
}

func TestBackoffDelay_QuotaExceededIsLongFixed(t *testing.T) {
	// Brave monthly cap doesn't reset until the billing cycle does, so
	// failure_count shouldn't compound the wait — a single long pause is
	// the right answer regardless of attempt count.
	short := BackoffDelay(ErrKindBraveQuotaExceeded, time.Hour, 0)
	long := BackoffDelay(ErrKindBraveQuotaExceeded, time.Hour, 10)
	require.Equal(t, short, long)
	require.GreaterOrEqual(t, short, time.Hour, "quota backoff should be hours, not minutes")
}

func TestBackoffDelay_RateLimitedGrowsButCaps(t *testing.T) {
	// Rate-limit backoff anchors on a short floor (not cadence) because
	// 429s typically clear in seconds, not hours.
	d0 := BackoffDelay(ErrKindRateLimited, time.Hour, 0)
	d1 := BackoffDelay(ErrKindRateLimited, time.Hour, 1)
	d4 := BackoffDelay(ErrKindRateLimited, time.Hour, 4)
	d20 := BackoffDelay(ErrKindRateLimited, time.Hour, 20)

	require.Less(t, d0, d1, "should grow with failure count")
	require.Less(t, d1, d4)
	require.LessOrEqual(t, d20, 24*time.Hour, "must cap")
	require.GreaterOrEqual(t, d0, 30*time.Second, "first retry must wait at least 30s")
}

func TestBackoffDelay_TransientAnchorsOnCadence(t *testing.T) {
	// Upstream-unavailable backoff scales with the user's chosen cadence —
	// hourly subs wait hours after upstream errors, not seconds.
	d := BackoffDelay(ErrKindUpstreamUnavailable, time.Hour, 0)
	require.Equal(t, time.Hour, d)

	// Exponential growth.
	d2 := BackoffDelay(ErrKindUpstreamUnavailable, time.Hour, 2)
	require.Equal(t, 4*time.Hour, d2)

	// Cap.
	dHuge := BackoffDelay(ErrKindUpstreamUnavailable, time.Hour, 50)
	require.LessOrEqual(t, dHuge, 24*time.Hour)
}

func TestBackoffDelay_EnforcesFloorForTinyCadence(t *testing.T) {
	// A misconfigured 1-second cadence shouldn't translate to 1-second
	// retry storms on transient errors.
	d := BackoffDelay(ErrKindUnknown, 1*time.Second, 0)
	require.GreaterOrEqual(t, d, time.Minute)
}

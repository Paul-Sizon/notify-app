-- +goose Up
-- Track failure state on each subscription so the scheduler can back off
-- after upstream errors instead of retrying every tick.
--
-- failure_count: number of consecutive failures since the last success.
--   Reset to 0 on RescheduleSubscription (success path).
--   Used as the exponent for transient backoff: delay = base * 2^failures.
--
-- last_error_kind: short string classifier set by the agent on failure:
--   "" (success/no error yet), "rate_limited", "brave_quota_exceeded",
--   "upstream_unavailable", "unknown". Surfaced to ops; also lets us pick
--   a long fixed pause for permanent errors (e.g. monthly cap exceeded).
ALTER TABLE subscriptions
  ADD COLUMN failure_count   INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN last_error_kind TEXT    NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE subscriptions
  DROP COLUMN failure_count,
  DROP COLUMN last_error_kind;

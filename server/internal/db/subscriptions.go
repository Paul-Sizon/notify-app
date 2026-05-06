package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SubscriptionInsert struct {
	DeviceID       uuid.UUID
	Query          string
	Type           string
	CadenceSeconds int
}

// subColumns is the canonical column list for subscription reads. Kept in
// one place so adding columns doesn't require touching every query.
const subColumns = `id, device_id, query, type, cadence_seconds, rolling_summary,
	last_run_at, next_run_at, created_at, failure_count, last_error_kind`

// scanSub fits the columns from subColumns. row is whatever supports Scan.
func scanSub(row interface{ Scan(...any) error }, s *Subscription) error {
	return row.Scan(
		&s.ID, &s.DeviceID, &s.Query, &s.Type, &s.CadenceSeconds, &s.RollingSummary,
		&s.LastRunAt, &s.NextRunAt, &s.CreatedAt, &s.FailureCount, &s.LastErrorKind,
	)
}

func (d *DB) InsertSubscription(ctx context.Context, in SubscriptionInsert) (Subscription, error) {
	const q = `
		INSERT INTO subscriptions (device_id, query, type, cadence_seconds)
		VALUES ($1,$2,$3,$4)
		RETURNING ` + subColumns
	var s Subscription
	err := scanSub(d.Pool.QueryRow(ctx, q, in.DeviceID, in.Query, in.Type, in.CadenceSeconds), &s)
	return s, err
}

func (d *DB) GetSubscription(ctx context.Context, id uuid.UUID) (Subscription, error) {
	const q = `SELECT ` + subColumns + ` FROM subscriptions WHERE id = $1`
	var s Subscription
	err := scanSub(d.Pool.QueryRow(ctx, q, id), &s)
	if errors.Is(err, pgx.ErrNoRows) {
		return Subscription{}, ErrNotFound
	}
	return s, err
}

func (d *DB) ListSubscriptionsByDevice(ctx context.Context, deviceID uuid.UUID) ([]Subscription, error) {
	const q = `SELECT ` + subColumns + ` FROM subscriptions WHERE device_id = $1 ORDER BY created_at DESC`
	rows, err := d.Pool.Query(ctx, q, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Subscription
	for rows.Next() {
		var s Subscription
		if err := scanSub(rows, &s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (d *DB) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	_, err := d.Pool.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	return err
}

// DeleteSubscriptionsByDevice removes every subscription belonging to the
// given device. Cascades to signals via the FK.
func (d *DB) DeleteSubscriptionsByDevice(ctx context.Context, deviceID uuid.UUID) (int64, error) {
	tag, err := d.Pool.Exec(ctx, `DELETE FROM subscriptions WHERE device_id = $1`, deviceID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// DeleteAllSubscriptions wipes every subscription across every device.
// Admin-only; signals cascade away too.
func (d *DB) DeleteAllSubscriptions(ctx context.Context) (int64, error) {
	tag, err := d.Pool.Exec(ctx, `DELETE FROM subscriptions`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// ListDueSubscriptions returns subscriptions whose next_run_at <= now, up to limit.
func (d *DB) ListDueSubscriptions(ctx context.Context, now time.Time, limit int) ([]Subscription, error) {
	const q = `SELECT ` + subColumns + `
		FROM subscriptions WHERE next_run_at <= $1 ORDER BY next_run_at ASC LIMIT $2`
	rows, err := d.Pool.Query(ctx, q, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Subscription
	for rows.Next() {
		var s Subscription
		if err := scanSub(rows, &s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// RescheduleSubscription is the SUCCESS path. It updates last_run_at=now,
// next_run_at=now+cadence, optionally rewrites rolling_summary, and resets
// failure_count to 0 / last_error_kind to '' so any prior backoff state
// is cleared.
func (d *DB) RescheduleSubscription(ctx context.Context, id uuid.UUID, now time.Time, rollingSummary string) error {
	if rollingSummary != "" {
		_, err := d.Pool.Exec(ctx, `
			UPDATE subscriptions
			SET last_run_at = $2,
			    next_run_at = ($2::timestamptz) + (cadence_seconds || ' seconds')::interval,
			    rolling_summary = $3,
			    failure_count = 0,
			    last_error_kind = ''
			WHERE id = $1`, id, now, rollingSummary)
		return err
	}
	_, err := d.Pool.Exec(ctx, `
		UPDATE subscriptions
		SET last_run_at = $2,
		    next_run_at = ($2::timestamptz) + (cadence_seconds || ' seconds')::interval,
		    failure_count = 0,
		    last_error_kind = ''
		WHERE id = $1`, id, now)
	return err
}

// BackoffSubscription is the FAILURE path. It records last_run_at=now (the
// run did happen, it just failed), advances next_run_at by `delay`,
// increments failure_count, and stamps last_error_kind. The caller computes
// `delay` based on error kind + current failure_count so this method stays
// purely about persistence.
//
// Crucial property: delay is always > 0, so the failing sub is removed from
// the "due" set until the next backoff window — no more 10-second retry
// storms while an upstream is down.
func (d *DB) BackoffSubscription(ctx context.Context, id uuid.UUID, now time.Time, delay time.Duration, errorKind string) error {
	_, err := d.Pool.Exec(ctx, `
		UPDATE subscriptions
		SET last_run_at = $2,
		    next_run_at = $3,
		    failure_count = failure_count + 1,
		    last_error_kind = $4
		WHERE id = $1`, id, now, now.Add(delay), errorKind)
	return err
}

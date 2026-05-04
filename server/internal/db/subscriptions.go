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

func (d *DB) InsertSubscription(ctx context.Context, in SubscriptionInsert) (Subscription, error) {
	const q = `
		INSERT INTO subscriptions (device_id, query, type, cadence_seconds)
		VALUES ($1,$2,$3,$4)
		RETURNING id, device_id, query, type, cadence_seconds, rolling_summary, last_run_at, next_run_at, created_at`
	var s Subscription
	err := d.Pool.QueryRow(ctx, q, in.DeviceID, in.Query, in.Type, in.CadenceSeconds).Scan(
		&s.ID, &s.DeviceID, &s.Query, &s.Type, &s.CadenceSeconds, &s.RollingSummary, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt,
	)
	return s, err
}

func (d *DB) GetSubscription(ctx context.Context, id uuid.UUID) (Subscription, error) {
	const q = `
		SELECT id, device_id, query, type, cadence_seconds, rolling_summary, last_run_at, next_run_at, created_at
		FROM subscriptions WHERE id = $1`
	var s Subscription
	err := d.Pool.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.DeviceID, &s.Query, &s.Type, &s.CadenceSeconds, &s.RollingSummary, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Subscription{}, ErrNotFound
	}
	return s, err
}

func (d *DB) ListSubscriptionsByDevice(ctx context.Context, deviceID uuid.UUID) ([]Subscription, error) {
	const q = `
		SELECT id, device_id, query, type, cadence_seconds, rolling_summary, last_run_at, next_run_at, created_at
		FROM subscriptions WHERE device_id = $1 ORDER BY created_at DESC`
	rows, err := d.Pool.Query(ctx, q, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.DeviceID, &s.Query, &s.Type, &s.CadenceSeconds, &s.RollingSummary, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt); err != nil {
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

// ListDueSubscriptions returns subscriptions whose next_run_at <= now, up to limit.
func (d *DB) ListDueSubscriptions(ctx context.Context, now time.Time, limit int) ([]Subscription, error) {
	const q = `
		SELECT id, device_id, query, type, cadence_seconds, rolling_summary, last_run_at, next_run_at, created_at
		FROM subscriptions WHERE next_run_at <= $1 ORDER BY next_run_at ASC LIMIT $2`
	rows, err := d.Pool.Query(ctx, q, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.DeviceID, &s.Query, &s.Type, &s.CadenceSeconds, &s.RollingSummary, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// RescheduleSubscription updates last_run_at=now and next_run_at=now+cadence.
// Optionally updates rolling_summary if non-empty.
func (d *DB) RescheduleSubscription(ctx context.Context, id uuid.UUID, now time.Time, rollingSummary string) error {
	if rollingSummary != "" {
		_, err := d.Pool.Exec(ctx, `
			UPDATE subscriptions
			SET last_run_at = $2,
			    next_run_at = ($2::timestamptz) + (cadence_seconds || ' seconds')::interval,
			    rolling_summary = $3
			WHERE id = $1`, id, now, rollingSummary)
		return err
	}
	_, err := d.Pool.Exec(ctx, `
		UPDATE subscriptions
		SET last_run_at = $2,
		    next_run_at = ($2::timestamptz) + (cadence_seconds || ' seconds')::interval
		WHERE id = $1`, id, now)
	return err
}

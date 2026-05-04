package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// InsertSignals bulk-inserts signals with ON CONFLICT DO NOTHING on
// (subscription_id, fingerprint). Returns IDs of rows actually inserted (i.e. new signals).
func (d *DB) InsertSignals(ctx context.Context, sigs []SignalInsert) ([]uuid.UUID, error) {
	if len(sigs) == 0 {
		return nil, nil
	}
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	ids := make([]uuid.UUID, 0, len(sigs))
	const q = `
		INSERT INTO signals (subscription_id, fingerprint, title, body, url, occurs_at, source_domains, confidence, payload)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (subscription_id, fingerprint) DO NOTHING
		RETURNING id`
	for _, s := range sigs {
		var id uuid.UUID
		err := tx.QueryRow(ctx, q,
			s.SubscriptionID, s.Fingerprint, s.Title, s.Body, s.URL, s.OccursAt, s.SourceDomains, s.Confidence, s.Payload,
		).Scan(&id)
		if err == nil {
			ids = append(ids, id)
			continue
		}
		if errors.Is(err, pgx.ErrNoRows) {
			continue // ON CONFLICT DO NOTHING fired
		}
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ids, nil
}

func (d *DB) ListSignalsBySubscription(ctx context.Context, subID uuid.UUID, before time.Time, limit int) ([]Signal, error) {
	const q = `
		SELECT id, subscription_id, fingerprint, title, body, url, occurs_at, source_domains, confidence, payload, first_seen_at, notified_at
		FROM signals
		WHERE subscription_id = $1 AND first_seen_at < $2
		ORDER BY first_seen_at DESC
		LIMIT $3`
	rows, err := d.Pool.Query(ctx, q, subID, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Signal
	for rows.Next() {
		var s Signal
		if err := rows.Scan(&s.ID, &s.SubscriptionID, &s.Fingerprint, &s.Title, &s.Body, &s.URL, &s.OccursAt, &s.SourceDomains, &s.Confidence, &s.Payload, &s.FirstSeenAt, &s.NotifiedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (d *DB) MarkSignalNotified(ctx context.Context, id uuid.UUID, when time.Time) error {
	_, err := d.Pool.Exec(ctx, `UPDATE signals SET notified_at = $2 WHERE id = $1`, id, when)
	return err
}

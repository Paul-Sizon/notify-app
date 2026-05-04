package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// UpsertDevice creates or refreshes a device row keyed by apns_token.
// Returns the device ID. Idempotent on apns_token.
func (d *DB) UpsertDevice(ctx context.Context, apnsToken string) (uuid.UUID, error) {
	const q = `
		INSERT INTO devices (apns_token)
		VALUES ($1)
		ON CONFLICT (apns_token) DO UPDATE SET last_seen_at = NOW()
		RETURNING id`
	var id uuid.UUID
	err := d.Pool.QueryRow(ctx, q, apnsToken).Scan(&id)
	return id, err
}

func (d *DB) GetDevice(ctx context.Context, id uuid.UUID) (Device, error) {
	const q = `SELECT id, apns_token, created_at, last_seen_at FROM devices WHERE id = $1`
	var dev Device
	err := d.Pool.QueryRow(ctx, q, id).Scan(&dev.ID, &dev.APNsToken, &dev.CreatedAt, &dev.LastSeenAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Device{}, ErrNotFound
	}
	return dev, err
}

func (d *DB) DeleteDevice(ctx context.Context, id uuid.UUID) error {
	_, err := d.Pool.Exec(ctx, `DELETE FROM devices WHERE id = $1`, id)
	return err
}

var ErrNotFound = errors.New("not found")

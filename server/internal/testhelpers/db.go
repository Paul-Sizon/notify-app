package testhelpers

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDBPool connects to TEST_DATABASE_URL, truncates all tables, and registers cleanup.
// Caller skips test if env var missing — works for CI without postgres.
func TestDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL missing")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	// truncate to fresh state
	_, err = pool.Exec(context.Background(), `TRUNCATE devices, subscriptions, signals RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

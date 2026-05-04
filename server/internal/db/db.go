package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *DB { return &DB{Pool: pool} }

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 8
	cfg.MaxConnLifetime = 30 * time.Minute
	return pgxpool.NewWithConfig(ctx, cfg)
}

type Device struct {
	ID         uuid.UUID
	APNsToken  string
	CreatedAt  time.Time
	LastSeenAt time.Time
}

type Subscription struct {
	ID             uuid.UUID
	DeviceID       uuid.UUID
	Query          string
	Type           string
	CadenceSeconds int
	RollingSummary string
	LastRunAt      *time.Time
	NextRunAt      time.Time
	CreatedAt      time.Time
}

type Signal struct {
	ID              uuid.UUID
	SubscriptionID  uuid.UUID
	Fingerprint     string
	Title           string
	Body            *string
	URL             *string
	OccursAt        *time.Time
	SourceDomains   []string
	Confidence      float32
	Payload         []byte // JSONB
	FirstSeenAt     time.Time
	NotifiedAt      *time.Time
}

type SignalInsert struct {
	SubscriptionID uuid.UUID
	Fingerprint    string
	Title          string
	Body           *string
	URL            *string
	OccursAt       *time.Time
	SourceDomains  []string
	Confidence     float32
	Payload        []byte
}

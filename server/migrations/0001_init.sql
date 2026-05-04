-- +goose Up
CREATE TABLE devices (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  apns_token    TEXT UNIQUE NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  device_id        UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  query            TEXT NOT NULL,
  type             TEXT NOT NULL CHECK (type IN ('event', 'news')),
  cadence_seconds  INTEGER NOT NULL CHECK (cadence_seconds >= 300),
  rolling_summary  TEXT NOT NULL DEFAULT '',
  last_run_at      TIMESTAMPTZ,
  next_run_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX subscriptions_device_idx   ON subscriptions(device_id);
CREATE INDEX subscriptions_next_run_idx ON subscriptions(next_run_at);

CREATE TABLE signals (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subscription_id  UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
  fingerprint      TEXT NOT NULL,
  title            TEXT NOT NULL,
  body             TEXT,
  url              TEXT,
  occurs_at        TIMESTAMPTZ,
  source_domains   TEXT[] NOT NULL DEFAULT '{}',
  confidence       REAL NOT NULL,
  payload          JSONB NOT NULL,
  first_seen_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  notified_at      TIMESTAMPTZ,
  UNIQUE (subscription_id, fingerprint)
);
CREATE INDEX signals_sub_seen_idx ON signals(subscription_id, first_seen_at DESC);

-- +goose Down
DROP TABLE signals;
DROP TABLE subscriptions;
DROP TABLE devices;

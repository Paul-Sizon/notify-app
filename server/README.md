# Signal Monitor — Server

Go backend. Web-grounded agent that monitors user-defined queries via Brave Answers + OpenAI structured extraction, dedupes signals, and pushes new findings via APNs.

## Quickstart

```bash
# 1. Postgres
make up
make migrate
make migrate-test

# 2. Unit tests (no API, no DB)
make test

# 3. Integration tests (DB + real OpenAI/Brave)
make test-integration

# 4. E2E (full HTTP stack, real APIs)
make test-e2e

# 5. Run server
make run
```

## Layout

```
cmd/server/         single binary (HTTP + scheduler)
internal/
  agent/            Searcher (Brave), Extractor (OpenAI), RunSubscription
  api/              chi handlers + middleware
  config/           env loader
  db/               pgx queries (no ORM)
  push/             Pusher iface + APNs + log stub
  scheduler/        polling ticker + in-flight lock
  testhelpers/      shared TestDBPool helper
e2e/                black-box HTTP client + tests
migrations/         goose .sql
```

## Test layers

| Layer | Build tag | What runs | Cost |
|---|---|---|---|
| unit | (none) | pure helpers, validation, fingerprint | free |
| integration | `integration` | DB tests, API+DB, real Brave, real OpenAI | low ($) |
| e2e | `e2e` | full HTTP stack with real APIs | ~$0.05/run |

## API

Base path `/v1`. Auth via `X-Device-Id` header (UUID returned from `POST /devices`).

| Method | Path | Notes |
|---|---|---|
| POST | `/devices` | idempotent on `apns_token` |
| POST | `/subscriptions` | `query`, `type` (`event`\|`news`), `cadence_seconds` ≥300 |
| GET  | `/subscriptions` | scoped to caller device |
| DELETE | `/subscriptions/:id` | cascades signals |
| GET  | `/subscriptions/:id/signals?limit=50&before=<RFC3339>` | newest-first |
| POST | `/subscriptions/:id/run` | force agent run |
| GET  | `/healthz` | liveness |

## Env

See `.env.example`. Required: `DATABASE_URL`, `OPENAI_API_KEY`, `BRAVE_API_KEY`. APNs vars optional — without them, `LogPusher` prints notifications to stdout (Day 1 demo mode).

## Notes

- Brave Answers does not return structured citations. URL fields are best-effort from prose.
- LLM nondeterminism: re-running the same query may produce slightly different fingerprints. Strict dedup is verified at the DB layer (`TestInsertSignals_OnConflictReturnsOnlyNew`); pipeline-level dedup is best-effort across runs.
- Single binary; api + scheduler share the process. Re-split later if needed.

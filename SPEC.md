# Signal Monitor — Tech Spec

A backend agent that periodically searches the web for user-defined queries, extracts and dedupes signals, and pushes only genuinely new findings to an iOS app.

---

## 1. Goals

- Backend in Go that runs N subscriptions on cadence, each producing 0..M new signals per tick.
- iOS app (KMP shared core + native SwiftUI) for creating subscriptions, browsing signals, receiving push notifications.
- Two subscription types with different dedup strategies:
  - `event` — discrete items with stable identity (concerts, meetups). Dedup by normalized fingerprint.
  - `news` — continuous topic monitoring (regulatory changes, market moves). Dedup by LLM diff against rolling summary.
- Push delivery via APNs.

## 2. Non-goals (MVP)

- No auth, no accounts. Device ID = user identity.
- No Android UI (KMP gives the option; do not build it).
- No subscription edit (delete + recreate).
- No agent self-subscribing to newsletters/APIs (Phase 2).
- No background app refresh on iOS — server pushes everything.

## 3. Stack

| Layer | Choice | Notes |
|---|---|---|
| Language (server) | Go 1.22+ | |
| HTTP | chi | |
| DB | Postgres 16 | `gen_random_uuid()` is built-in, no extension needed |
| DB access | sqlc | typed queries from SQL |
| Migrations | goose | |
| Job queue | river | Postgres-backed, no Redis needed |
| Search | Tavily | `search_depth=advanced`, `include_raw_content=true`. Falls back to Brave behind an interface if needed |
| LLM | OpenAI `gpt-4o-mini` | use `response_format: json_schema` (Structured Outputs), not plain JSON mode |
| Push | sideshow/apns2 | direct APNs, .p8 auth key |
| Mobile shared | Kotlin Multiplatform + Ktor client + kotlinx.serialization | |
| Mobile UI | SwiftUI, iOS 17+ target | |

All external services behind interfaces (`Searcher`, `Extractor`, `Pusher`) so they can be swapped or stubbed in tests.

## 4. Architecture

```
                                  ┌──────────────┐
        iOS app  ───── HTTPS ───→ │  api server  │ ──→ Postgres
            ↑                     └──────────────┘
            │                            ↑
            │                            │ enqueue
            │                     ┌──────────────┐
            │                     │  scheduler   │  (river periodic)
            │                     └──────────────┘
            │                            │
            │                            ↓ run agent job
            │                     ┌──────────────┐
            │                     │   worker     │ ──→ Tavily, OpenAI
            │                     └──────────────┘
            │                            │
            │                            ↓ enqueue push job
            │                     ┌──────────────┐
            └────── APNs ──────── │  push sender │
                                  └──────────────┘
```

Three Go binaries sharing one module:
- `cmd/api` — HTTP server
- `cmd/worker` — runs all river jobs (agent + push)
- `cmd/scheduler` — periodic enqueuer (or fold into worker via river's periodic jobs)

For MVP, fold scheduler into worker. One binary in production: `cmd/worker` that does both periodic scheduling and job execution. Keep `cmd/api` separate so it can scale independently.

## 5. Data model

```sql
-- 0001_init.sql
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
  occurs_at        TIMESTAMPTZ,           -- nullable; populated for events with known date
  source_domains   TEXT[] NOT NULL DEFAULT '{}',
  confidence       REAL   NOT NULL,
  payload          JSONB  NOT NULL,       -- raw LLM extraction for that item
  first_seen_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  notified_at      TIMESTAMPTZ,
  UNIQUE (subscription_id, fingerprint)
);
CREATE INDEX signals_sub_seen_idx ON signals(subscription_id, first_seen_at DESC);
```

`fingerprint` is the dedup primary key. `UNIQUE (subscription_id, fingerprint)` + `INSERT ... ON CONFLICT DO NOTHING RETURNING id` is how the agent identifies what's actually new.

## 6. HTTP API

Base path: `/v1`. All requests JSON. Auth: `X-Device-Id` header (the device ID returned from `POST /devices`). No bearer tokens for MVP.

### `POST /devices`
Register or update an APNs token. Idempotent on `apns_token`.

```
Request:  { "apns_token": "<hex>" }
Response: { "device_id": "<uuid>" }
```

### `POST /subscriptions`
```
Headers:  X-Device-Id: <uuid>
Request:  {
  "query": "blockchain events curitiba",
  "type": "event",
  "cadence_seconds": 3600
}
Response: <subscription object>
```

Validations:
- `cadence_seconds >= 300` (5 min minimum)
- `type in ('event','news')`
- `query` length 3–200

### `GET /subscriptions`
Returns subscriptions for the device, with `last_signal_at` and `unread_count` joined.

### `DELETE /subscriptions/:id`
Cascade-deletes signals.

### `GET /subscriptions/:id/signals?limit=50&before=<iso>`
Cursor pagination by `first_seen_at`.

### `POST /subscriptions/:id/run`
Force an immediate agent run (used by pull-to-refresh). Idempotent within a 30s window.

### `GET /subscriptions/:id/status` (SSE)
Streams the current run's phase events. Optional for MVP — see §10.

```
event: phase
data: { "phase": "searching", "detail": "Tavily query..." }

event: phase
data: { "phase": "extracting" }

event: done
data: { "new_signals": 2 }
```

If SSE is too much for Day 3, fake it client-side with timed phases. The user-visible value is the *visualization*, not real-time accuracy.

## 7. Agent pipeline

Single function `RunSubscription(ctx, subID) (newSignalIDs []uuid.UUID, err error)`. Pure orchestration — each step is its own injected dependency.

```
1. Load subscription from DB.
2. Search        → Tavily.Search(query, type)            → []SearchResult
3. Extract       → LLM extract with type-specific prompt → []Candidate
4. Verify        → for c.confidence < 0.8, re-search; require ≥2 distinct domains, else drop.
5. Fingerprint   → per type (see below)
6. Insert        → batch upsert with ON CONFLICT DO NOTHING RETURNING id
7. Push          → enqueue push job per new signal
8. Reschedule    → update last_run_at, next_run_at = NOW() + cadence_seconds
```

### 7.1 Search

For `type=event`: append year hints, `search_depth=advanced`, `include_raw_content=true`.
For `type=news`: `topic="news"`, `time_range="week"`, `include_raw_content=true`.

Cap raw_content to ~2k chars per result before sending to LLM.

### 7.2 Fingerprinting

**Events:**
```
fp = sha256( normalize(title) + "|" + date.Format("2006-01-02") + "|" + normalize(venue) )
```
where `normalize` = lowercase, strip punctuation, collapse whitespace, drop stopwords (`the`, `a`, `live`, `tour`, `concert`, `feat`, `featuring`).

**News:**
- LLM is asked, in the same extraction call, to flag each item `is_new_development: bool` by comparing to `rolling_summary`.
- Only items with `is_new_development=true` proceed.
- `fp = sha256( normalize(headline) )` after the LLM returns a canonicalized headline.
- After a successful run, persist the LLM-returned `updated_summary` back to `subscriptions.rolling_summary`.

This is the weakest link in the design. Mitigation: keep the rolling summary short (≤5 sentences) and rewrite it each tick rather than appending. Test with the demo queries before showtime.

### 7.3 Verification

Only triggered when `confidence < 0.8`. Second Tavily call for the specific item title. Drop if fewer than 2 distinct second-level domains corroborate. Promotes confidence to 0.8 if it passes.

### 7.4 Concurrency

- River concurrency = N workers.
- Per-subscription locking: river's unique-job feature, key = `agent_run:<sub_id>`. Prevents overlapping runs of the same subscription.
- Push jobs are independent and parallel.

## 8. LLM prompts

Both prompts use OpenAI Structured Outputs (`response_format: { type: "json_schema", ... }`). Schemas live in `internal/agent/schemas.go`.

### 8.1 Event extraction

System:
```
You extract upcoming events from web search results. You are precise. You discard
speculation, retrospectives, and irrelevant content. You never invent details.
```

User template:
```
Query: {query}
Today: {today_iso}

Search results:
{json: [{url, title, content_excerpt}]}

Return JSON matching the schema. Hard rules:
- Only events with a concrete date or date range, in the future relative to today.
- Only events that match the query intent.
- If the same event appears in multiple results, emit it once. Pick the most canonical URL
  (prefer official site > ticket page > listing aggregator > article).
- Empty array is the correct answer when nothing qualifies.
```

Response schema:
```json
{
  "type": "object",
  "properties": {
    "events": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "title":     { "type": "string" },
          "date":      { "type": ["string", "null"], "description": "ISO 8601" },
          "venue":     { "type": ["string", "null"] },
          "city":      { "type": ["string", "null"] },
          "url":       { "type": "string" },
          "confidence":{ "type": "number", "minimum": 0, "maximum": 1 }
        },
        "required": ["title","date","venue","city","url","confidence"],
        "additionalProperties": false
      }
    }
  },
  "required": ["events"],
  "additionalProperties": false
}
```

### 8.2 News extraction

System:
```
You monitor news for material new developments on a topic. You distinguish genuinely
new facts from commentary, retrospectives, and rephrased coverage of known events.
You prefer primary sources (regulators, official announcements) over secondary coverage.
```

User template:
```
Topic: {query}
Today: {today_iso}

What has already been reported to the user (do not repeat substantively similar items):
"""
{rolling_summary}
"""

Search results:
{json: [{url, title, content_excerpt, published_at}]}

Return JSON matching the schema. Hard rules:
- is_new_development must be false for any item that is just commentary on, or rephrasing
  of, something already in the rolling summary above.
- Discard opinion pieces and speculation.
- updated_summary must be 3–5 sentences capturing all material developments now known,
  including any new ones from this batch. Rewrite it; do not append.
```

Response schema:
```json
{
  "type": "object",
  "properties": {
    "items": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "headline":              { "type": "string" },
          "canonical_headline":    { "type": "string" },
          "summary":               { "type": "string" },
          "url":                   { "type": "string" },
          "published_at":          { "type": ["string", "null"] },
          "confidence":            { "type": "number", "minimum": 0, "maximum": 1 },
          "is_new_development":    { "type": "boolean" }
        },
        "required": ["headline","canonical_headline","summary","url","published_at","confidence","is_new_development"],
        "additionalProperties": false
      }
    },
    "updated_summary": { "type": "string" }
  },
  "required": ["items","updated_summary"],
  "additionalProperties": false
}
```

## 9. Push notifications

APNs production + sandbox. Use the bundle ID + Team ID + Key ID + .p8 file.

Payload:
```json
{
  "aps": {
    "alert":    { "title": "<subscription.query>", "body": "<signal.title>" },
    "sound":    "default",
    "badge":    1,
    "thread-id":"<subscription_id>",
    "mutable-content": 1
  },
  "subscription_id": "<uuid>",
  "signal_id":       "<uuid>",
  "url":             "<signal.url>"
}
```

Push job retry policy: 3 attempts, exponential backoff. On `BadDeviceToken` from APNs, soft-delete the device row.

Notification grouping uses `thread-id = subscription_id` so multiple signals from the same subscription stack on the lock screen.

## 10. iOS app

### Shared (KMP)

```
shared/src/commonMain/kotlin/
  com/signal/
    api/
      ApiClient.kt          // Ktor client, all endpoints
      Models.kt             // @Serializable Subscription, Signal, Device
    DeviceIdProvider.kt     // expect/actual; iOS uses Keychain
```

The iOS layer interacts only with `ApiClient` and the model classes. No `expect/actual` for UI anything.

### iOS (Swift)

```
iosApp/SignalApp/
  App.swift                  // @main, AppDelegate adapter for APNs
  AppContainer.swift         // DI: ApiClient, NotificationCenter, DeepLinker
  APNs/
    APNsRegistration.swift   // request permission, register, post token to /devices
    NotificationHandler.swift// foreground/background/tap routing
  Features/
    Subscriptions/
      SubscriptionsListView.swift
      SubscriptionRowView.swift
      SubscriptionsViewModel.swift
    AddSubscription/
      AddSubscriptionView.swift  // sheet
      AddSubscriptionViewModel.swift
    SignalDetail/
      SignalDetailView.swift     // SFSafariViewController wrapper
    LiveAgent/
      LiveAgentView.swift        // SSE consumer or simulated phases
  DeepLink.swift               // signal://subscription/<id>/signal/<id>
```

### Screens

**SubscriptionsListView**
- Cards: query, type icon, cadence chip, last-signal-relative-time, unread badge.
- Pull to refresh fires `POST /subscriptions/:id/run` on each visible row.
- Swipe-to-delete with confirmation haptic.
- Empty state with one good line, no stock illustration.

**AddSubscriptionView (sheet)**
- TextField for query.
- Segmented control: Event / News.
- Picker: 5 min / 15 min / 1 hour / 6 hours / 1 day.
- Create button uses `.sensoryFeedback(.success, trigger:)`.
- Validation client-side, mirrored server-side.

**SignalDetailView**
- Title, body, source domains as chips.
- "Open source" → SFSafariViewController.
- Tapping a notification deep-links here directly.

**LiveAgentView**
- Modal sheet that appears when a sub is mid-run (manually triggered or in flight from pull-to-refresh).
- Phases stream in: Searching → Extracting → Verifying → Deduping → Done.
- Each phase has a subtle `matchedGeometryEffect` transition.
- This is the moat made visible. Spend disproportionate polish time here.

### Polish requirements

- All transitions 0.25–0.35s, spring damping 0.8.
- `matchedGeometryEffect` between list cards and detail view.
- Haptics: `.impactOccurred(.medium)` on create, `.selectionChanged` on picker, `.impactOccurred(.soft)` on delete.
- `ContainerRelativeShape` on cards for adaptive corner radius.
- iOS 17 `Liquid Glass` material on the LiveAgentView background.
- Empty states: custom copy, not generic.
- Dark mode tested.

## 11. Project layout

```
.
├── backend/
│   ├── cmd/
│   │   ├── api/main.go
│   │   └── worker/main.go
│   ├── internal/
│   │   ├── api/           # chi handlers, middleware
│   │   ├── agent/         # search, extract, verify, fingerprint
│   │   ├── push/          # APNs sender
│   │   ├── queue/         # river job definitions
│   │   ├── db/            # sqlc-generated
│   │   └── config/        # env loading
│   ├── migrations/        # goose .sql files
│   ├── sqlc.yaml
│   ├── Dockerfile
│   ├── docker-compose.yml # api + worker + postgres
│   └── go.mod
├── ios/
│   ├── shared/            # KMP module
│   └── iosApp/            # Xcode project
└── SPEC.md                # this file
```

## 12. Build & run

```bash
# DB up
cd backend && docker compose up -d postgres

# Run migrations (goose). -dir is the migration folder; "up" applies all pending.
goose -dir migrations postgres "$DATABASE_URL" up

# Generate typed query code from SQL (sqlc reads sqlc.yaml).
sqlc generate

# API server (port 8080).
go run ./cmd/api

# Worker + scheduler in one process.
go run ./cmd/worker

# One-shot agent run for debugging a single subscription.
# This bypasses the scheduler and exercises the full pipeline once.
go run ./cmd/worker -once -subscription-id=<uuid>

# iOS: open Xcode workspace, set team + bundle ID + APNs key in Signing.
open ios/iosApp/SignalApp.xcworkspace
```

`DATABASE_URL`, `OPENAI_API_KEY`, `TAVILY_API_KEY`, `APNS_KEY_PATH`, `APNS_KEY_ID`, `APNS_TEAM_ID`, `APNS_BUNDLE_ID`, `APNS_PRODUCTION` are environment variables. Load via `.env` for dev; mount as Docker secrets for deploy.

## 13. Day-by-day acceptance criteria

### Day 1 — Backend skeleton + agent end-to-end (no UI, no push)

- [ ] `goose up` applies cleanly from empty DB.
- [ ] `sqlc generate` produces `internal/db/*.go` with no errors.
- [ ] `POST /v1/devices` and `POST /v1/subscriptions` return 200 and persist rows (verified via `psql`).
- [ ] `Searcher`, `Extractor`, `Pusher` interfaces defined; Tavily and OpenAI implementations behind them.
- [ ] `go run ./cmd/worker -once -subscription-id=<uuid>` runs the full pipeline (search → extract → verify → fingerprint → insert) and prints `would_notify: [...]` to stdout. **No APNs yet.**
- [ ] Re-running the same one-shot produces zero new signals (dedup works).

### Day 2 — Scheduler + APNs

- [ ] River periodic job picks up subscriptions where `next_run_at <= NOW()` and enqueues agent runs.
- [ ] APNs key configured. Test command (`./cmd/worker -test-push -device-id=<uuid>`) sends a known payload and the device receives it.
- [ ] Agent run on tick produces push notifications for new signals only.
- [ ] iOS shell app: launches, requests notification permission, registers token, hits `POST /devices`. No real screens yet.
- [ ] **Hard fallback**: if APNs is unresolved by 17:00, switch to FCM. Decision deadline is firm.

### Day 3 — iOS app

- [ ] SubscriptionsList, AddSubscription, SignalDetail all wired to backend.
- [ ] Notification tap opens app and deep-links to signal detail.
- [ ] LiveAgentView shows phases (real SSE or simulated).
- [ ] All polish requirements in §10 met.
- [ ] Build runs on a real device, not just the simulator.

### Day 4 — Edges + demo

- [ ] Empty results: no push, no crash, log + metric.
- [ ] Malformed LLM response: retry once with stricter prompt, then skip.
- [ ] APNs returns BadDeviceToken: device soft-deleted.
- [ ] Network failure mid-run: subscription requeued, no partial state.
- [ ] Notification permission denied: app shows in-app banner, list still works.
- [ ] Three demo subscriptions pre-loaded **the day before** with real signal history:
  1. Blockchain events in Curitiba (your itch).
  2. Concert tickets for an active touring band.
  3. Cryptocurrency regulation in Vietnam (slow-cadence news).
- [ ] Backup demo video recorded.
- [ ] 90-second pitch script written and rehearsed twice.

## 14. Risks & mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| APNs first-time setup eats Day 2 | High | Hard 17:00 fallback to FCM. |
| LLM JSON inconsistency | Medium | Use Structured Outputs (`response_format: json_schema`), not JSON mode. Retry once on parse failure with stricter prompt. |
| News dedup false positives (rolling-summary diff weakness) | Medium | Keep rolling summary short. Pre-test demo queries. Have a fallback news query that's known to behave well. |
| Tavily free-tier monthly cap | Medium | Verify cap before launch. Enforce minimum 1h cadence on free tier. Upgrade if needed. |
| Event fingerprint false negatives (same event, different titles) | Medium | Normalize aggressively; if two candidates have same date+venue but different titles, treat as duplicate. |
| Wifi at venue is bad during demo | Low–Medium | Backup video. |
| Real APNs delivery has multi-second latency unpredictably | Medium | Don't promise <1s in pitch. The demo subscription that "fires live" should have signal already queued so the push arrives within a known window after creating the sub. |

## 15. Phase 2 (out of scope)

Documented here so it doesn't leak into the build:

- Agent self-subscription: per-agent email inbox parsed by LLM; OAuth flows for Eventbrite/Meetup/Songkick; Bandsintown subscription via official API.
- Multi-user accounts + auth.
- Android UI.
- Subscription edit (currently delete + recreate).
- Watch app for the LiveAgent stream.
- Per-signal user feedback ("relevant" / "noise") to fine-tune verification thresholds per subscription.

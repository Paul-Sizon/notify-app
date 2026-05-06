# Pi Deploy Spec — notify backend

Self-contained deployment spec. Hand this to Claude Code running on the Pi.
The Pi already runs Docker + Caddy + Pi-hole + Tailscale Funnel. Goal: replace
Caddy's `myapp` placeholder with the real notify Go backend, exposed publicly
via Tailscale Funnel at `https://raspberrypi.taile76757.ts.net`.

## Pre-state on Pi (verified)

- OS: Debian (linux/arm64, Pi 4 or 5)
- Docker + Docker Compose v2 installed
- Caddy running in a container, listens on host `127.0.0.1:8080`,
  reverse-proxies to placeholder container `myapp` (nginx demo)
- Pi-hole occupies host `:80` — do not bind there
- Tailscale daemon running, Funnel exposes Caddy port 8080 publicly at
  `https://raspberrypi.taile76757.ts.net`
- Repo path on Pi: clone fresh to `~/notify` (or `git pull` if exists)

## Repo

`https://github.com/Paul-Sizon/notify-app.git`, branch `main`. Commit with the
prod compose is already pushed (server/docker-compose.prod.yml,
server/Dockerfile.migrate, server/.env.prod.example).

## Final architecture

```
Internet
    │ HTTPS (Tailscale Funnel)
    ▼
raspberrypi.taile76757.ts.net  ─── Tailscale daemon
    │
    ▼ proxies to local
127.0.0.1:8080  ──── Caddy container
    │
    ▼ reverse_proxy
notify server container :8080  (joined to Caddy's docker network)
    │
    ▼
postgres container  (notify_pg_prod volume)
```

Server binds inside its container on `:8080`. Caddy reaches it by container
name across a shared docker network. Server is NOT published to the host —
public access only via Caddy + Funnel.

## Step 0 — Discover existing Caddy setup

Pi-Claude must record:

```bash
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Ports}}'
# expect rows for caddy, myapp, pihole, etc.

# Caddy container name (commonly "caddy"):
CADDY=$(docker ps --filter ancestor=caddy --format '{{.Names}}' | head -1)
echo "CADDY=$CADDY"

# Network Caddy is on:
docker inspect "$CADDY" --format '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{"\n"}}{{end}}'
# Pick the non-default one (not "bridge"). Save it.
CADDY_NET=<the name>

# Caddyfile location — usually mounted from host:
docker inspect "$CADDY" --format '{{range .Mounts}}{{.Source}} -> {{.Destination}}{{"\n"}}{{end}}'
# Look for the mount whose Destination is /etc/caddy/Caddyfile or similar.
CADDYFILE=<host path>
```

If Caddy is not running in Docker but as a host binary, adjust: there is no
`CADDY_NET`; instead expose the server on a host port (see fallback in Step 4B).

## Step 1 — Clone or update repo

```bash
cd ~
if [ -d notify ]; then
    cd notify && git pull --ff-only
else
    git clone https://github.com/Paul-Sizon/notify-app.git notify
    cd notify
fi
cd server
```

## Step 2 — Create .env (Pi-local secrets)

```bash
cp -n .env.prod.example .env
```

Then edit `~/notify/server/.env` to fill values. Required:

```
OPENAI_API_KEY=sk-proj-...
BRAVE_SEARCH_API_KEY=BSAK...
HTTP_ADDR=:8080
```

User Paul will paste real keys. Do NOT commit `.env` (root .gitignore already
excludes it). If `.env` already exists from a prior run, leave it untouched —
the `-n` flag prevents overwrite.

## Step 3 — Build + start stack

First build is slow on Pi (~5-8 min: pulls postgres:16, builds goose binary
from source, builds Go server). Subsequent rebuilds are fast.

```bash
docker compose -f docker-compose.prod.yml up -d --build
```

The compose has 3 services:
- `postgres` (postgres:16, volume `notify_pg_prod`, healthcheck)
- `migrate` (one-shot, builds from `Dockerfile.migrate`, runs goose `up`,
  exits 0 on success)
- `server` (builds from `Dockerfile`, depends on migrate completion)

Verify all came up healthy:

```bash
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs --tail=80 server
# expect log line: "http listening" addr=:8080
docker compose -f docker-compose.prod.yml logs migrate
# expect: "goose: successfully migrated database to version: 2"
```

Smoke-test from Pi host:

```bash
docker compose -f docker-compose.prod.yml exec server wget -qO- http://localhost:8080/healthz
# expect: ok
```

## Step 4 — Connect server container to Caddy network

Two routes — pick A (cleaner) unless Caddy is not in Docker.

### 4A. Caddy is in Docker (expected case)

Join the notify-server container to Caddy's network so they can talk by name:

```bash
SERVER=$(docker compose -f docker-compose.prod.yml ps -q server)
docker network connect "$CADDY_NET" "$SERVER"

# Confirm:
docker inspect "$SERVER" --format '{{range $k,$v := .NetworkSettings.Networks}}{{$k}} {{end}}'
# should list both the notify default network AND $CADDY_NET
```

To make this survive `docker compose down`/`up`, append to
`docker-compose.prod.yml` after Pi-Claude is satisfied — but skip during the
hackathon and just run `docker network connect` again after restarts. (Or add
`networks:` block referencing `$CADDY_NET` as `external: true` and attach to
the `server` service.)

The Caddy config will then reverse-proxy to `<server-container-name>:8080`.
Find the exact container name:

```bash
docker inspect "$SERVER" --format '{{.Name}}' | sed 's|^/||'
# e.g. server-server-1
```

### 4B. Fallback — Caddy not in Docker, or you can't share a network

Republish server to a host port:

In `docker-compose.prod.yml` under `server:` add (only in this fallback):

```yaml
    ports:
      - "127.0.0.1:8081:8080"
```

(Note: the file as committed already has this port mapping. Keep it for 4B,
remove it for 4A to avoid the loopback binding when a network is shared.
Either is harmless; the binding is loopback-only.)

Then in Caddy reverse-proxy to `host.docker.internal:8081`. If Caddy is in
Docker, its compose must include:

```yaml
extra_hosts:
  - "host.docker.internal:host-gateway"
```

`host-gateway` is a special Docker value (Linux, since 20.10) that resolves
to the host's bridge gateway IP.

## Step 5 — Edit Caddyfile

Open the Caddyfile path discovered in Step 0 (e.g. `/home/paul/caddy/Caddyfile`).
Locate the existing block fronting `myapp`. Replace its upstream.

Before (likely):

```
:80 {
    reverse_proxy myapp:80
}
```

After (Step 4A, container-name routing):

```
:80 {
    reverse_proxy <server-container-name>:8080
}
```

After (Step 4B, host loopback routing):

```
:80 {
    reverse_proxy host.docker.internal:8081
}
```

Caddy listens on `:80` *inside its container*. The host port mapping
`127.0.0.1:8080:80` is what makes it appear on the Pi's loopback at 8080.
Don't change the listen address; only the upstream.

Validate before reloading:

```bash
docker exec "$CADDY" caddy validate --config /etc/caddy/Caddyfile
```

(Adjust the in-container path if the Caddyfile is mounted elsewhere — the
output of Step 0 mount inspection has the Destination.)

## Step 6 — Reload Caddy (zero downtime)

```bash
docker exec "$CADDY" caddy reload --config /etc/caddy/Caddyfile
```

Reload swaps config without dropping connections. Don't `caddy stop` — that
drops Funnel briefly.

## Step 7 — Verify end-to-end

From inside the Pi:

```bash
curl -i http://127.0.0.1:8080/healthz
# expect 200, body: ok
```

From outside the Pi (e.g. user's Mac, or any internet host):

```bash
curl -i https://raspberrypi.taile76757.ts.net/healthz
# expect 200, body: ok

curl -s -X POST https://raspberrypi.taile76757.ts.net/v1/onboarding/suggest \
    -H 'content-type: application/json' \
    -d '{"city":"Lisbon","country":"PT","role":"developer","interests":["tech_meetups","ai_ml"]}' | head -c 500
# expect JSON like: {"suggestions":[{"query":"...","type":"...","cadenceSeconds":...,"reason":"..."}, ...],"fallback":false}
```

If onboarding returns `{"fallback":true}`, OPENAI_API_KEY is missing or
invalid. Check `~/notify/server/.env` and `docker compose logs server`.

DB sanity:

```bash
docker compose -f docker-compose.prod.yml exec postgres \
    psql -U notify -d notify -c '\dt'
# expect tables: subscriptions, signals, devices, goose_db_version
```

## Step 8 — Drop placeholder myapp

Once Step 7 passes:

```bash
docker rm -f myapp
```

If `myapp` is defined in a Caddy compose file, also remove its service block
to prevent it being recreated on next `docker compose up`.

## Available endpoints (for reference / future client wiring)

Listed in `server/internal/api/handler.go`:

```
GET  /healthz                                       — liveness
POST /v1/devices                                    — register device, returns id
POST /v1/onboarding/suggest                         — stateless LLM suggest, no auth
POST /v1/subscriptions          (X-Device-Id req)   — create
GET  /v1/subscriptions          (X-Device-Id req)   — list
DELETE /v1/subscriptions/{id}   (X-Device-Id req)
GET  /v1/subscriptions/{id}/signals   (X-Device-Id req)
POST /v1/subscriptions/{id}/run  (X-Device-Id req)  — force agent run
```

Background scheduler ticks every 10s, runs due subscriptions through the
agent (Brave search → OpenAI extract → push notify if APNs configured).

### Onboarding payload schema

Server validates these enums (see `server/internal/api/onboarding.go`):

- `role`: `developer | founder | designer | investor | student | other`
  (if `other`, also send `role_other` 1-60 chars)
- `interests`: any of `concerts tech_meetups crypto_web3 fintech startups_vc
  ai_ml sports art_design food_restaurants politics_policy gaming film_tv`

## Troubleshooting

**`migrate` exits non-zero:** check `docker compose logs migrate`. Most
likely postgres healthcheck never went healthy or DSN typo. Re-run
`docker compose -f docker-compose.prod.yml up -d --build migrate`.

**Server logs `apns not configured — using log-only stub pusher`:** expected
and intended. Push notifications are disabled for the hackathon; iOS app UI
flows still work.

**Caddy 502:** server container unreachable. Verify in Step 4A that the
network connect succeeded; in 4B that `extra_hosts: host-gateway` is set.

**Funnel returns Cloudflare/Tailscale error:** verify `tailscale status` on
Pi shows funnel enabled (`tailscale funnel status`). Should show
`https://raspberrypi.taile76757.ts.net (Funnel on) → http://127.0.0.1:8080`.

**Goose build fails on `go install`:** Pi may be on slow network; retry. If
network unreliable, change `Dockerfile.migrate` to pin a tag instead of
`@latest`, e.g. `@v3.21.1`.

## Rollback

```bash
cd ~/notify/server
docker compose -f docker-compose.prod.yml down
# DB volume notify_pg_prod is preserved unless you also pass -v
```

Restore Caddyfile from git or backup, then `docker exec "$CADDY" caddy
reload --config /etc/caddy/Caddyfile`. If `myapp` was removed, recreate from
its prior compose definition.

## Files modified / created on Pi

Repo (already committed upstream):
- `server/docker-compose.prod.yml`
- `server/Dockerfile.migrate`
- `server/.env.prod.example`

Pi-local, not in git:
- `~/notify/server/.env` (secrets)
- Existing Caddyfile (single-line upstream change)

Containers added: `notify-postgres-1`, `notify-migrate-1` (exits), `notify-server-1`.
Containers removed: `myapp`.

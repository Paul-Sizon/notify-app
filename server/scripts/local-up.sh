#!/usr/bin/env bash
# Multi-pane zellij stack runner.
#
# Edit the PANE_* arrays + LAYOUT block below, then run from anywhere:
#   ./scripts/local-up.sh
#
# Requires zellij. Install: brew install zellij
# Inside zellij:
#   - Move between panes:  Ctrl+P then h/j/k/l (or arrows)
#   - Fullscreen a pane:   Ctrl+P then f   (Esc to return)
#   - Detach session:      Ctrl+O then d   (`zellij attach` to resume)
#   - Quit:                Ctrl+Q
#
# Wait conditions for PANE_WAITS (each pane waits before running its command):
#   none                   - run immediately
#   delay:N                - sleep N seconds
#   port:HOST:PORT         - wait for TCP port open
#   http:URL               - wait for HTTP 2xx/3xx response
#   rpc:URL                - wait for Solana JSON-RPC getHealth
#   file:PATH              - wait for file to exist and be non-empty
#   cmd:SHELL              - poll until shell command exits 0

set -euo pipefail

# ─── EDIT THIS BLOCK ────────────────────────────────────────────────────────
# Resolve repo root regardless of where script is invoked from.
REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Cleanup stale sentinels from prior runs.
rm -f /tmp/notify-migrate-dev.done /tmp/notify-migrate-test.done /tmp/notify-integration.done

# Helper: source .env so pane commands see DATABASE_URL, TEST_DATABASE_URL,
# OPENAI_API_KEY, BRAVE_SEARCH_API_KEY.
ENVPREFIX="set -a; source $REPO/.env; set +a"

PANE_NAMES=(postgres migrate-dev migrate-test unit integration e2e server logs)
PANE_CWDS=("$REPO" "$REPO" "$REPO" "$REPO" "$REPO" "$REPO" "$REPO" "$REPO")
PANE_CMDS=(
    "docker compose up postgres"
    "$ENVPREFIX; goose -dir migrations postgres \"\$DATABASE_URL\" up && echo done > /tmp/notify-migrate-dev.done && echo '✓ dev DB migrated'; tail -f /dev/null"
    "$ENVPREFIX; goose -dir migrations postgres \"\$TEST_DATABASE_URL\" up && echo done > /tmp/notify-migrate-test.done && echo '✓ test DB migrated'; tail -f /dev/null"
    "go test -count=1 ./...; echo '── unit done ──'; tail -f /dev/null"
    "$ENVPREFIX; go test -tags=integration -count=1 -timeout 5m ./...; status=\$?; echo done > /tmp/notify-integration.done; echo \"── integration exit=\$status ──\"; tail -f /dev/null"
    "$ENVPREFIX; go test -tags=e2e -count=1 -timeout 5m ./e2e/...; echo '── e2e done ──'; tail -f /dev/null"
    "$ENVPREFIX; go run ./cmd/server"
    "exec bash -i"
)
PANE_WAITS=(
    "none"
    "port:127.0.0.1:5433"
    "port:127.0.0.1:5433"
    "none"
    "file:/tmp/notify-migrate-test.done"
    "file:/tmp/notify-integration.done"
    "file:/tmp/notify-migrate-dev.done"
    "none"
)

LAYOUT="rows"               # grid (4 panes only) | rows | columns
TAB_NAME="notify-server"
INTERACTIVE_LAST_PANE=false # if true, last pane prompts ENTER before running its command
# ────────────────────────────────────────────────────────────────────────────

if ! command -v zellij >/dev/null 2>&1; then
    cat >&2 <<EOF
✗ zellij not found in PATH

Install:
  brew install zellij
EOF
    exit 1
fi

N=${#PANE_NAMES[@]}
[ "$N" -lt 1 ] && { echo "no panes defined" >&2; exit 1; }
[ "${#PANE_CWDS[@]}"  -ne "$N" ] && { echo "PANE_CWDS length mismatch"  >&2; exit 1; }
[ "${#PANE_CMDS[@]}"  -ne "$N" ] && { echo "PANE_CMDS length mismatch"  >&2; exit 1; }
[ "${#PANE_WAITS[@]}" -ne "$N" ] && { echo "PANE_WAITS length mismatch" >&2; exit 1; }

WORK=$(mktemp -d -t zellij-stack.XXXXXX)
echo "▶ tmp pane scripts at $WORK"

emit_wait() {
    local w="$1"
    case "$w" in
        none|"") ;;
        delay:*)
            echo "sleep ${w#delay:}"
            ;;
        port:*)
            local rest="${w#port:}"
            local host="${rest%%:*}"
            local port="${rest#*:}"
            cat <<EOF
echo '⏳ waiting for tcp $host:$port ...'
until nc -z '$host' '$port' 2>/dev/null; do sleep 0.5; done
echo '✓ $host:$port open'
EOF
            ;;
        http:*)
            local url="${w#http:}"
            cat <<EOF
echo '⏳ waiting for http $url ...'
until curl -sf -o /dev/null --max-time 1 '$url'; do sleep 0.5; done
echo '✓ $url ready'
EOF
            ;;
        rpc:*)
            local url="${w#rpc:}"
            cat <<EOF
echo '⏳ waiting for solana RPC $url ...'
until curl -s -o /dev/null --max-time 1 -X POST -H 'content-type: application/json' \\
    -d '{"jsonrpc":"2.0","id":1,"method":"getHealth"}' '$url'; do sleep 0.5; done
echo '✓ RPC up'
EOF
            ;;
        file:*)
            local path="${w#file:}"
            cat <<EOF
echo '⏳ waiting for file $path ...'
until [ -s '$path' ]; do sleep 0.5; done
echo '✓ $path present'
EOF
            ;;
        cmd:*)
            local sh="${w#cmd:}"
            cat <<EOF
echo '⏳ waiting for: $sh'
until $sh >/dev/null 2>&1; do sleep 0.5; done
echo '✓ ready'
EOF
            ;;
        *)
            echo "echo 'unknown wait directive: $w' >&2; exit 1"
            ;;
    esac
}

for i in $(seq 0 $((N-1))); do
    NAME="${PANE_NAMES[$i]}"
    CWD="${PANE_CWDS[$i]}"
    CMD="${PANE_CMDS[$i]}"
    WAIT="${PANE_WAITS[$i]}"
    SCRIPT="$WORK/p${i}.sh"
    {
        echo '#!/usr/bin/env bash'
        echo 'set -uo pipefail'
        echo "echo '── pane $((i+1)): $NAME ──'"
        emit_wait "$WAIT"
        echo "cd '$CWD'"
        if [ "$INTERACTIVE_LAST_PANE" = "true" ] && [ "$i" -eq "$((N-1))" ]; then
            echo "echo"
            echo "echo '════════════════════════════════════════════════════════════'"
            echo "echo '  Press ENTER to run: $NAME'"
            echo "echo '  (Focus this pane: Ctrl+P then arrow keys)'"
            echo "echo '════════════════════════════════════════════════════════════'"
            echo "read -r"
        fi
        echo "$CMD"
        echo 'status=$?'
        echo "echo; echo \"── $NAME exited (status \$status) ──\""
        echo "read -p 'press enter to close this pane '"
    } > "$SCRIPT"
    chmod +x "$SCRIPT"
done

# ─── KDL layout ─────────────────────────────────────────────────────────────
LAYOUT_FILE="$WORK/layout.kdl"
{
    echo 'layout {'
    echo "    tab name=\"$TAB_NAME\" {"
    case "$LAYOUT" in
        grid)
            if [ "$N" -ne 4 ]; then
                echo "grid layout requires exactly 4 panes (got $N)" >&2
                exit 1
            fi
            cat <<KDL
        pane split_direction="horizontal" {
            pane split_direction="vertical" {
                pane name="${PANE_NAMES[0]}" command="$WORK/p0.sh"
                pane name="${PANE_NAMES[1]}" command="$WORK/p1.sh"
            }
            pane split_direction="vertical" {
                pane name="${PANE_NAMES[2]}" command="$WORK/p2.sh"
                pane name="${PANE_NAMES[3]}" command="$WORK/p3.sh" focus=true
            }
        }
KDL
            ;;
        rows)
            echo '        pane split_direction="horizontal" {'
            for i in $(seq 0 $((N-1))); do
                FOCUS=""
                [ "$i" -eq "$((N-1))" ] && FOCUS=" focus=true"
                echo "            pane name=\"${PANE_NAMES[$i]}\" command=\"$WORK/p${i}.sh\"$FOCUS"
            done
            echo '        }'
            ;;
        columns)
            echo '        pane split_direction="vertical" {'
            for i in $(seq 0 $((N-1))); do
                FOCUS=""
                [ "$i" -eq "$((N-1))" ] && FOCUS=" focus=true"
                echo "            pane name=\"${PANE_NAMES[$i]}\" command=\"$WORK/p${i}.sh\"$FOCUS"
            done
            echo '        }'
            ;;
        *)
            echo "unknown LAYOUT: $LAYOUT (use grid|rows|columns)" >&2
            exit 1
            ;;
    esac
    echo '    }'
    echo '}'
} > "$LAYOUT_FILE"

echo "✓ launching zellij ($LAYOUT, $N panes)"
exec zellij --layout "$LAYOUT_FILE"

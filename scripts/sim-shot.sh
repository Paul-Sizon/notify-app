#!/usr/bin/env bash
# Capture iOS sim screenshot + upload to litterbox (24h ephemeral).
# Usage: ./scripts/sim-shot.sh <name> [device-id]
# Default device: iPhone 17 Pro (5FC1A093-A9DC-4C6E-B868-5DE5A833E881).
set -euo pipefail
NAME="${1:-shot-$(date +%s)}"
DEV="${2:-5FC1A093-A9DC-4C6E-B868-5DE5A833E881}"
PNG="/tmp/notify-${NAME}.png"
xcrun simctl io "$DEV" screenshot "$PNG" >/dev/null 2>&1
URL=$(curl --max-time 30 --silent \
    -F "reqtype=fileupload" -F "time=24h" \
    -F "fileToUpload=@${PNG}" \
    https://litterbox.catbox.moe/resources/internals/api.php)
echo "${PNG} -> ${URL}"

#!/usr/bin/env bash
# Tap iOS sim at iOS-coord (points). 17 Pro = 402 x 874.
# Usage: sim-tap.sh <ios_x_pt> <ios_y_pt>
# Auto-fetches sim window pos from System Events, computes letterbox.
set -euo pipefail
IOS_X="$1"; IOS_Y="$2"
DEV_W=402; DEV_H=874

# Make sim frontmost — needed for cliclick to land in iOS view.
osascript -e 'tell application "Simulator" to activate' >/dev/null
sleep 0.25

GEOM=$(osascript -e 'tell application "System Events" to tell process "Simulator" to set p to position of window 1' \
                  -e 'tell application "System Events" to tell process "Simulator" to set s to size of window 1' \
                  -e 'return (item 1 of p as text) & " " & (item 2 of p as text) & " " & (item 1 of s as text) & " " & (item 2 of s as text)')
read WX WY WW WH <<<"$GEOM"

CHROME=28
CW=$WW
CH=$((WH - CHROME))

# Letterbox: scale device into content area preserving aspect.
SCALE=$(python3 -c "print(min($CW/$DEV_W, $CH/$DEV_H))")
DRAWN_W=$(python3 -c "print($DEV_W * $SCALE)")
DRAWN_H=$(python3 -c "print($DEV_H * $SCALE)")
PAD_X=$(python3 -c "print(($CW - $DRAWN_W) / 2)")
PAD_Y=$(python3 -c "print(($CH - $DRAWN_H) / 2)")

MX=$(python3 -c "print(round($WX + $PAD_X + $IOS_X * $SCALE))")
MY=$(python3 -c "print(round($WY + $CHROME + $PAD_Y + $IOS_Y * $SCALE))")

echo "ios=($IOS_X,$IOS_Y) -> mac=($MX,$MY)"
cliclick "c:$MX,$MY"

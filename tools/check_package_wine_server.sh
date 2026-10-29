#!/bin/bash
# Inside the existing isolated Wine image; X11 socket shared with the GUI driver.
set -euo pipefail
test -e /.dockerenv
version=${1:?full version}
[[ "$version" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
export DISPLAY=:99 WINEPREFIX=/wine WINEDEBUG=-all LANG=C.utf8 LC_ALL=C.utf8
export ALSA_CONFIG_PATH=/tmp/wl-package-alsa.conf
printf 'pcm.!default { type null }\n' > "$ALSA_CONFIG_PATH"
evidence=/evidence/full-local-windows-wine
mkdir -p "$evidence"
game_pid= xvfb_pid=
cleanup() {
    [ -z "$game_pid" ] || kill "$game_pid" 2>/dev/null || true
    wineserver -k 2>/dev/null || true
    if [ -n "$xvfb_pid" ]; then kill "$xvfb_pid" 2>/dev/null || true; wait "$xvfb_pid" 2>/dev/null || true; fi
}
trap cleanup EXIT INT TERM
Xvfb :99 -screen 0 2560x1600x24 -ac -nolisten tcp > "$evidence/xvfb.log" 2>&1 &
xvfb_pid=$!
for n in $(seq 1 50); do [ ! -S /tmp/.X11-unix/X99 ] || break; sleep .1; done
wine /pkg/bin/wl-setup.exe -version > "$evidence/wl-setup-version.txt" 2> "$evidence/wine-init.log"
wine /pkg/bin/wasteland.exe -version > "$evidence/wasteland-version.txt" 2>> "$evidence/wine-init.log"
wine cmd /c 'Z:\pkg\開始遊戲.bat' > "$evidence/game.log" 2>&1 &
game_pid=$!
rc=0
wait "$game_pid" || rc=$?
game_pid=
printf '%s\n' "$rc" > "$evidence/exit-code.txt"
exit "$rc"

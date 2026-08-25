#!/usr/bin/env bash
# 容器內錄製正式遊戲視窗；由 record_gameplay.sh 啟動。
set -euo pipefail

OUT=${1:-/out}
mkdir -p "$OUT"
printf 'pcm.!default { type null }\nctl.!default { type null }\n' > /tmp/asound.conf
export DISPLAY=:99 ALSA_CONFIG_PATH=/tmp/asound.conf LIBGL_ALWAYS_SOFTWARE=1

game= keys=
Xvfb :99 -screen 0 1280x720x24 -nolisten tcp >/tmp/xvfb.log 2>&1 & xvfb=$!
cleanup() {
    test -z "$keys" || kill "$keys" 2>/dev/null || true
    test -z "$game" || kill "$game" 2>/dev/null || true
    kill "$xvfb" 2>/dev/null || true
}
trap cleanup EXIT

go build -o /tmp/wasteland ./cmd/wasteland

capture() {
    mode=$1 out=$2
    /tmp/wasteland \
        -rom workplace/orig/wastland \
        -image workplace/analysis/unpacked/wl.merged.exe \
        -mode play -skip-title -scale 1 \
        -lang translations/zh-Hant.cat -font workplace/eten \
        -music '' -quicksave '' -art-root artpacks -art-mode "$mode" \
        >"/tmp/game-$mode.log" 2>&1 &
    game=$!
    win=
    for _ in $(seq 1 60); do
        win=$(xdotool search --name 'Wasteland' 2>/dev/null | head -1 || true)
        test -n "$win" && break
        sleep 0.2
    done
    if test -z "$win"; then
        sed -n '1,180p' "/tmp/game-$mode.log" >&2
        return 1
    fi

    # 正常玩家輸入：行走 → 開設定 → 顯示／切換配樂版本 → 關面板 → 再行走。
    (
        sleep 2
        xdotool key --window "$win" j
        sleep 1
        xdotool key --window "$win" l
        sleep 1
        xdotool key --window "$win" j
        sleep 1
        xdotool key --window "$win" F2
        sleep 2
        xdotool key --window "$win" b
        sleep 2
        xdotool key --window "$win" Escape
        sleep 1
        xdotool key --window "$win" i
    ) & keys=$!

    ffmpeg -hide_banner -loglevel error -y \
        -f x11grab -framerate 30 -video_size 1280x720 -i :99 -t 12 \
        -c:v libx264 -preset veryfast -crf 20 -pix_fmt yuv420p "$OUT/$out"
    wait "$keys" || true
    keys=
    kill "$game" 2>/dev/null || true
    wait "$game" 2>/dev/null || true
    game=
}

capture original original-live.mp4
capture reimagined reimagined-live.mp4

for clip in original-live reimagined-live; do
    dur=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$OUT/$clip.mp4")
    sha=$(sha256sum "$OUT/$clip.mp4" | cut -d' ' -f1)
    printf '{"mode":"%s","source":"cmd/wasteland via Xvfb+x11grab","input":"j,l,j,F2,b,Escape,i","duration":%s,"sha256":"%s"}\n' \
        "${clip%-live}" "$dur" "$sha" > "$OUT/$clip.json"
done


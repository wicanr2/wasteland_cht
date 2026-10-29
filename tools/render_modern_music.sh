#!/usr/bin/env bash
# 以 FluidR3 GM 產生使用者定案的振奮版十首配樂。先驗證候選，再安裝主混音。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUT="${1:-$ROOT/workplace/music/modern}"
IMAGE="${WASTELAND_MEDIA_IMAGE:-wasteland-media:bookworm-modern-r1}"
OUT_PARENT="$(dirname "$OUT")"
OUT_NAME="$(basename "$OUT")"

mkdir -p "$OUT_PARENT"
test -d "$ROOT"
test -d "$OUT_PARENT"
docker run --rm --network none --memory 3g --cpus 2 --pids-limit 512 \
    --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" -e XDG_CONFIG_HOME=/tmp/wasteland-music-config \
    -e PYTHONDONTWRITEBYTECODE=1 \
    -v "$ROOT:/src:ro" -v "$OUT_PARENT:/music" -w /src "$IMAGE" \
    timeout 900 python3 tools/render_modern_music.py /music/modern-uplift-candidate \
        --install-to "/music/$OUT_NAME"

echo "現代配樂在 $OUT"

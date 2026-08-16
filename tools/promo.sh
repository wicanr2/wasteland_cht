#!/usr/bin/env bash
# 合成推廣片。
#
#	tools/promo.sh [截圖目錄] [輸出目錄]
#
# 預設讀 workplace/promo/shots/、寫 workplace/promo/out/。
# 需要先跑過 `tools/render_music.sh`（配樂）與 `docker build -f docker/media.Dockerfile`。
#
# 截圖從哪來：`cmd/wl-shot` 無頭產生（見 `tools/promo/shots.sh`）。
#
# ⚠ **產物不進版控**：影片裡有原版的畫面與 MT-32 算出來的聲音，
# 與原版資料、倚天字型同一個政策。腳本進版控、成品不進。
#
# 邊界寫在腳本本體：--rm、--network none、限 2 核、截圖與配樂唯讀掛載、
# 退出前把產物 chown 回使用者。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SHOTS="${1:-$ROOT/workplace/promo/shots}"
OUT="${2:-$ROOT/workplace/promo/out}"
MUSIC="$ROOT/workplace/music"
IMAGE=wasteland-media:latest

for d in "$SHOTS" "$MUSIC"; do
    if [ ! -d "$d" ]; then
        echo "找不到 $d" >&2
        exit 1
    fi
done
if [ ! -f "$MUSIC/theme.ogg" ]; then
    echo "沒有配樂：先跑 tools/render_music.sh" >&2
    exit 1
fi
if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
    echo "先建工具 image：docker build -f docker/media.Dockerfile -t $IMAGE ." >&2
    exit 1
fi

mkdir -p "$OUT"

docker run --rm --cpus=2 --network none \
    --log-opt max-size=10m --log-opt max-file=3 \
    -v "$SHOTS:/shots:ro" -v "$MUSIC:/music:ro" -v "$OUT:/out" \
    -v "$ROOT/tools/promo/theme.sh:/theme.sh:ro" \
    -v "$ROOT/tools/promo/make_promo.sh:/make.sh:ro" \
    -e HOST_UID="$(id -u)" -e HOST_GID="$(id -g)" \
    "$IMAGE" bash /make.sh

echo "---"
echo "產出在 $OUT（不入版控，見檔頭）"

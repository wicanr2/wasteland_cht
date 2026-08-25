#!/usr/bin/env bash
# 只負責啟動一次性 Docker；遊戲、Xvfb、鍵盤事件與 FFmpeg 全在容器內。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT="${1:-$ROOT/workplace/promo/gameplay}"
IMAGE=wasteland-go:1.24-x11-record-r1
test -d "$OUT" || { echo "請先建立輸出目錄：$OUT" >&2; exit 1; }

docker run --rm --network none --memory 4g --cpus 2 --pids-limit 512 \
    -u "$(id -u):$(id -g)" \
    -e HOME=/tmp/wl-home -e GOPATH=/go -e GOCACHE=/tmp/go-build \
    -v "$ROOT:/src:ro" -v /home/anr2/go/pkg/mod:/go/pkg/mod:ro \
    -v "$ROOT/workplace/orig/wastland:/src/workplace/orig/wastland:ro" \
    -v "$ROOT/workplace/eten:/src/workplace/eten:ro" \
    -v "$OUT:/out" -w /src "$IMAGE" \
    bash /src/tools/promo/record_gameplay_container.sh /out


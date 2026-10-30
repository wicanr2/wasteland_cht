#!/usr/bin/env bash
# 從已驗證的 V1／V2 本機包製作 README 圖庫，不重跑舊版 wl-shot。
# 截圖逐位元組保留；素材總覽從最終 v2 ZIP 讀取，核對 manifest 與 SHA-256。
# 展示圖進 docs/images/v1.0.0、v2.0.0，原始圖包與 dist-all 仍留本機。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUT="$ROOT/docs/images"
IMAGE="san1-go:1.26.7-ebiten2.9.9-r1"
test -d "$OUT"
test -f "$ROOT/tools/make_readme_gallery.py"
test -d "$ROOT/dist-all/v.1.0.0-20261029"
test -d "$ROOT/dist-all/v.2.0.0-20261029"
docker image inspect "$IMAGE" >/dev/null
timeout 180s docker run --rm --network none --user "$(id -u):$(id -g)" \
    --memory 2g --cpus 2 --pids-limit 128 \
    --log-opt max-size=10m --log-opt max-file=1 \
    -v "$ROOT:/src:ro" -v "$OUT:/src/docs/images:rw" -w /src \
    "$IMAGE" timeout 150s python3 tools/make_readme_gallery.py

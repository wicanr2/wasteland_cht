#!/usr/bin/env bash
# Versioned actual-package capture and local trailer. Host only orchestrates Docker.
set -euo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
VERSION=${1:?完整版本}
[[ "$VERSION" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
IMAGE=san1-go:1.26.7-ebiten2.9.9-r1
if [ "${2:-}" != --render-only ]; then
    bash "$ROOT/tools/promo/record_packaged.sh" "$VERSION"
fi
RAW="$ROOT/workplace/promo/$VERSION/gameplay"
VERSION_ROOT="$ROOT/dist-all/$VERSION"
test -d "$RAW" && test -f "$RAW/capture.json" && test -d "$VERSION_ROOT"
timeout 30 docker run --rm --name wasteland-promo-output-prepare --network none --memory 256m --cpus 1 --pids-limit 64 \
    --log-opt max-size=3m --log-opt max-file=1 --user "$(id -u):$(id -g)" \
    -v "$VERSION_ROOT:/version" --entrypoint python3 "$IMAGE" -c \
    'from pathlib import Path; p=Path("/version/promo"); p.mkdir(exist_ok=True); assert p.stat().st_uid!=0'
OUT="$VERSION_ROOT/promo"
test -d "$OUT"
timeout 600 docker run --rm --name wasteland-promo-render --network none --memory 2g --cpus 2 --pids-limit 256 \
    --log-opt max-size=3m --log-opt max-file=1 --user "$(id -u):$(id -g)" \
    -v "$ROOT:/src:ro" -v "$RAW:/raw:ro" -v "$OUT:/out" -w /src --entrypoint python3 "$IMAGE" \
    tools/promo/render_packaged.py "$VERSION"

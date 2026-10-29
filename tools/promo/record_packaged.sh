#!/usr/bin/env bash
# Host orchestration only. All recording work runs in the existing Go/X11 image.
set -euo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
VERSION=${1:?完整版本}
[[ "$VERSION" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
IMAGE=san1-go:1.26.7-ebiten2.9.9-r1
test -f "$ROOT/dist-all/$VERSION/full-local/linux-x64/wasteland-cht-$VERSION-linux-x64-full-local.AppImage"
test -d "$ROOT/workplace/promo"
timeout 30 docker run --rm --name wasteland-promo-prepare --network none --memory 256m --cpus 1 --pids-limit 64 \
  --log-opt max-size=3m --log-opt max-file=1 --user "$(id -u):$(id -g)" \
  -v "$ROOT/workplace/promo:/work" --entrypoint python3 "$IMAGE" -c \
  'from pathlib import Path; import sys; p=Path("/work")/sys.argv[1]/"gameplay"; p.mkdir(parents=True,exist_ok=True); assert not (p/"capture.json").exists(),"Refusing to overwrite verified capture"' "$VERSION"
OUT="$ROOT/workplace/promo/$VERSION/gameplay"
test -d "$OUT"
timeout 600 docker run --rm --name wasteland-promo-record --network none --memory 3g --cpus 2 --pids-limit 384 \
  --log-opt max-size=3m --log-opt max-file=1 --user "$(id -u):$(id -g)" \
  -v "$ROOT:/src:ro" -v "$OUT:/out" -w /src --entrypoint bash "$IMAGE" \
  tools/promo/record_packaged_container.sh "$VERSION"

#!/usr/bin/env bash
# 六個版本化封包；不覆寫既有版本，不自動發布或收錄歷史影片。
set -euo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
VERSION=${1:-}
[[ "$VERSION" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]] || { echo '必須指定 v.<主>.<次>.<修訂>-YYYYMMDD' >&2; exit 2; }
for mode in patch full-local; do
    for platform in linux-x64 windows-x64 macos-universal; do
        "$ROOT/tools/package.sh" "$VERSION" "$mode" "$platform"
    done
done
PACK_IMAGE=${WL_PKG_IMAGE:-san1-go:1.26.7-ebiten2.9.9-r1}
GO_IMAGE=${WL_GO_IMAGE:-san1-go:1.26.7-ebiten2.9.9-r1}
MAC_IMAGE=${WL_OSXCROSS_IMAGE:-san1-macos:1.26.7-sdk15.5-r1}
test -d "$ROOT/dist-all/$VERSION"
timeout 240s docker run --rm --network none --memory 2g --cpus 2 --pids-limit 128 \
    --log-opt max-size=10m --log-opt max-file=3 --user "$(id -u):$(id -g)" \
    -e WL_GO_IMAGE_ID="$(docker image inspect --format '{{.Id}}' "$GO_IMAGE")" \
    -e WL_OSXCROSS_IMAGE_ID="$(docker image inspect --format '{{.Id}}' "$MAC_IMAGE")" \
    -e WL_PKG_IMAGE_ID="$(docker image inspect --format '{{.Id}}' "$PACK_IMAGE")" \
    -v "$ROOT:/src:ro" -v "$ROOT/dist-all/$VERSION:/src/dist-all/$VERSION" \
    -w /src --entrypoint python3 "$PACK_IMAGE" tools/finalize_dist.py "$VERSION"

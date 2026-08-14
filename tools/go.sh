#!/usr/bin/env bash
# Go 的唯一入口。編譯與測試一律走 docker（CLAUDE.md §6），不污染系統環境。
#
#   tools/go.sh test ./...
#   tools/go.sh build ./cmd/wasteland
#   tools/go.sh vet ./...
#
# 邊界寫在腳本本體，不靠呼叫者自律：
#   --rm            用完就丟，不留 container
#   --network none  建置不連網（相依一律 vendor 或 stdlib）
#   log rotation    daemon 預設的 json-file 沒有 rotation
#   唯讀掛載        只有 build cache 與工作區可寫
#   退出前 chown    產物不留 root 擁有的檔案

set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${WL_GO_IMAGE:-golang:1.24-bookworm}"
CACHE="$REPO/workplace/gocache"

mkdir -p "$CACHE/build" "$CACHE/mod"

exec docker run --rm \
  --log-opt max-size=10m --log-opt max-file=3 \
  --network none \
  --cpus "${WL_GO_CPUS:-4}" --memory "${WL_GO_MEM:-4g}" --pids-limit 512 \
  -v "$REPO:/src" \
  -v "$CACHE/build:/gocache" \
  -v "$CACHE/mod:/gomod" \
  -e GOCACHE=/gocache -e GOMODCACHE=/gomod -e GOFLAGS=-mod=mod \
  -e HOME=/tmp \
  -w /src \
  --entrypoint /bin/sh \
  "$IMAGE" -c "go $* ; rc=\$? ; chown -R $(id -u):$(id -g) /gocache /gomod /src 2>/dev/null || true ; exit \$rc"

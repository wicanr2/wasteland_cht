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
# Ebiten 的視窗層走 cgo，需要 X11／GL 標頭，所以用本專案的 image
# （建置來源在 docker/wasteland-go.Dockerfile，不是臨時裝套件）。
IMAGE="${WL_GO_IMAGE:-wasteland-go:1.24-x11}"
CACHE="$REPO/workplace/gocache"

mkdir -p "$CACHE/build" "$CACHE/mod"

# 相依從**本機模組快取的唯讀掛載**取得，不開網路（使用者 2026-08-15 定案）。
# GOPROXY 指到那份快取的 download 目錄——file proxy 只讀不寫，
# 所以共用的 ~/go/pkg/mod 不會被這個容器動到（唯讀掛載是第二道保險）。
HOSTMOD="${WL_GO_HOSTMOD:-$HOME/go/pkg/mod}"
PROXY_ARGS=()
if [ -d "$HOSTMOD/cache/download" ]; then
  # GONOSUMCHECK：離線就查不到 sum.golang.org，改用 go.sum 自己驗
  # （模組雜湊仍然會比對，只是不去問公開的 checksum database）。
  PROXY_ARGS=(
    -v "$HOSTMOD/cache/download:/hostmod:ro"
    -e GOPROXY=file:///hostmod
    -e GOFLAGS="-mod=mod -buildvcs=false"
    -e GONOSUMDB='*'
    -e GONOSUMCHECK=1
    -e GOSUMDB=off
  )
else
  PROXY_ARGS=(-e GOPROXY=off -e GOFLAGS=-buildvcs=false)
fi

exec docker run --rm \
  --log-opt max-size=10m --log-opt max-file=3 \
  --network none \
  "${PROXY_ARGS[@]}" \
  --cpus "${WL_GO_CPUS:-4}" --memory "${WL_GO_MEM:-4g}" --pids-limit 512 \
  -v "$REPO:/src" \
  -v "$CACHE/build:/gocache" \
  -v "$CACHE/mod:/gomod" \
  -e GOCACHE=/gocache -e GOMODCACHE=/gomod \
  -e HOME=/tmp \
  -w /src \
  --entrypoint /bin/sh \
  "$IMAGE" -c "go $* ; rc=\$? ; chown -R $(id -u):$(id -g) /gocache /gomod /src 2>/dev/null || true ; exit \$rc"

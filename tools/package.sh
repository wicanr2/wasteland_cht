#!/usr/bin/env bash
# 主機只組織 Docker 呼叫；建置、封裝、雜湊與驗證均在容器內。
set -euo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
VERSION=${1:-}; MODE=${2:-}; PLATFORM=${3:-}
[[ "$VERSION" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]] || { echo '用法：tools/package.sh v.1.0.0-YYYYMMDD <patch|full-local> <平台>' >&2; exit 2; }
case "$MODE" in patch|full-local) ;; *) echo '模式必須為 patch 或 full-local' >&2; exit 2 ;; esac
case "$PLATFORM" in linux-x64|windows-x64|macos-universal) ;; *) echo '未知平台' >&2; exit 2 ;; esac

BUILD_IMAGE=${WL_GO_IMAGE:-san1-go:1.26.7-ebiten2.9.9-r1}
[ "$PLATFORM" != macos-universal ] || BUILD_IMAGE=${WL_OSXCROSS_IMAGE:-san1-macos:1.26.7-sdk15.5-r1}
PACK_IMAGE=${WL_PKG_IMAGE:-san1-go:1.26.7-ebiten2.9.9-r1}
DATA_DIR=${WL_DATA:-$ROOT/workplace/orig/wastland}
FONT_DIR=${WL_ETEN:-$ROOT/workplace/eten}
MUSIC_DIR=${WL_MUSIC:-$ROOT/workplace/music}
INPUT_ARGS=(); INPUT=; FONT_INPUT=; MUSIC_INPUT=
READONLY_INPUTS=()
for input_dir in workplace/orig workplace/art-reference artwork/source; do
    [ ! -d "$ROOT/$input_dir" ] || READONLY_INPUTS+=(-v "$ROOT/$input_dir:/src/$input_dir:ro")
done
if [ "$MODE" = full-local ]; then
    [ -f "$DATA_DIR/wl.exe" ] && [ -d "$FONT_DIR" ] && [ -d "$MUSIC_DIR/modern" ] && [ -d "$MUSIC_DIR/retro" ] || { echo '缺少本機原版、字型或雙配樂來源' >&2; exit 1; }
    INPUT_ARGS=(-v "$DATA_DIR:/input:ro" -v "$FONT_DIR:/font-input:ro" -v "$MUSIC_DIR:/music-input:ro")
    INPUT=/input; FONT_INPUT=/font-input; MUSIC_INPUT=/music-input
fi
RUNTIME="$ROOT/workplace/appimage/runtime-x86_64"
[ "$PLATFORM" != linux-x64 ] || [ -f "$RUNTIME" ] || { echo '缺少已封存的 AppImage runtime，見 docs/packaging.md' >&2; exit 1; }
PKG="wasteland-cht-${VERSION}-${PLATFORM}-${MODE}"
SOURCE_COMMIT=$(git -C "$ROOT" rev-parse HEAD)

# 先在既有工作樹掛載中建立明確輸出目錄，避免 dockerd 以 root 建立缺少的來源。
timeout 30s docker run --rm --network none --memory 256m --cpus 1 --pids-limit 64 \
    --log-opt max-size=10m --log-opt max-file=3 --user "$(id -u):$(id -g)" \
    -v "$ROOT:/src" "${READONLY_INPUTS[@]}" -w /src --entrypoint /bin/bash "$PACK_IMAGE" -c '
        set -eu
        out="/src/dist-all/$1/$2/$3"
        for suffix in AppImage zip; do test ! -e "$out/$4.$suffix" || { echo "拒絕覆寫 $out/$4.$suffix" >&2; exit 1; }; done
        for p in /src/dist-all /src/dist-build; do [ ! -e "$p" ] || stat -c "%u:%g %n" "$p"; done
        mkdir -p "/src/dist-build/$1/$2/$3" "/src/dist-all/$1/$2" /src/workplace/gocache/build /src/workplace/gocache/mod
    ' _ "$VERSION" "$MODE" "$PLATFORM" "$PKG"

run() {
    local image=$1; shift
    timeout "${WL_PACKAGE_TIMEOUT:-900}s" docker run --rm --network none \
        --log-opt max-size=10m --log-opt max-file=3 \
        --cpus "${WL_GO_CPUS:-2}" --memory "${WL_GO_MEM:-3g}" --pids-limit 512 \
        --user "$(id -u):$(id -g)" -e HOME=/tmp \
        -e GOCACHE=/gocache -e GOMODCACHE=/gomod -e GOPROXY=off -e GOFLAGS=-buildvcs=false \
        -e WL_RELEASE_VERSION="$VERSION" -e WL_BUILD_IMAGE="$BUILD_IMAGE" -e WL_SOURCE_COMMIT="$SOURCE_COMMIT" \
        -v "$ROOT:/src:ro" -v "$ROOT/workplace/gocache/build:/gocache" -v "$ROOT/workplace/gocache/mod:/gomod" \
        -v "$ROOT/dist-build/$VERSION/$MODE/$PLATFORM:/stage" \
        -v "$ROOT/dist-all/$VERSION/$MODE:/out" "${INPUT_ARGS[@]}" \
        -w /src --entrypoint /bin/bash "$image" "$@"
}
run "$BUILD_IMAGE" /src/tools/package_container.sh "$MODE" "$PLATFORM" /stage "$INPUT" "$FONT_INPUT" "$MUSIC_INPUT"
run "$PACK_IMAGE" /src/tools/pack_wrap.sh "$MODE" "$PLATFORM" "/stage/.stage/$PKG" /out \
    "$([ "$PLATFORM" != linux-x64 ] || printf /src/workplace/appimage/runtime-x86_64)"

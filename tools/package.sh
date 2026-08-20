#!/usr/bin/env bash
# 三平台封裝。
#
#	tools/package.sh <public|local-full> <linux-x64|windows-x64|macos-universal>
#
# 兩份的差別只有一件事：**要不要把原版素材放進去**。
#
#	public      引擎、翻譯、文件、`wl-setup`。玩家自備原版資料與倚天字型。
#	local-full  上面全部 ＋ 原版資料 ＋ 倚天字型 ＋ 背景音樂 ＋ 合成映像。
#	            **一律不可散布**（`CLAUDE.md` §7），只能自己留檔。
#
# 兩步走：先在編譯用的 image 裡排出「舞台目錄」，再在封裝用的 image 裡封成
# AppImage／zip。分開的理由是 macOS 要 osxcross 那個 image，而 AppImage 要的
# `mksquashfs` 只在 `wasteland-pkg` 裡，兩者不在同一個 image。
#
# 邊界寫在腳本本體：--rm、--network none、限 2 核／4 GB、原版素材唯讀掛載、
# 產物 chown 回使用者（容器內就用呼叫者的 uid 跑）。
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
MODE=${1:-}; PLATFORM=${2:-}
case "$MODE" in public|local-full) ;; *) echo "模式要是 public 或 local-full" >&2; exit 2 ;; esac
case "$PLATFORM" in linux-x64|windows-x64|macos-universal) ;; *) echo "平台要是 linux-x64／windows-x64／macos-universal" >&2; exit 2 ;; esac

DATA_DIR=${WL_DATA:-$ROOT/workplace/orig/wastland}
FONT_DIR=${WL_ETEN:-$ROOT/workplace/eten}
MUSIC_DIR=${WL_MUSIC:-$ROOT/workplace/music}

# 產物一律落在被 gitignore 的 `dist-all/`——**原版素材絕不能因為打包而進版控**。
# 所有交付物集中在同一棵樹底下（使用者定案 2026-08-20）。
OUT_ROOT="$ROOT/dist-all/$MODE"
mkdir -p "$OUT_ROOT"
git -C "$ROOT" check-ignore -q "$OUT_ROOT" || { echo "輸出路徑沒有被 gitignore：$OUT_ROOT" >&2; exit 1; }

INPUT_ARGS=(); INPUT=""; FONT_INPUT=""; MUSIC_INPUT=""
if [ "$MODE" = local-full ]; then
    [ -f "$DATA_DIR/wl.exe" ] || { echo "local-full 需要含 wl.exe 的原版資料目錄（$DATA_DIR）" >&2; exit 1; }
    INPUT_ARGS+=(-v "$DATA_DIR:/input:ro"); INPUT=/input
    if [ -d "$FONT_DIR" ]; then
        INPUT_ARGS+=(-v "$FONT_DIR:/font-input:ro"); FONT_INPUT=/font-input
    else
        echo "   略過倚天字型（找不到 $FONT_DIR）——這個包只跑得出英文"
    fi
    if compgen -G "$MUSIC_DIR/*.ogg" >/dev/null; then
        INPUT_ARGS+=(-v "$MUSIC_DIR:/music-input:ro"); MUSIC_INPUT=/music-input
    else
        echo "   略過背景音樂（$MUSIC_DIR 沒有 ogg）"
    fi
fi

if [ "$PLATFORM" = macos-universal ]; then
    BUILD_IMAGE=${WL_OSXCROSS_IMAGE:-wolong-osxcross-go:20260811-event10-r4}
else
    BUILD_IMAGE=${WL_GO_IMAGE:-wasteland-go:1.24-x11}
fi
PACK_IMAGE=${WL_PKG_IMAGE:-wasteland-pkg:1.24-x11}

# AppImage 的 runtime 是上游那顆固定的靜態 ELF。放在被忽略的 workplace/，
# 雜湊寫死在這裡——換了要有人明確改這一行，不會靜悄悄換掉。
RUNTIME="$ROOT/workplace/appimage/runtime-x86_64"
RUNTIME_SHA=1cc49bcf1e2ccd593c379adb17c9f85a36d619088296504de95b1d06215aebbf
RUNTIME_URL=https://github.com/AppImage/type2-runtime/releases/download/continuous/runtime-x86_64
if [ "$PLATFORM" = linux-x64 ]; then
    if [ ! -f "$RUNTIME" ]; then
        mkdir -p "$(dirname "$RUNTIME")"
        echo "[package] 取 AppImage runtime：$RUNTIME_URL"
        curl -sfL -o "$RUNTIME" "$RUNTIME_URL"
    fi
    echo "$RUNTIME_SHA  $RUNTIME" | sha256sum -c - >/dev/null \
        || { echo "AppImage runtime 雜湊不符" >&2; exit 1; }
fi

CACHE="$ROOT/workplace/gocache"
mkdir -p "$CACHE/build" "$CACHE/mod"
HOSTMOD="${WL_GO_HOSTMOD:-$HOME/go/pkg/mod}"
PROXY_ARGS=(-e GOPROXY=off -e GOFLAGS=-buildvcs=false)
if [ -d "$HOSTMOD/cache/download" ]; then
    PROXY_ARGS=(
        -v "$HOSTMOD/cache/download:/hostmod:ro"
        -e GOPROXY=file:///hostmod
        -e GOFLAGS="-mod=mod -buildvcs=false"
        -e GONOSUMDB='*' -e GONOSUMCHECK=1 -e GOSUMDB=off
    )
fi

run() { # $1 image，其餘是指令
    local image=$1; shift
    docker run --rm --network none \
        --log-opt max-size=10m --log-opt max-file=3 \
        --cpus "${WL_GO_CPUS:-2}" --memory "${WL_GO_MEM:-4g}" --pids-limit 512 \
        -u "$(id -u):$(id -g)" -e HOME=/tmp \
        -e GOCACHE=/gocache -e GOMODCACHE=/gomod "${PROXY_ARGS[@]}" \
        -v "$ROOT:/src" -v "$CACHE/build:/gocache" -v "$CACHE/mod:/gomod" \
        "${INPUT_ARGS[@]}" -w /src \
        -e WL_BUILD_IMAGE="$BUILD_IMAGE" \
        "$image" "$@"
}

STAMP=$(git -C "$ROOT" rev-parse --short=12 HEAD)
PKG="wasteland-cht-${PLATFORM}-${MODE}-${STAMP}"
CONTAINER_OUT="/src/dist-all/$MODE"

run "$BUILD_IMAGE" bash /src/tools/package_container.sh \
    "$MODE" "$PLATFORM" "$CONTAINER_OUT" "$INPUT" "$FONT_INPUT" "$MUSIC_INPUT"
run "$PACK_IMAGE" bash /src/tools/pack_wrap.sh \
    "$MODE" "$PLATFORM" "$CONTAINER_OUT/.stage/$PKG" "$CONTAINER_OUT" \
    "$([ "$PLATFORM" = linux-x64 ] && echo /src/workplace/appimage/runtime-x86_64 || echo '')"

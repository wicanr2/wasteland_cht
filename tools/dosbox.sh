#!/usr/bin/env bash
# 在 headless DOSBox 裡跑原版 Wasteland，送鍵、截圖。全程 docker。
#
#   tools/dosbox.sh [timeline] [cycles] [執行檔]
#
# timeline：';' 分隔的步驟（見 docker/dosbox/entrypoint.sh）
#   wait:N / key:KEYSYM / type:STRING / shot:NAME
# 不給就等 6 秒截一張 default.png。
#
# 例：
#   tools/dosbox.sh
#   tools/dosbox.sh "wait:4;shot:01-title;key:Return;wait:2;shot:02-menu"
#
# ⚠ **cycles 一律寫死**（預設 "fixed 3000"）。`auto` 會讓同一串按鍵每次跑到
# 不同的遊戲內時間點，截圖對不起來——可重現性的敵人
# （~/.claude/knowledge-base/retro/dosbox-game-configs.md）。
#
# ⚠ DOSBox 是 **IDA 的驗證工具，不是替代品**（CLAUDE.md §1）。
# 拿它來確認「讀出來的行為對不對」，不要拿它來取代讀程式碼。
#
# 遊戲資料：第一次跑會從 workplace/orig/wastland/ 複製一份**可寫副本**到
# workplace/dosbox/game/（原版目錄唯讀，不可動）。要還原成乾淨初始狀態就
# 手動刪掉 workplace/dosbox/game/ 再跑一次。
#
# 截圖輸出：workplace/dosbox/shots/*.png（gitignore）。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

TIMELINE="${1:-}"
CYCLES="${2:-fixed 3000}"
EXE="${3:-wl}"

IMAGE=wasteland-dosbox:latest
DOCKER_DIR="$ROOT/docker/dosbox"
GAME_DIR="$ROOT/workplace/dosbox/game"
SHOTS_DIR="$ROOT/workplace/dosbox/shots"
ORIG_DIR="$ROOT/workplace/orig/wastland"

need_build=0
if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
    need_build=1
else
    created=$(docker image inspect "$IMAGE" --format '{{.Created}}')
    img_epoch=$(date -d "$created" +%s 2>/dev/null || echo 9999999999)
    src_epoch=$(stat -c %Y "$DOCKER_DIR/Dockerfile" "$DOCKER_DIR/entrypoint.sh" | sort -n | tail -1)
    [[ "$src_epoch" -gt "$img_epoch" ]] && need_build=1
fi
if [[ "$need_build" == 1 ]]; then
    echo "[dosbox.sh] build $IMAGE ..."
    docker build -q -t "$IMAGE" "$DOCKER_DIR" >/dev/null
fi

if [[ ! -d "$ORIG_DIR" ]]; then
    echo "[dosbox.sh] 找不到 $ORIG_DIR —— 原版資料玩家自備（CLAUDE.md §7）" >&2
    exit 1
fi
if [[ ! -d "$GAME_DIR" ]]; then
    echo "[dosbox.sh] 第一次跑：從 $ORIG_DIR 複製可寫副本到 $GAME_DIR"
    mkdir -p "$GAME_DIR"
    cp -r "$ORIG_DIR"/. "$GAME_DIR"/
fi
mkdir -p "$SHOTS_DIR"

# 邊界寫在腳本本體（CLAUDE.md §6）：--rm、限資源、log 上限、原版目錄不掛進去。
docker run --rm \
    --log-opt max-size=10m --log-opt max-file=3 \
    --cpus 2 --memory 1g --pids-limit 256 \
    -v "$GAME_DIR:/game" \
    -v "$SHOTS_DIR:/shots" \
    "$IMAGE" "$EXE" "$TIMELINE" "$CYCLES"

# 容器內是 root，把產出的擁有者換回來。
docker run --rm -v "$SHOTS_DIR:/shots" --entrypoint chown "$IMAGE" \
    -R "$(id -u):$(id -g)" /shots
echo "[dosbox.sh] 截圖在 $SHOTS_DIR"

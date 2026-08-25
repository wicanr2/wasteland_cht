#!/usr/bin/env bash
# 產生 README 用的截圖（無頭，走 cmd/wl-shot）。
#
#	tools/readme_shots.sh
#
# ⚠ **改過畫面就要重跑。** 這幾張是 README 唯一的視覺說明，
# 而截圖過期不會讓任何測試變紅——它只會讓文件描述一個已經不存在的畫面。
#
# 輸出進 `docs/images/`（**進版控**：README 要靠它們，而它們是重製版自己的畫面，
# 不含原版素材以外的東西——原版的圖磚與字模是原版的，這一點與 `dist.sh`
# 的散布規則不同，README 的截圖屬於專案說明的一部分）。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUT="$ROOT/docs/images"

mkdir -p "$OUT"

shot() { # $1 檔名，其餘傳給 wl-shot
    local name="$1"; shift
    (cd "$ROOT" && ./tools/go.sh run ./cmd/wl-shot -mode play "$@" \
        -out "docs/images/$name") | tail -1
}

shot 00-title.png    -title
shot 01-map.png
# 訊息視窗：走一步踩到有敘述的格子。
shot 04-message.png  -map 1 -at 3,6
shot 02-journal.png  -journal 1
shot 03-use.png      -keys U1S
shot 07-question.png -map 1 -at 3,4 -keys i
# 戰鬥：`ENC` 在有敵人的地方直接開打。面板在右上、名單在下半、
# 指令列照留——版面照原版（`docs/re/105` §2）。
shot 08-combat.png   -map 4 -at 18,2 -keys E
# 名單展開（空白鍵 ＝ 原版 ds:46B9h；引號要包進參數，go.sh 的 sh -c 會攤平引號）
shot 09-roster.png   -keys '"' '"'
# 角色畫面：地圖按 1 開第一個人的資料頁（docs/re/131）
shot 10-charview.png -keys 1
shot 05-ending.png   -ending
shot 06-epilogue.png -ending -ending-ticks 300

echo "---"
echo "截圖在 $OUT"

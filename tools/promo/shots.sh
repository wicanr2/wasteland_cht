#!/usr/bin/env bash
# 產生推廣片要用的截圖（無頭，走 cmd/wl-shot）。
#
#	tools/promo/shots.sh [輸出目錄]
#
# 預設寫 workplace/promo/shots/。需要玩家自備的原版資料與倚天字型。
#
# ⚠ 每一張都要**可重現**：座標、地圖、按鍵全部寫死在這裡。
# 有兩張刻意載不到翻譯目錄（`-lang no-such.cat`）當英文對照組——
# `wl-shot` 對載不到翻譯是容忍的（`docs/spec/11` §7），所以這樣拿得到英文畫面。
# 空字串 `-lang ""` 不行：經過 shell 會被吃掉，`-out` 跟著錯位寫到 shot.png。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUT="${1:-$ROOT/workplace/promo/shots}"
mkdir -p "$OUT"

# ⚠ `-out` 要給**相對於 repo 根目錄**的路徑：`tools/go.sh` 在容器裡把 repo
# 掛到 `/src` 並且 `-w /src`，主機的絕對路徑在裡面不存在，
# 症狀是 `no such file or directory` 指著一個明明存在的目錄。
case "$OUT" in
    "$ROOT"/*) REL="${OUT#"$ROOT"/}" ;;
    *) echo "輸出目錄要在 repo 底下（收到 $OUT）" >&2; exit 1 ;;
esac

shot() { # $1 檔名，其餘傳給 wl-shot
    name=$1; shift
    (cd "$ROOT" && ./tools/go.sh run ./cmd/wl-shot -mode play "$@" -out "$REL/$name") | tail -1
}

# 標題畫面與核心遊戲畫面改用正式新版美術包；仍由真正 play.Scene 合成。
shot title.png -title -art-preview reimagined -art-root artpacks -art-width 1280 -art-height 720

# 荒漠地圖：出廠存檔的起點附近
shot 01-map.png -at 55,60 -art-preview reimagined -art-root artpacks -art-width 1280 -art-height 720
shot faithful-map.png -at 55,60 -art-preview faithful-hd -art-root artpacks

# 戰鬥：地圖 4 的 (18,2) 一定開得起來（`internal/play` 的肖像門檻用同一組）
shot combat.png -map 4 -at 18,2 -fight -art-preview reimagined -art-root artpacks -art-width 1280 -art-height 720

# 遊俠中心：從南邊走上去踩進設施格 (55,62)
shot facility.png -at 55,63 -keys i -art-preview faithful-hd -art-root artpacks

# 技能清單：USE → 選第一個人 → S。英文那張只是載不到翻譯目錄
shot 03-use.png   -keys U1S
shot skills-en.png -keys U1S -lang no-such.cat

# F1 說明面板（保留原版模式，作為三模式之一的對照）
shot help.png -fn help

# 角色畫面：地圖按 1 開第一個人（docs/re/131）——裝備／裝填／交給隊友的入口
shot charview.png -keys 1

# 手札、問答、結局：沿用既有的截圖流程
shot 02-journal.png -journal 1
shot 07-question.png -map 1 -at 3,4 -keys i
shot 05-ending.png -ending -ending-ticks 130 -art-preview reimagined -art-root artpacks -art-width 1280 -art-height 720

echo "---"
echo "截圖在 $OUT"

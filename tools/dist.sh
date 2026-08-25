#!/usr/bin/env bash
# 把交付物集中到 dist-all/。
#
#	tools/dist.sh [--skip-packages]
#
# 產出四塊，全部在同一棵樹底下：
#
#	dist-all/public/<平台>/      **可散布**：引擎、翻譯、文件、wl-setup
#	dist-all/local-full/<平台>/  **不可散布**：上面全部 ＋ 原版資料、字型、音樂
#	dist-all/music/              復古／現代兩套配樂與權利紀錄
#	dist-all/promo/              推廣片
#
# 三平台的包由 `tools/package.sh` 產生（做法見 `docs/packaging.md`），
# 這一支只負責**跑滿六個組合並把其他交付物收進同一棵樹**。
# `--skip-packages` 是「包已經跑過了，只想重收音樂／推廣片與校驗碼」。
#
# ⚠ 分兩種模式的理由是硬規則不是潔癖（`CLAUDE.md` §7）：原版執行檔、資料檔、
# 美術、音樂與倚天字型一律不散布，玩家自備合法原版。背景音樂的 ogg 也算——
# 曲子是我們寫的，但波形裡有 Roland MT-32 的 PCM 取樣。

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$ROOT/dist-all"
SKIP=0
[ "${1:-}" = "--skip-packages" ] && SKIP=1

STAMP=$(git -C "$ROOT" log -1 --format=%cs 2>/dev/null || echo unknown)
REV=$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo unknown)

if [ "$SKIP" = 0 ]; then
    echo "== 清掉舊的包"
    rm -rf "$OUT/public" "$OUT/local-full"
    for mode in public local-full; do
        for platform in linux-x64 windows-x64 macos-universal; do
            echo "== $mode $platform"
            "$ROOT/tools/package.sh" "$mode" "$platform"
        done
    done
fi

# ── 配樂單獨放一份 ──────────────────────────────────────────────────────
#
# 完整版的包裡已經有 ogg（給遊戲讀），這一份是**給人拿的**：單獨聽、丟進影片、
# 換音源重算。`.mid` 與 `.ogg` 的散布條件不一樣，所以分開放並在 README 講明。
echo "== 收配樂"
MUSICSRC="${WL_MUSIC:-$ROOT/workplace/music}"
rm -rf "$OUT/music"; mkdir -p "$OUT/music/retro" "$OUT/music/modern"
for variant in retro modern; do
    [ -d "$MUSICSRC/$variant" ] && cp -a "$MUSICSRC/$variant/." "$OUT/music/$variant/"
done
# 舊平面目錄視為 retro，讓尚未重跑新版工具鏈的工作區仍可交付。
find "$MUSICSRC" -maxdepth 1 \( -name '*.ogg' -o -name '*.mid' \) \
    -exec cp {} "$OUT/music/retro/" \; 2>/dev/null || true
if [ -z "$(find "$OUT/music" -name '*.ogg' -print -quit 2>/dev/null)" ]; then
    echo "   略過配樂（$MUSICSRC 是空的，先跑 tools/render_music.sh）"
fi

cat > "$OUT/music/README.md" <<'MUSICNOTE'
# 配樂

**原版《Wasteland》沒有背景音樂**——DOS 版只有九首 PC 喇叭音效，
其中只有音效 4 是旋律，1.8 秒。這裡的十首曲子是重製版自己寫的新作品，
不是抽自原版，也不冒充原版音色。

| 曲名 | 什麼時候放 |
|---|---|
| `theme` | 標題畫面 |
| `desert` | 大地圖，白天 |
| `night` | 大地圖，夜間（6 時前、18 時後）|
| `town` | 一般城鎮與室內 |
| `facility` | 商店、醫生、訓練師、遊俠中心 |
| `sewer` | 下水道與地底隧道 |
| `vegas` | 拉斯維加斯 |
| `base` | 科奇斯基地 |
| `combat` | 戰鬥 |
| `ending` | 結局 |

## 兩個配樂版本

- **`retro/`**：專案自作 MIDI 與 MT-32／CM-32L 算出的 OGG；OGG 波形含
  Roland 韌體取樣，不公開散布。
- **`modern/`**：同一批專案自作主題的現代重新配器；以 MIT 授權的 FluidR3 GM
  產生，目錄內附權利文字與可重現 metadata。

要自己算一份：

```bash
tools/render_music.sh <輸出目錄> <你的 MT-32 ROM 目錄>
tools/render_modern_music.sh <輸出目錄>
```

## 音源為什麼選 MT-32

1988 年 DOS 遊戲的高階音源，年代與機種都對得上。聲部安排照 MT-32 的規矩
不是 General MIDI：channel 1 不用（八個聲部吃 2–9、節奏在 10）、
音色編號走內建表、節奏鍵位是量出來的（`docs/mt32-rhythm-probe.md`）。
MUSICNOTE

# ── 推廣片 ──────────────────────────────────────────────────────────────
echo "== 收推廣片"
rm -rf "$OUT/promo"; mkdir -p "$OUT/promo"
copy_if() { # $1 來源 $2 目的 $3 說明
    if [ -e "$1" ]; then cp -r "$1" "$2"; else echo "   略過 $3（找不到 $1）"; fi
}
copy_if "$ROOT/workplace/promo/out/wasteland-cht-promo.mp4" \
    "$OUT/promo/wasteland-cht-promo.mp4" "推廣片"
copy_if "$ROOT/workplace/promo/out/promo-metadata.json" \
    "$OUT/promo/promo-metadata.json" "推廣片驗收 metadata"
if [ -d "$ROOT/workplace/promo/gameplay" ]; then
    mkdir -p "$OUT/promo/gameplay"
    for stem in original-live reimagined-live; do
        copy_if "$ROOT/workplace/promo/gameplay/$stem.mp4" \
            "$OUT/promo/gameplay/$stem.mp4" "$stem 實機錄影"
        copy_if "$ROOT/workplace/promo/gameplay/$stem.json" \
            "$OUT/promo/gameplay/$stem.json" "$stem 錄影 metadata"
    done
fi
for frame in "$ROOT"/workplace/promo/out/frame-*.png; do
    [ -f "$frame" ] && cp "$frame" "$OUT/promo/"
done
copy_if "$ROOT/docs/promo-video.md" "$OUT/promo/怎麼重跑.md" "推廣片說明"

# ── 檢查：public/ 不准混進不可散布的東西 ─────────────────────────────────
#
# 包自己在 `package_container.sh` 裡已經擋過一次，這裡是**收尾再擋一次**：
# 只看副檔名擋不住 zip 裡面的東西，所以這裡查的是「有沒有人把別的檔案
# 手動丟進 public/」。
echo "== 檢查 public/"
bad=$(find "$OUT/public" -type f ! -name '*.AppImage' ! -name '*.zip' \
    ! -name '*.sha256' -print 2>/dev/null || true)
if [ -n "$bad" ]; then
    echo "   ✗ public/ 底下只該有包本身，卻有：" >&2
    printf '%s\n' "$bad" >&2
    exit 1
fi
echo "   ✓ 只有包本身"

cat > "$OUT/README.md" <<TOP
# dist-all —— 交付物

版本 $REV（$STAMP）。全部交付物集中在這一棵樹底下。

| 目錄 | 可以給別人嗎 | 內容 |
|---|---|---|
| \`public/\` | **可以** | 三平台的包（AppImage／zip／.app）。不含任何原版素材，玩家自備原版與字型；第一次啟動時包裡的 \`wl-setup\` 會產生合成映像 |
| \`local-full/\` | **不可以** | 同樣三個平台，另外含原版資料、倚天字型、背景音樂與合成映像。自己留檔用 |
| \`music/\` | 依子目錄 | \`retro/\` 是專案自作 MIDI 與 MT-32 波形；\`modern/\` 是同主題的現代重新配器，附 FluidR3 GM 權利與重現紀錄 |
| \`promo/\` | 看情況 | 推廣片、原始／新版正式遊戲實機錄影及驗收 metadata；畫面含原版素材，公開前仍須確認權利 |

⚠ 三個「不可以」的理由都不是潔癖：原版資料與倚天字型有授權，
MT-32 算出來的波形裡有 Roland 的 PCM 取樣。譜與程式碼是我們自己的，那些可以給。

\`SHA256SUMS\` 收全部檔案的校驗碼。

重跑：\`tools/dist.sh\`（只想重收音樂與影片就 \`tools/dist.sh --skip-packages\`）
封裝的做法與每個決定的理由：\`docs/packaging.md\`
TOP

# README 也屬於交付物；必須最後才算，否則接著重寫 README 會讓清單當場過期。
echo "== 校驗碼"
( cd "$OUT" && find . -type f ! -name SHA256SUMS -print0 | sort -z |
    xargs -0 sha256sum > SHA256SUMS )

echo
du -sh "$OUT"/* 2>/dev/null || true
echo "---"
echo "產出在 $OUT（不入版控）"

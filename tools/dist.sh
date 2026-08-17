#!/usr/bin/env bash
# 把交付物集中到 dist-all/。
#
#	tools/dist.sh [輸出目錄]      # 預設 dist-all/
#
# 產出四塊：
#
#	dist-all/release/   **可散布**：引擎、翻譯、文件、工具。不含任何原版衍生素材
#	dist-all/local/     **不可散布**：release 的全部 ＋ 原版資料、倚天字型、背景音樂
#	dist-all/music/     配樂單獨一份：`midi/`（可散布）與 `ogg/`（不可散布）
#	dist-all/promo/     推廣片
#
# ⚠ 分兩份的理由是硬規則不是潔癖（`CLAUDE.md` §7）：原版執行檔、資料檔、
# 美術、音樂與倚天字型一律不散布，玩家自備合法原版。背景音樂的 ogg 也算——
# 曲子是我們寫的，但波形裡有 Roland MT-32 的 PCM 取樣。
#
# 腳本結尾會**檢查** release/ 裡沒有混進不該散布的東西，混到就中止。
# 靠自律不夠：兩個目錄長得很像，複製錯一次就散布出去了。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUT="${1:-$ROOT/dist-all}"

REL="$OUT/release"
LOC="$OUT/local"
PROMO="$OUT/promo"
MUSIC="$OUT/music"

STAMP=$(git -C "$ROOT" log -1 --format=%cs 2>/dev/null || echo unknown)
REV=$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo unknown)

echo "== 清掉舊的"
rm -rf "$OUT"
mkdir -p "$REL" "$LOC" "$PROMO" "$MUSIC"

echo "== 編譯（docker）"
"$ROOT/tools/go.sh" build -o dist-build/wasteland ./cmd/wasteland
"$ROOT/tools/go.sh" build -o dist-build/wl-shot ./cmd/wl-shot
"$ROOT/tools/go.sh" build -o dist-build/wl-play ./cmd/wl-play

# ── 可散布的那一份 ───────────────────────────────────────────────────────
echo "== 組 release/"
install -m 755 "$ROOT/dist-build/wasteland" "$REL/wasteland"
install -m 755 "$ROOT/dist-build/wl-shot" "$REL/wl-shot"

mkdir -p "$REL/translations"
cp "$ROOT/translations/zh-Hant.cat" "$ROOT/translations/paragraphs-zh-Hant.cat" "$REL/translations/"
cp "$ROOT/translations/glossary.md" "$ROOT/translations/README.md" "$REL/translations/"
cp -r "$ROOT/translations/zh-Hant" "$REL/translations/zh-Hant"

mkdir -p "$REL/docs"
for d in manual manual-cht paragraphs walkthrough spec; do
    cp -r "$ROOT/docs/$d" "$REL/docs/$d"
done
mkdir -p "$REL/docs/re/generated"
cp "$ROOT/docs/re/generated/paragraph-refs.tsv" "$REL/docs/re/generated/"
cp "$ROOT/README.md" "$ROOT/CONTEXT.md" "$ROOT/WORKLIST.md" "$REL/docs/"

# 玩家要自己從原版產生合成映像，所以工具一起附上（純 stdlib）。
mkdir -p "$REL/tools"
cp "$ROOT/tools/unpack_exepack.py" "$ROOT/tools/apply_overlay.py" "$REL/tools/"
cp "$ROOT/tools/make_music.py" "$ROOT/tools/mt32_probe.py" "$ROOT/tools/render_music.sh" "$REL/tools/"
# 24 點中文字型在倚天光碟上是 ETUNPACK 壓縮的，玩家要自己解。
cp "$ROOT/tools/etunpack.py" "$REL/tools/"
mkdir -p "$REL/docs"
cp "$ROOT/docs/mt32-rhythm-probe.md" "$REL/docs/"

cat > "$REL/setup.sh" <<'SETUP'
#!/usr/bin/env bash
# 從你自己的原版《Wasteland》產生遊戲要用的合成映像。
#
#	./setup.sh <解壓好的原版資料目錄>
#
# 這一步不能替你做：`wl.merged.exe` 是原版執行檔解包之後再疊上 `wla.bin`，
# 屬於原版素材的衍生物，不隨這份散布。工具是純 stdlib 的 Python 3。
set -euo pipefail
SRC="${1:-}"
if [ -z "$SRC" ] || [ ! -f "$SRC/wl.exe" ]; then
    echo "用法：./setup.sh <含 wl.exe 的原版資料目錄>" >&2
    exit 1
fi
mkdir -p build
python3 tools/unpack_exepack.py "$SRC/wl.exe" build/wl.unpacked.exe build/unpack.json
python3 tools/apply_overlay.py build/wl.unpacked.exe "$SRC/wla.bin" \
    build/wl.merged.exe build/overlay.json
echo
echo "好了。開遊戲："
echo "  ./wasteland -rom \"$SRC\" -image build/wl.merged.exe -font <倚天字型目錄>"
SETUP
chmod 755 "$REL/setup.sh"

cat > "$REL/README.md" <<README
# 荒野遊俠（WASTELAND, 1988）繁體中文重製版

版本 $REV（$STAMP）。Linux x86-64。

**這份不含任何原版素材。** 要玩需要你自己準備三樣東西：

1. 合法的原版《Wasteland》資料（解壓成一個目錄，裡面有 \`wl.exe\`、\`game1\`…）
2. 倚天點陣字型目錄（中文顯示要用；沒有就跑英文）
3. Python 3（只在第一次的 \`setup.sh\` 用到，純標準函式庫）

### 字型：優先放 24 點

畫面是 960 × 600（原版的三倍），一個字元格 24 × 24，所以**倚天 24 點的字剛好填滿**：

| 檔案 | 從哪來 |
|---|---|
| \`STDFONT.24\` | 光碟上是 ETUNPACK 壓縮的 \`STD.24M\`（明體），用 \`python3 tools/etunpack.py STD.24M STDFONT.24\` 解開 |
| \`SPCFONT.24\` | 光碟上直接可用（全形標點，**少了它標點會全部缺字**）|
| \`ASCFONT.24\` | 光碟上直接可用（半形英數，**少了它英文會變成粗胖的放大字**）|

只有 15 點（\`STDFONT.15\` 等）也跑得動，字會小一號、畫在格子中央。

## 開始玩

\`\`\`bash
./setup.sh /path/to/wasteland-data      # 產生 build/wl.merged.exe
./wasteland -rom /path/to/wasteland-data -image build/wl.merged.exe \\
            -font /path/to/eten
\`\`\`

按鍵：方向鍵走路，\`U E O D V S R\` 是指令列，\`P\` 開手札。
\`F1\` 說明、\`F2\` 設定、\`F5\`／\`F9\` 快速存讀檔、\`F10\` 離開（會先問，也會先存檔）。
**\`ESC\` 只取消、退一層，任何一層都不會結束遊戲。**

## 背景音樂

原版**沒有**背景音樂（只有九首 PC 喇叭音效）。這份重製版另外寫了十首曲子，
用 Roland MT-32／CM-32L 算成音檔。**算好的音檔不隨這份散布**（波形裡有
Roland 的 PCM 取樣），要有音樂就自備 MT-32 ROM 自己算一次：

\`\`\`bash
docker build -f docker/media.Dockerfile -t wasteland-media .   # 見專案 repo
tools/render_music.sh music ~/mt32
./wasteland ... -music music
\`\`\`

沒有音樂照樣玩得完，遊戲裡按 \`F2\` 也隨時關得掉。

## 裡面有什麼

| 目錄 | 內容 |
|---|---|
| \`translations/\` | 繁中翻譯目錄（4,902 條）與 162 段劇本 |
| \`docs/manual-cht/\` | 軟體世界 1990 年中文說明書逐頁轉錄 |
| \`docs/manual/\` | 官方英文手冊中英對照 |
| \`docs/paragraphs/\` | 段落書整理與翻譯 |
| \`docs/walkthrough/\` | 自建攻略（資料從遊戲檔倒出來的） |
| \`docs/spec/\` | 27 份可實作規格 |
| \`tools/\` | 解包與譜曲工具 |

專案位址：https://github.com/wicanr2/wasteland_cht
README

# ── 本機完整版（不可散布）───────────────────────────────────────────────
echo "== 組 local/"
cp -r "$REL/." "$LOC/"
install -m 755 "$ROOT/dist-build/wl-play" "$LOC/wl-play"
rm -f "$LOC/setup.sh"   # 完整版不用重跑，映像直接附上

copy_if() { # $1 來源 $2 目的 $3 說明
    if [ -e "$1" ]; then
        mkdir -p "$(dirname "$2")"
        cp -r "$1" "$2"
    else
        echo "   略過 $3（找不到 $1）"
    fi
}
copy_if "$ROOT/workplace/orig/wastland" "$LOC/data" "原版資料"
copy_if "$ROOT/workplace/eten" "$LOC/eten" "倚天字型"
copy_if "$ROOT/workplace/analysis/unpacked/wl.merged.exe" "$LOC/build/wl.merged.exe" "合成映像"
mkdir -p "$LOC/music"
for f in "$ROOT"/workplace/music/*.ogg; do
    [ -e "$f" ] && cp "$f" "$LOC/music/"
done

cat > "$LOC/run.sh" <<'RUN'
#!/usr/bin/env bash
# 開遊戲（完整包：資料、字型、音樂都在旁邊）。
cd "$(dirname "$0")"
exec ./wasteland -rom data -image build/wl.merged.exe -font eten \
    -lang translations/zh-Hant.cat \
    -paragraphs translations/paragraphs-zh-Hant.cat \
    -refs docs/re/generated/paragraph-refs.tsv \
    -music music "$@"
RUN
chmod 755 "$LOC/run.sh"

cat > "$LOC/請勿散布.md" <<LOCALNOTE
# 這一份不可以散布

版本 $REV（$STAMP）。

\`data/\`、\`eten/\`、\`build/wl.merged.exe\`、\`music/\` 四樣都是原版素材或其衍生物：

| 目錄 | 是什麼 | 為什麼不能散布 |
|---|---|---|
| \`data/\` | 原版《Wasteland》資料檔 | Interplay／EA 的原版素材 |
| \`build/wl.merged.exe\` | 原版執行檔解包 ＋ 疊 overlay | 同上，是它的衍生物 |
| \`eten/\` | 倚天點陣字型（24 點與 15 點） | 字型本身有授權 |
| \`music/\` | 背景音樂 ogg | 曲子是自己寫的，但波形裡有 Roland MT-32 的 PCM 取樣 |

要給別人的是隔壁的 \`release/\`，那一份不含上面任何一項。

直接玩：\`./run.sh\`
LOCALNOTE

# ── 配樂單獨放一份 ──────────────────────────────────────────────────────
#
# `local/music/` 是給遊戲讀的，這一份是給人拿的（單獨聽、丟進影片、換音源重算）。
# **`.mid` 與 `.ogg` 的散布條件不一樣**，所以分開放並在 README 講明：
# 譜是我們自己寫的，波形裡有 Roland MT-32 的 PCM 取樣。
echo "== 組 music/"
mkdir -p "$MUSIC/ogg" "$MUSIC/midi"
for f in "$ROOT"/workplace/music/*.ogg; do
    [ -e "$f" ] && cp "$f" "$MUSIC/ogg/"
done
for f in "$ROOT"/workplace/music/*.mid; do
    [ -e "$f" ] && cp "$f" "$MUSIC/midi/"
done
if [ -z "$(ls -A "$MUSIC/ogg" 2>/dev/null)" ]; then
    echo "   略過配樂（workplace/music/ 是空的，先跑 tools/render_music.sh）"
fi

cat > "$MUSIC/README.md" <<'MUSICNOTE'
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

## 兩個目錄的散布條件不一樣

- **`midi/`（可散布）**：譜本身，我們自己寫的。由 `tools/make_music.py` 產生，
  純標準函式庫、不含任何第三方素材。想換音源就拿這個去算。
- **`ogg/`（不可散布）**：用 Roland MT-32／CM-32L 算出來的波形。
  曲子是我們的，但**取樣是 Roland 的韌體**——與倚天字型、原版遊戲資料同一個政策。

要自己算一份：

```bash
tools/render_music.sh <輸出目錄> <你的 MT-32 ROM 目錄>
```

## 音源為什麼選 MT-32

1988 年 DOS 遊戲的高階音源，年代與機種都對得上。聲部安排照 MT-32 的規矩
不是 General MIDI：channel 1 不用（八個聲部吃 2–9、節奏在 10）、
音色編號走內建表、節奏鍵位是量出來的（`docs/mt32-rhythm-probe.md`）。
MUSICNOTE

# ── 推廣片 ──────────────────────────────────────────────────────────────
echo "== 組 promo/"
copy_if "$ROOT/workplace/promo/out/wasteland-cht-promo.mp4" \
    "$PROMO/wasteland-cht-promo.mp4" "推廣片"
copy_if "$ROOT/docs/promo-video.md" "$PROMO/怎麼重跑.md" "推廣片說明"

# ── 檢查：release/ 不准混進不可散布的東西 ────────────────────────────────
echo "== 檢查 release/"
bad=0
while IFS= read -r f; do
    echo "   ✗ release/ 裡有不該散布的檔：${f#"$REL/"}"
    bad=1
done < <(find "$REL" \( -name '*.ogg' -o -name '*.15' -o -name 'wl*.exe' \
    -o -name 'game1' -o -name 'game2' -o -name 'allpics*' -o -name 'allhtds*' \
    -o -name '*.pic' -o -name '*.cpa' -o -name '*.fnt' -o -name '*.wlf' \
    -o -name 'paragraphs.txt' \) -type f)
# 原文那一側不進可散布包。
#
# ⚠ **這與「它進不進版控」是兩件事**：repo 收它是為了讓譯文有佐證、
# 讓乾淨 clone 重建得出 `.cat`（使用者定案 2026-08-17）；
# 可散布給玩家的那一包仍然只放譯文與目錄，原文一律不放。
if [ -d "$REL/translations/source" ]; then
    echo "   ✗ release/ 裡有 translations/source（原文那一側，不散布）"
    bad=1
fi
if [ "$bad" != 0 ]; then
    echo "release/ 有不可散布的內容，已中止。" >&2
    exit 1
fi
echo "   ✓ 沒有原版衍生素材"

rm -rf "$ROOT/dist-build"

echo "== 校驗碼"
( cd "$OUT" && find . -type f ! -name SHA256SUMS -print0 | sort -z |
    xargs -0 sha256sum > SHA256SUMS )

cat > "$OUT/README.md" <<TOP
# dist-all —— 交付物

版本 $REV（$STAMP）。

| 目錄 | 可以給別人嗎 | 內容 |
|---|---|---|
| \`release/\` | **可以** | 引擎、翻譯、文件、工具。不含任何原版素材，玩家自備原版與字型 |
| \`local/\` | **不可以** | 上面全部 ＋ 原版資料、倚天字型、合成映像、背景音樂。自己留檔用 |
| \`music/\` | 分兩半 | 十首配樂。\`midi/\` 是我們寫的譜（**可以**），\`ogg/\` 是 MT-32 算的波形（**不可以**）|
| \`promo/\` | 看情況 | 推廣片。畫面是原版的、配樂是自己寫的，公開前先想清楚 |

⚠ 三個「不可以」的理由都不是潔癖：原版資料與倚天字型有授權，
MT-32 算出來的波形裡有 Roland 的 PCM 取樣。譜與程式碼是我們自己的，那些可以給。

\`SHA256SUMS\` 收全部檔案的校驗碼。

重跑：\`tools/dist.sh\`
TOP

echo
du -sh "$REL" "$LOC" "$MUSIC" "$PROMO" 2>/dev/null || true
echo "---"
echo "產出在 $OUT（不入版控）"

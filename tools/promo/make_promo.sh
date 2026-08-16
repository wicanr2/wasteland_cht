#!/usr/bin/env bash
# 在容器裡跑：把截圖 ＋ 自製配樂合成 70 秒左右的推廣片。
#
# 由 `tools/promo.sh` 呼叫，不要直接跑（它要 /shots、/music、/out 三個掛載點）。
#
# ⚠ **不用 zoompan。** `-loop 1 -t S` 配上 `fps` 濾鏡會讓 zoompan 的 `d`
# 變成「每個輸入幀輸出 d 幀」——6 秒 25fps 算成兩萬多幀，CPU 燒滿好幾分鐘，
# 而輸出一直是空的。靜態圖加淡入淡出就夠了。
#
# 版面有六種輪流用（單一版面重複十幾段會很單調）：
# 標題卡、框內截圖、上下對照、引用卡、數據卡、結尾卡。

set -eu

. /theme.sh

SHOT=/shots; MUSIC=/music; OUT=/out; TMP=/tmp/promo
mkdir -p "$TMP" "$OUT"

# ── 素材：掃描線圖磚（母題）─────────────────────────────────────────────
# 4 像素一條暗線。合成時整張鋪過去，讓漸層底帶一點 CRT 味。
convert -size 4x4 xc:none -fill '#00000038' -draw 'rectangle 0,0 3,0' "$TMP/scan.png"

# mkbg：漸層底 ＋ 掃描線 ＋ 四角框線記號。
mkbg() {
    convert -size ${W}x${H} "gradient:${BG_LITE}-${BG_DEEP}" \
        \( -size ${W}x${H} tile:"$TMP/scan.png" \) -composite \
        -stroke "$RUST" -strokewidth 2 -fill none \
        -draw "line 36,36 116,36"   -draw "line 36,36 36,116" \
        -draw "line $((W-36)),36 $((W-116)),36" -draw "line $((W-36)),36 $((W-36)),116" \
        -draw "line 36,$((H-36)) 116,$((H-36))" -draw "line 36,$((H-36)) 36,$((H-116))" \
        -draw "line $((W-36)),$((H-36)) $((W-116)),$((H-36))" \
        -draw "line $((W-36)),$((H-36)) $((W-36)),$((H-116))" \
        "$TMP/_bg.png"
}

# shot3x：把截圖擺成 960 × 600。
#
# 截圖本身就是 960 × 600（原版 320 × 200 的 3 倍），所以這裡是原尺寸放上去。
# ⚠ 真的需要縮放時**一律 `-filter point` 配整數倍**：非整數倍會讓有些像素
# 兩格寬、有些三格寬，pixel art 看起來就髒了。
shot3x() { convert "$SHOT/$1" -filter point -resize 960x600 "$2"; }

# ── 版面一：標題卡 ───────────────────────────────────────────────────────
card() { # $1 out $2 中標 $3 英標 $4 副標
    mkbg
    convert "$TMP/_bg.png" -gravity center \
        -font "$FT" -fill "$EMBER" -pointsize 104 -annotate +5-55 "$2" \
        -fill "$SAND" -pointsize 104 -annotate +0-60 "$2" \
        -font "$FB" -fill "$RUST" -pointsize 40 -annotate +0+40 "$3" \
        -font "$FB" -fill "$BONE" -pointsize 27 -annotate +0+125 "$4" \
        "$1"
}

# ── 版面二：框內截圖 ＋ 下方字幕 ─────────────────────────────────────────
slide() { # $1 out $2 截圖 $3 字幕
    mkbg
    shot3x "$2" "$TMP/_sc.png"
    convert "$TMP/_sc.png" -bordercolor "$RUST" -border 3 "$TMP/_scb.png"
    convert "$TMP/_bg.png" "$TMP/_scb.png" -gravity north -geometry +0+16 -composite \
        -font "$FB" -fill "$BONE" -gravity south -pointsize 30 -annotate +0+34 "$3" \
        "$1"
}

# ── 版面三：上下對照（同一個畫面的英文與中文）───────────────────────────
#
# 只裁訊息視窗那一條：整張截圖並排會小到看不清字，而**要比的就是那幾行字**。
# 訊息視窗在原版是字元列 18–23，換算到 960 × 600 的截圖是 y ∈ [432, 576)。
compare2() { # $1 out $2 英文截圖 $3 中文截圖 $4 標題 $5 底部說明
    mkbg
    for i in 1 2; do
        [ $i = 1 ] && src=$2 || src=$3
        # ⚠ 停在字元列 24 之前：指令列在那一列（960 × 600 的 y ≥ 576），
        # 切太高會把它的上半截一起帶進來，看起來像壞掉的字。
        # 訊息視窗是字元列 18–23 → y ∈ [432, 576)。
        convert "$SHOT/$src" -crop 960x150+0+426 +repage \
            -bordercolor "$RUST" -border 2 "$TMP/_c$i.png"
    done
    convert "$TMP/_bg.png" \
        -font "$FT" -fill "$SAND" -gravity north -pointsize 40 -annotate +0+66 "$4" \
        "$TMP/_c1.png" -gravity north -geometry +0+200 -composite \
        "$TMP/_c2.png" -gravity north -geometry +0+420 -composite \
        -font "$FB" -fill "$DIM" -gravity northwest -pointsize 26 -annotate +160+164 "原版" \
        -font "$FB" -fill "$SAND" -gravity northwest -pointsize 26 -annotate +160+384 "繁體中文" \
        -font "$FB" -fill "$BONE" -gravity south -pointsize 30 -annotate +0+70 "$5" \
        "$1"
}

# ── 版面四：引用卡（一手史料）───────────────────────────────────────────
quote() { # $1 out $2 引文（可含 \n） $3 出處
    mkbg
    convert "$TMP/_bg.png" \
        -font "$FT" -fill "$RUST" -gravity northwest -pointsize 150 -annotate +110+40 '“' \
        -font "$FB" -fill "$BONE" -gravity northwest -pointsize 34 -annotate +140+250 "$2" \
        -font "$FB" -fill "$DIM" -gravity southeast -pointsize 24 -annotate +150+110 "$3" \
        "$1"
}

# ── 版面五：數據卡 ───────────────────────────────────────────────────────
stat5() { # $1 out ; 之後五組「數字|標籤」
    out=$1; shift   # ⚠ 先接住再 shift：底下 "$@" 是資料，$1 已經不是輸出檔名了
    mkbg
    # 欄位起點寫死不用等距：「4,873」比別的數字寬得多，等距排會擠到下一欄。
    set -- "$@"
    xs="72 302 532 812 1022"
    cmd=(convert "$TMP/_bg.png" -gravity northwest)
    i=1
    for pair in "$@"; do
        num=${pair%%|*}; lab=${pair#*|}
        x=$(echo $xs | cut -d' ' -f$i)
        cmd+=(-font "$FT" -fill "$SAND" -pointsize 72 -annotate +${x}+340 "$num")
        cmd+=(-font "$FB" -fill "$BONE" -pointsize 24 -annotate +${x}+430 "$lab")
        i=$((i + 1))
    done
    cmd+=(-font "$FT" -fill "$RUST" -gravity north -pointsize 40 -annotate +0+150 "四道閘門走完之後")
    cmd+=(-font "$FB" -fill "$DIM" -gravity south -pointsize 27 -annotate +0+120 \
        "每一份逆向筆記都要登記「remake 用上了沒有」，說接了卻沒人引用就會紅。")
    cmd+=("$out")
    "${cmd[@]}"
}

# ── 一段靜態影片（淡入淡出，不用 zoompan）───────────────────────────────
clip() { # $1 png $2 mp4 $3 秒
    fo=$(awk "BEGIN{print $3-0.6}")
    ffmpeg -y -loglevel error -loop 1 -i "$1" -t "$3" -r $FPS \
        -vf "fade=t=in:st=0:d=0.6,fade=t=out:st=$fo:d=0.6,format=yuv420p" \
        -threads 2 -c:v libx264 -preset veryfast -pix_fmt yuv420p "$2"
}

# ═══ 分鏡 ═══════════════════════════════════════════════════════════════
# 敘事骨架：**保存敘事** —— 世界觀 → 一手史料 → 玩得到的東西 → 這些字怎麼來的 → 結局 → 帳。

n=0
add() { # $1 png $2 秒
    n=$((n + 1))
    id=$(printf %02d $n)
    clip "$1" "$TMP/v$id.mp4" "$2"
    echo "file '$TMP/v$id.mp4'" >> "$TMP/list.txt"
}
: > "$TMP/list.txt"

card "$TMP/p01.png" '荒野遊俠' 'WASTELAND · Interplay 1988' \
    '逆向重製 · 繁體中文 · 連 1990 年的台灣說明書一起保存'
add "$TMP/p01.png" $PACE_CARD

slide "$TMP/p02.png" title.png \
    '1998 年，衛星從天上消失，美蘇把九成的核彈射向天空。'
add "$TMP/p02.png" $PACE_SHOT

quote "$TMP/p03.png" \
    '擁有一世紀前德州（Texas）與亞利桑納州（Arizona）\n流浪者優良傳統的 Desert Rangers 誕生了' \
    '—— 軟體世界中文說明書〈一、故事介紹〉，1990'
add "$TMP/p03.png" $PACE_QUOTE

slide "$TMP/p04.png" 01-map.png \
    '四個沙漠遊俠，一張 64 × 64 的荒漠。在野外走一步，時鐘跳四分鐘。'
add "$TMP/p04.png" $PACE_SHOT

slide "$TMP/p05.png" combat.png \
    '命中、傷害、護甲、行動順序——公式全部從 1988 年的執行檔裡讀出來。'
add "$TMP/p05.png" $PACE_SHOT

compare2 "$TMP/p06.png" skills-en.png 03-use.png '螢幕上的字都換過了' \
    '4,902 條譯文，連戰鬥表頭與店家招牌這種不在字串表裡的都補上了。'
add "$TMP/p06.png" $PACE_CARD

slide "$TMP/p07.png" 07-question.png \
    '密語、暗號、控制面板：答案逐位元組比對，答錯就是答錯。'
add "$TMP/p07.png" $PACE_SHOT

quote "$TMP/p08.png" \
    '不再有虛幻飄渺的法術，不再有猙獰可怖的妖物，\n在這個核戰後的世界裏，手上緊握的自動步槍、\n手榴彈、反坦克榴彈，甚至雷射武器將會是你的神劍魔刀。' \
    '—— 軟體世界 CD 盒背面文案，1990'
add "$TMP/p08.png" $PACE_QUOTE

slide "$TMP/p09.png" 02-journal.png \
    '162 段劇情原本印在紙上，用來擋盜版。重製版收進手札，不移植防拷。'
add "$TMP/p09.png" $PACE_SHOT

slide "$TMP/p10.png" facility.png \
    '遊俠中心、商店、醫生、訓練師——招牌、選單、清單都是中文。'
add "$TMP/p10.png" $PACE_SHOT

slide "$TMP/p11.png" help.png \
    '畫面拉到 960 × 600：中文換 24 × 24，英數用倚天同高的半形字。'
add "$TMP/p11.png" $PACE_SHOT

slide "$TMP/p12.png" 05-ending.png \
    '科奇斯基地的自毀倒數走完 240 步，然後是結局。'
add "$TMP/p12.png" $PACE_SHOT

stat5 "$TMP/p13.png" '42|張地圖全部解開' '100|份逆向筆記' \
    '4,902|條文本譯成中文' '162|段劇本' '10|首自製配樂'
add "$TMP/p13.png" $PACE_CARD

card "$TMP/p14.png" '荒野遊俠' 'github.com/wicanr2/wasteland_cht' \
    'Go / Ebiten 重製 · 原版資料與字型玩家自備 · 不散布任何原版素材'
add "$TMP/p14.png" $PACE_CARD

# ═══ 接起來 ＋ 鋪配樂 ═══════════════════════════════════════════════════
ffmpeg -y -loglevel error -f concat -safe 0 -i "$TMP/list.txt" \
    -threads 2 -c:v libx264 -preset veryfast -pix_fmt yuv420p "$TMP/silent.mp4"

DUR=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$TMP/silent.mp4")
AFO=$(awk "BEGIN{print $DUR-4}")

# ⚠ **不要用 `-shortest`。** 配樂比影片短的時候它會以音軌為準，
# 把結尾整張卡截掉——症狀是 ffprobe 出來視訊長度變成配樂長度。
# 先 `aloop` 無限循環再 `atrim` 剪到影片長度，兩軌自然等長。
ffmpeg -y -loglevel error -i "$TMP/silent.mp4" -i "$MUSIC/theme.ogg" \
    -filter_complex "[1:a]aloop=loop=-1:size=2000000000,atrim=0:$DUR,\
afade=t=in:st=0:d=2,afade=t=out:st=$AFO:d=4[a]" \
    -map 0:v -map '[a]' -threads 2 \
    -c:v libx264 -preset veryfast -pix_fmt yuv420p -c:a aac -b:a 192k \
    -movflags +faststart "$OUT/wasteland-cht-promo.mp4"

# 抽幾幀出來給人看（判斷版面單不單調）。
for t in 3 15 28 45 62; do
    ffmpeg -y -loglevel error -ss $t -i "$OUT/wasteland-cht-promo.mp4" \
        -frames:v 1 "$OUT/frame-$t.png" || true
done

echo "--- 長度與軌道"
ffprobe -v error -select_streams v:0 -show_entries stream=width,height,duration \
    -of default=nw=1 "$OUT/wasteland-cht-promo.mp4"
ffprobe -v error -select_streams a:0 -show_entries stream=codec_name,duration \
    -of default=nw=1 "$OUT/wasteland-cht-promo.mp4"

chown -R "$HOST_UID:$HOST_GID" "$OUT"

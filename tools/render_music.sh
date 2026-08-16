#!/usr/bin/env bash
# 把 tools/make_music.py 產生的 MIDI 用 Roland MT-32／CM-32L 算成 ogg。
#
#   tools/render_music.sh [輸出目錄] [ROM 目錄]
#
# 預設輸出 workplace/music/、ROM 讀 ~/cht/mt32/。
#
# 為什麼是 MT-32：1988 年 DOS 遊戲的高階音源，年代與機種都對得上。
# munt 不在 Debian 套件庫，工具鏈在 docker/media.Dockerfile 自己編。
#
# ⚠ **ROM 不進版控、也不進 image**：MT-32／CM-32L 的控制與 PCM ROM 是 Roland
# 的韌體，與倚天字型、原版遊戲資料同一個政策——玩家（或開發者）自備，
# 執行時唯讀掛載進來。
#
# ⚠ 產出的 ogg 也**不進版控**：曲子是我們寫的，但波形裡有 MT-32 的 PCM 取樣。
# 要音樂就自己跑這支；沒跑遊戲照樣玩，只是沒有 BGM。
#
# 邊界寫在腳本本體：--rm、--network none、限 2 核、ROM 唯讀掛載、
# 退出前把產物 chown 回使用者。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

OUT="${1:-$ROOT/workplace/music}"
ROMS="${2:-$HOME/cht/mt32}"
IMAGE=wasteland-media:latest

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
    echo "先建工具 image：docker build -f docker/media.Dockerfile -t $IMAGE ." >&2
    exit 1
fi
if [ ! -d "$ROMS" ]; then
    echo "找不到 MT-32 ROM 目錄：$ROMS（自備，不入版控）" >&2
    exit 1
fi

mkdir -p "$OUT"
python3 "$ROOT/tools/make_music.py" "$OUT"

docker run --rm --cpus=2 --network none \
    --log-opt max-size=10m --log-opt max-file=3 \
    -v "$OUT:/music" -v "$ROMS:/roms:ro" \
    -e HOST_UID="$(id -u)" -e HOST_GID="$(id -g)" \
    "$IMAGE" bash -c '
set -eu
cd /music
for mid in *.mid; do
    name="${mid%.mid}"
    # -i cm32l：CM-32L 的音色庫比 MT-32 多一組 PCM，1989 年之後的遊戲都用它。
    # ⚠ `-q` **不是** quiet，它吃一個整數（音質）。要安靜就把 stdout 丟掉。
    mt32emu-smf2wav -f -m /roms -i cm32l -o "$name.wav" "$mid" >/dev/null
    # 尾巴會有一小段殘響，留著；只把整體音量正規化到 -3 dBFS 免得爆音。
    ffmpeg -y -loglevel error -i "$name.wav" -af "loudnorm=I=-18:TP=-3:LRA=11" \
        -c:a libvorbis -q:a 5 "$name.ogg"
    rm -f "$name.wav"
    # 驗收（`rulebook/93` 鐵則 2）：**不准只看檔案存在**。
    # 一首安靜的曲子與一首好曲子在 `ls` 底下長得一模一樣，
    # 而最常見的壞法就是某個聲部整條沒聲音。這裡量長度、聲道、平均與尖峰音量；
    # 尖峰貼到 0 dB 是削波，平均低於 −40 dB 基本上是靜音。
    dur=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$name.ogg")
    chn=$(ffprobe -v error -select_streams a:0 -show_entries stream=channels -of csv=p=0 "$name.ogg")
    # ⚠ `volumedetect` 把結果印在 **info** 層：習慣性加上 `-v error` 會把它
    # 一起壓掉，而那時候這一欄是空的、指令仍然 exit 0——看起來像「量不到」，
    # 其實是自己把輸出關掉了。
    vol=$(ffmpeg -hide_banner -nostats -v info -i "$name.ogg" \
        -af volumedetect -f null /dev/null 2>&1 |
        sed -n "s/.*_volume: \(.*\) dB/\1/p" | tr "\n" " ")
    printf "%-10s %6.1f 秒  %s 聲道  平均/尖峰 %s dB\n" "$name" "$dur" "$chn" "$vol"
done
chown -R "$HOST_UID:$HOST_GID" /music
'

echo "---"
echo "產出在 $OUT（.ogg 不入版控，見檔頭）"

#!/usr/bin/env bash
# 以 FluidR3 GM 產生「現代版」十首配樂。樂譜仍由 tools/make_music.py 單一來源產生，
# 但配器、GM 鼓組、空間與動態處理和 MT-32 復古版分開。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUT="${1:-$ROOT/workplace/music/modern}"
IMAGE=wasteland-media:latest

mkdir -p "$OUT"
docker run --rm --network none --memory 3g --cpus 2 --pids-limit 512 \
    -u "$(id -u):$(id -g)" -e HOME=/tmp/wl-music-home \
    -v "$ROOT:/src:ro" -v "$OUT:/out" -w /src "$IMAGE" bash -c '
set -euo pipefail
python3 tools/make_music.py --profile modern /out
sf=/usr/share/sounds/sf2/FluidR3_GM.sf2
for mid in /out/*.mid; do
    name=$(basename "${mid%.mid}")
    fluidsynth -ni -F "/out/$name.wav" -r 44100 "$sf" "$mid" >/dev/null
    ffmpeg -y -loglevel error -i "/out/$name.wav" \
        -af "highpass=f=32,lowpass=f=16500,acompressor=threshold=-18dB:ratio=2.4:attack=18:release=220,loudnorm=I=-17:TP=-2:LRA=10" \
        -ar 44100 -ac 2 -c:a libvorbis -q:a 6 "/out/$name.ogg"
    rm -f "/out/$name.wav"
    dur=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "/out/$name.ogg")
    vol=$(ffmpeg -hide_banner -nostats -v info -i "/out/$name.ogg" -af volumedetect \
        -f null /dev/null 2>&1 | sed -n "s/.*_volume: \(.*\) dB/\1/p" | tr "\n" " ")
    printf "%-10s %6.1f 秒  平均/尖峰 %s dB\n" "$name" "$dur" "$vol"
done
cp /usr/share/doc/fluid-soundfont-gm/copyright /out/FluidR3-GM-LICENSE.txt
sfsha=$(sha256sum "$sf" | cut -d" " -f1)
pkg=$(dpkg-query -W -f="\${Version}" fluid-soundfont-gm)
printf "{\n  \"schema\": 1,\n  \"label\": \"new remake arrangement\",\n  \"profile\": \"modern\",\n  \"generator\": \"tools/make_music.py\",\n  \"soundfont_package\": \"fluid-soundfont-gm %s\",\n  \"soundfont_sha256\": \"%s\",\n  \"sample_rate\": 44100,\n  \"channels\": 2\n}\n" "$pkg" "$sfsha" > /out/metadata.json
'

echo "現代配樂在 $OUT"

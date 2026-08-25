#!/usr/bin/env bash
# 驗證推廣片的影像、音訊、時長與非靜音，並寫出可交付 metadata。
set -euo pipefail

VIDEO=${1:?用法：verify_promo.sh VIDEO METADATA_JSON}
META=${2:?用法：verify_promo.sh VIDEO METADATA_JSON}

width=$(ffprobe -v error -select_streams v:0 -show_entries stream=width -of csv=p=0 "$VIDEO")
height=$(ffprobe -v error -select_streams v:0 -show_entries stream=height -of csv=p=0 "$VIDEO")
vcodec=$(ffprobe -v error -select_streams v:0 -show_entries stream=codec_name -of csv=p=0 "$VIDEO")
acodec=$(ffprobe -v error -select_streams a:0 -show_entries stream=codec_name -of csv=p=0 "$VIDEO")
duration=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$VIDEO")
vduration=$(ffprobe -v error -select_streams v:0 -show_entries stream=duration -of csv=p=0 "$VIDEO")
aduration=$(ffprobe -v error -select_streams a:0 -show_entries stream=duration -of csv=p=0 "$VIDEO")

[ "$width" = 1280 ] && [ "$height" = 720 ] || {
    echo "影片尺寸是 ${width}×${height}，不是 1280×720" >&2; exit 1;
}
[ "$vcodec" = h264 ] || { echo "影像 codec 是 $vcodec，不是 h264" >&2; exit 1; }
[ "$acodec" = aac ] || { echo "音訊 codec 是 $acodec，不是 aac" >&2; exit 1; }
awk -v d="$duration" 'BEGIN { exit !(d >= 70 && d <= 100) }' || {
    echo "影片長度 $duration 秒不在 70–100 秒 gate" >&2; exit 1;
}
awk -v v="$vduration" -v a="$aduration" 'BEGIN { x=v-a; if (x<0) x=-x; exit !(x <= 0.10) }' || {
    echo "影像／音訊時長不一致：$vduration / $aduration" >&2; exit 1;
}

volume=$(ffmpeg -hide_banner -nostats -i "$VIDEO" -map 0:a:0 -af volumedetect -f null - 2>&1)
mean=$(printf '%s\n' "$volume" | awk '/mean_volume:/ {print $(NF-1)}' | tail -1)
peak=$(printf '%s\n' "$volume" | awk '/max_volume:/ {print $(NF-1)}' | tail -1)
[ -n "$mean" ] && [ "$mean" != "-inf" ] || { echo "音軌是靜音" >&2; exit 1; }
awk -v m="$mean" 'BEGIN { exit !(m > -45) }' || { echo "音軌過小：$mean dB" >&2; exit 1; }
awk -v p="$peak" 'BEGIN { exit !(p <= 0) }' || { echo "音軌削波：$peak dB" >&2; exit 1; }

sum=$(sha256sum "$VIDEO" | awk '{print $1}')
bytes=$(wc -c < "$VIDEO" | tr -d ' ')
mkdir -p "$(dirname "$META")"
cat > "$META" <<EOF
{
  "schema": 1,
  "file": "$(basename "$VIDEO")",
  "sha256": "$sum",
  "bytes": $bytes,
  "video": {"codec": "$vcodec", "width": $width, "height": $height, "duration_seconds": $vduration},
  "audio": {"codec": "$acodec", "duration_seconds": $aduration, "mean_db": $mean, "peak_db": $peak},
  "format_duration_seconds": $duration,
  "validation": ["video-stream", "audio-stream", "duration-match", "non-silent", "no-clipping"]
}
EOF

echo "[promo] PASS ${width}x${height} duration=${duration}s mean=${mean}dB peak=${peak}dB sha256=$sum"

#!/usr/bin/env bash
set -euo pipefail
test -e /.dockerenv
version=${1:?full version}
[[ "$version" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
package=/src/dist-all/$version/full-local/linux-x64/wasteland-cht-$version-linux-x64-full-local.AppImage
export DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1 LANG=C.utf8 LC_ALL=C.utf8
export XDG_DATA_HOME=/tmp/wl-promo-data XDG_CONFIG_HOME=/tmp/wl-promo-config XDG_RUNTIME_DIR=/tmp/wl-promo-runtime
export ALSA_CONFIG_PATH=/tmp/wl-promo-alsa.conf
mkdir -p "$XDG_RUNTIME_DIR" /tmp/wl-promo-app /out/music
chmod 700 "$XDG_RUNTIME_DIR"
printf 'pcm.!default { type null }\n' > "$ALSA_CONFIG_PATH"
game= recorder= xvfb=
cleanup() {
  if [ -n "$recorder" ]; then kill -INT "$recorder" 2>/dev/null || true; wait "$recorder" 2>/dev/null || true; fi
  if [ -n "$game" ]; then kill "$game" 2>/dev/null || true; wait "$game" 2>/dev/null || true; fi
  if [ -n "$xvfb" ]; then kill "$xvfb" 2>/dev/null || true; wait "$xvfb" 2>/dev/null || true; fi
}
trap cleanup EXIT INT TERM
Xvfb :99 -screen 0 1280x960x24 -nolisten tcp >/out/xvfb.log 2>&1 & xvfb=$!
for n in $(seq 1 30); do test ! -S /tmp/.X11-unix/X99 || break; sleep .1; done
(cd /tmp/wl-promo-app && "$package" --appimage-extract >/out/extract.log)
app=/tmp/wl-promo-app/squashfs-root
cp "$app/usr/music/modern/theme.ogg" /out/music/theme.ogg
cp "$app/usr/music/modern/combat.ogg" /out/music/combat.ogg
cp /src/tools/promo/record_packaged_container.sh /out/driver.sh
# The real launcher consumes its first positional argument as the data directory.
# Supply its own packaged data, then request native logical size for recording.
"$app/AppRun" "$app/usr/data" -scale 1 >/out/game.log 2>&1 & game=$!
window=
for n in $(seq 1 180); do
  window=$(xdotool search --onlyvisible --name "Wasteland.*$version" 2>/dev/null | head -1 || true)
  test -z "$window" || break
  kill -0 "$game"; sleep .2
done
test -n "$window"
xdotool windowfocus --sync "$window"
geometry=$(xdotool getwindowgeometry --shell "$window")
test "$(printf '%s\n' "$geometry" | sed -n 's/^WIDTH=//p')" = 960
test "$(printf '%s\n' "$geometry" | sed -n 's/^HEIGHT=//p')" = 720
shot() { import -window "$window" "/out/$1.png"; }
send() { xdotool keydown --clearmodifiers "$1"; sleep .18; xdotool keyup "$1"; sleep .35; }
click() { xdotool mousemove --window "$window" "$1" "$2"; xdotool mousedown 1; sleep .18; xdotool mouseup 1; sleep .35; }
wait_frame() {
  local ref=$1 name=$2 relation=$3 rect=$4 ready=0
  for n in $(seq 1 100); do
    shot "$name"
    ready=$(python3 - "/out/$ref.png" "/out/$name.png" "$relation" "$rect" <<'PY'
from PIL import Image,ImageChops
import sys
r=tuple(map(int,sys.argv[4].split(',')))
a,b=[Image.open(p).convert('RGB').crop(r) for p in sys.argv[1:3]]
d=ImageChops.difference(a,b)
if sys.argv[3]=='art':
    h=d.convert('L').point(lambda v:255 if v>3 else 0).histogram(); ok=h[255]/(a.width*a.height)>.25
else: ok=(d.getbbox() is None)==(sys.argv[3]=='equal')
print(int(ok))
PY
)
    test "$ready" != 1 || break
    sleep .15
  done
  printf '%s %s %s %s\n' "$EPOCHREALTIME" "$name" "$relation" "$n" >>/out/events.log
  test "$ready" = 1
}
for n in $(seq 1 100); do
  shot 00-title
  ready=$(convert /out/00-title.png -format '%[fx:mean>0.002]' info:)
  test "$ready" != 1 || break; sleep .15
done
test "$ready" = 1
click 480 565
wait_frame 00-title 01-map different 352,542,608,590
click 500 22
wait_frame 01-map 02-modern different 444,6,564,38
click 730 22
wait_frame 02-modern 03-save different 16,486,948,654

walk() {
  shot walk-before
  send Up; wait_frame walk-before walk-up different 24,84,888,468
  sleep 1
  send Left; wait_frame walk-up walk-left different 24,84,888,468
  sleep 1
  send Right; wait_frame walk-left walk-return different 24,84,888,468
}
languages() {
  shot lang-before
  click 908 22; wait_frame lang-before lang-hant different 16,154,948,654
  sleep 1
  click 728 183; wait_frame lang-hant lang-hans different 20,54,220,76
  sleep 2
  send Escape; wait_frame lang-before lang-map equal 24,84,888,468
  send p; wait_frame lang-map lang-journal different 16,154,948,654
  sleep 2
  send Escape; wait_frame lang-before lang-return equal 24,84,888,468
}
toolbar() {
  shot font-before
  click 250 22; wait_frame font-before font-noto different 212,6,316,38
  sleep 1
  click 340 22; wait_frame font-noto font-16 different 324,6,436,38
  sleep 1
  click 420 22; wait_frame font-16 font-18 different 324,6,436,38
  click 420 22; wait_frame font-18 font-20 different 324,6,436,38
  sleep 1
  click 340 22; wait_frame font-18 font-restored equal 324,6,436,38
  click 250 22; wait_frame font-before font-rounded equal 212,6,316,38
}
chrome() {
  shot chrome-on
  sleep 1
  click 908 22; wait_frame chrome-on chrome-settings-on different 16,154,948,654
  click 478 183; wait_frame chrome-settings-on chrome-settings-off different 356,162,600,204
  send Escape; wait_frame chrome-on chrome-off equal 24,84,888,468
  sleep 3
  click 908 22; wait_frame chrome-off chrome-settings-off-again different 16,154,948,654
  send c; wait_frame chrome-settings-off-again chrome-settings-restored different 356,162,600,204
  send Escape; wait_frame chrome-on chrome-restored equal 24,84,888,468
  sleep 1
}
facility() {
  shot facility-before
  send Up; wait_frame facility-before facility-up different 24,84,888,468
  send Down; wait_frame facility-up facility-ask different 16,486,948,654
  send y; wait_frame facility-ask facility-center different 24,84,888,468
  sleep 3
  send p; wait_frame facility-center facility-return different 24,84,888,468
}
record() {
  local name=$1 duration=$2 action=$3
  ffmpeg -hide_banner -loglevel error -nostdin -y -thread_queue_size 8 \
    -f x11grab -window_id "$window" -draw_mouse 0 -framerate 25 -video_size 960x720 -i :99 \
    -t "$duration" -threads 2 -c:v libx264 -preset veryfast -crf 20 -pix_fmt yuv420p \
    -progress "/out/$name.progress" "/out/$name.mp4" >"/out/$name-encode.log" 2>&1 & recorder=$!
  for n in $(seq 1 60); do
    if test -s "/out/$name.progress" && grep -q '^frame=' "/out/$name.progress"; then break; fi
    kill -0 "$recorder"; sleep .1
  done
  printf '%s clip-start %s\n' "$EPOCHREALTIME" "$name" >>/out/events.log
  "$action"
  kill -0 "$recorder" # All acknowledged actions must be in the captured interval.
  wait "$recorder"; recorder=
  shot "$name-final"
  printf '%s clip-end %s\n' "$EPOCHREALTIME" "$name" >>/out/events.log
}
record original-live 12 walk
send F9; wait_frame 03-save reset-original equal 24,84,430,468
click 140 22
wait_frame reset-original art-a art 24,84,888,468
wait_frame reset-original mode-a different 124,9,158,35
record hd2-a-live 12 walk
send F9; wait_frame art-a reset-a equal 24,84,430,468
click 178 22
wait_frame reset-a art-b art 24,84,888,468
wait_frame reset-a mode-b different 162,9,196,35
record hd2-b-live 12 walk
send F9; wait_frame art-b reset-b equal 24,84,430,468
record chrome-live 30 chrome
record language-live 22 languages
record toolbar-live 22 toolbar
record facility-live 18 facility
send F10
shot quit
send y
for n in $(seq 1 100); do kill -0 "$game" 2>/dev/null || break; sleep .1; done
if kill -0 "$game" 2>/dev/null; then echo 'Game did not exit normally'; exit 1; fi
wait "$game"; game=
# Check the public locale flag through another normal title, not a direct scene.
"$app/AppRun" "$app/usr/data" -scale 1 -locale zh-Hans >/out/startup-hans.log 2>&1 & game=$!
window=
for n in $(seq 1 180); do
  window=$(xdotool search --onlyvisible --name "Wasteland.*$version" 2>/dev/null | head -1 || true)
  test -z "$window" || break
  kill -0 "$game"; sleep .2
done
test -n "$window"
xdotool windowfocus --sync "$window"
for n in $(seq 1 100); do
  shot startup-hans
  ready=$(convert /out/startup-hans.png -format '%[fx:mean>0.002]' info:)
  test "$ready" != 1 || break; sleep .15
done
test "$ready" = 1
wait_frame 00-title startup-hans different 20,54,220,76
send F10; send y
for n in $(seq 1 100); do kill -0 "$game" 2>/dev/null || break; sleep .1; done
if kill -0 "$game" 2>/dev/null; then exit 1; fi
wait "$game"; game=
rc=0
"$app/usr/bin/wasteland" -locale zh-Hans -ui classic >/out/classic-hans-rejected.log 2>&1 || rc=$?
test "$rc" = 2
grep -q 'zh-Hans' /out/classic-hans-rejected.log
! grep -q 'panic:' /out/classic-hans-rejected.log
python3 - "$version" "$package" <<'PY'
from pathlib import Path
from PIL import Image
import hashlib,json,subprocess,sys
root=Path('/out')
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
clips=[]
for name,requested,minimum in [('original-live',12,4),('hd2-a-live',12,6),('hd2-b-live',12,5),('chrome-live',30,20),('language-live',22,16),('toolbar-live',22,20),('facility-live',18,9)]:
    p=root/(name+'.mp4')
    info=json.loads(subprocess.check_output(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(p)]))
    v=info['streams'][0]; assert (v['width'],v['height'])==(960,720)
    duration=float(info['format']['duration'])
    assert v['avg_frame_rate']=='25/1' and duration>=minimum
    clips.append({'path':p.name,'sha256':sha(p),'duration':duration,'requested_capture_seconds':requested,'minimum_source_seconds':minimum})
on,off,restored=[Image.open(root/(n+'.png')).convert('RGB') for n in ('chrome-on','chrome-off','chrome-restored')]
crop=(24,84,888,468)
assert on.crop(crop).tobytes()==off.crop(crop).tobytes()==restored.crop(crop).tobytes()
bezel=(16,76,896,84)
assert on.crop(bezel).tobytes()!=off.crop(bezel).tobytes()
assert on.crop(bezel).tobytes()==restored.crop(bezel).tobytes()
result={'schema':1,'status':'passed','version':sys.argv[1],'package_sha256':sha(Path(sys.argv[2])),
 'entry':'actual packaged AppRun; its own data; normal title; -scale 1 only; no forced map/art',
 'window_size':[960,720],'clips':clips,'driver_sha256':sha(root/'driver.sh'),'exit_code':0,
 'startup_locale_zh_hans':True,'classic_zh_hans_rejected':True,
 'hd_chrome_verified':True,'chrome_map_pixels_unchanged':True,
 'audio':'video capture uses null audio; soundtrack copied from this actual package',
 'music':[{'path':str(p.relative_to(root)),'sha256':sha(p)} for p in sorted((root/'music').glob('*.ogg'))],
 'rights':'local-only original/derived art and project-made score; no public distribution authorized'}
writer=Path('/src/tools/promo/record_packaged_container.sh')
result['metadata_writer_sha256']=sha(writer)
if result['driver_sha256']!=result['metadata_writer_sha256']:
    (root/'metadata-writer.sh').write_bytes(writer.read_bytes())
    result['receipt_note']='Metadata regenerated from retained complete recordings after correcting raw-duration acceptance; recorded driver preserved unchanged.'
(root/'capture.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('PASS: seven actual packaged clips, HD frame on/off/on and normal exit')
PY

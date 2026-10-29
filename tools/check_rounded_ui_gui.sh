#!/bin/bash
# Execute inside the existing Go/X11 Docker image. No host workload.
set -euo pipefail
test -e /.dockerenv
evidence=/src/workplace/art-preview/ui-gui
mkdir -p "$evidence" /tmp/wl-ui-runtime
chmod 700 /tmp/wl-ui-runtime
export DISPLAY=:98 XDG_RUNTIME_DIR=/tmp/wl-ui-runtime LIBGL_ALWAYS_SOFTWARE=1
export ALSA_CONFIG_PATH=/tmp/wl-ui-null-alsa.conf
printf 'pcm.!default { type null }\n' > "$ALSA_CONFIG_PATH"
game_pid= xvfb_pid=
cleanup() {
    if [ -n "$game_pid" ]; then kill "$game_pid" 2>/dev/null || true; wait "$game_pid" 2>/dev/null || true; fi
    if [ -n "$xvfb_pid" ]; then kill "$xvfb_pid" 2>/dev/null || true; wait "$xvfb_pid" 2>/dev/null || true; fi
}
trap cleanup EXIT INT TERM
Xvfb :98 -screen 0 1280x960x24 -nolisten tcp > "$evidence/xvfb.log" 2>&1 &
xvfb_pid=$!
for n in $(seq 1 50); do [ ! -S /tmp/.X11-unix/X98 ] || break; sleep .1; done
/src/workplace/game-test/wasteland-ui-dev -mode play -scale 1 \
    -rom /src/workplace/orig/wastland -image /src/workplace/analysis/unpacked/wl.merged.exe \
    -lang /src/translations/zh-Hant.cat -font /tmp/no-eten-font \
    -quicksave "$evidence/quick.wlqs" > "$evidence/game.log" 2>&1 &
game_pid=$!
window=
for n in $(seq 1 120); do
    window=$(xdotool search --onlyvisible --name 'Wasteland.*荒野遊俠' 2>/dev/null | head -1 || true)
    [ -z "$window" ] || break
    kill -0 "$game_pid"
    sleep .25
done
test -n "$window"
xdotool windowfocus --sync "$window"
shot() { import -window "$window" "$evidence/$1.png"; }
send() { xdotool keydown --clearmodifiers "$1"; sleep .15; xdotool keyup "$1"; sleep .35; }
click() { xdotool mousemove --window "$window" "$1" "$2"; xdotool mousedown 1; sleep .15; xdotool mouseup 1; sleep .35; }
for n in $(seq 1 100); do
    shot 00-title
    ready=$(convert "$evidence/00-title.png" -format '%[fx:mean>0.002]' info:)
    [ "$ready" != 1 ] || break
    sleep .15
done
test "$ready" = 1
click 480 565
shot 01-original
click 250 22
shot 02-noto
click 340 22
shot 02-noto-16
click 420 22
click 420 22
shot 03-noto-20
click 340 22
click 250 22
wait_art() {
    local previous=$1 current=$2 changed=0
    for n in $(seq 1 120); do
        shot "$current"
        changed=$(python3 - "$evidence/$previous.png" "$evidence/$current.png" <<'PY'
from PIL import Image, ImageChops
import sys
a,b=[Image.open(p).convert('RGB').crop((24,84,888,468)) for p in sys.argv[1:]]
h=ImageChops.difference(a,b).convert('L').point(lambda v:255 if v>3 else 0).histogram()
print(int(h[255]/(864*384)>.25))
PY
)
        [ "$changed" != 1 ] || break
        sleep .25
    done
    test "$changed" = 1
}
click 140 22
wait_art 01-original 04-hd2-a
click 178 22
wait_art 04-hd2-a 05-hd2-b
click 500 22
click 670 22
click 622 22
click 644 22
click 644 22
shot 06-hd2-b-modern
click 730 22
shot 07-saved
send Up
shot 08-moved
click 820 22
shot 09-loaded
send Up
send Down
send y
shot 21-ranger-center
send p
shot 22-ranger-return
send F9
send 1
shot 10-character
send Return
shot 11-items
click 38 264
shot 12-item-action
send Escape
send Escape
send 1
send Return
send Return
shot 13-skills
send Escape
send p
shot 14-journal
send Page_Down
shot 15-journal-scroll
send Escape
click 908 22
shot 16-settings
click 900 182
send e
shot 17-encounter-ask
send y
shot 18-combat
send Escape
send F9
send F10
shot 19-quit
click 250 22
click 730 22
shot 20-quit-click-blocked
send n
send F10
send y
for n in $(seq 1 100); do kill -0 "$game_pid" 2>/dev/null || break; sleep .1; done
if kill -0 "$game_pid" 2>/dev/null; then echo 'Game did not exit normally'; exit 1; fi
wait "$game_pid"
game_pid=
python3 - "$evidence" <<'PY'
from pathlib import Path
from PIL import Image, ImageChops
import hashlib,json,sys
p=Path(sys.argv[1]);crop=(24,84,888,468)
images={f.stem:Image.open(f).convert('RGB') for f in p.glob('*.png')}
assert all(im.size==(960,720) for im in images.values())
for a,b in [('01-original','04-hd2-a'),('04-hd2-a','05-hd2-b'),('01-original','05-hd2-b')]:
    assert ImageChops.difference(images[a].crop(crop),images[b].crop(crop)).getbbox(),(a,b)
assert images['07-saved'].crop(crop).tobytes()==images['09-loaded'].crop(crop).tobytes()
assert images['19-quit'].tobytes()==images['20-quit-click-blocked'].tobytes()
assert (p/'quick.wlqs').read_bytes().startswith(b'WLQS\x02')
result={'status':'passed','scope':'Linux normal title and actual toolbar/keyboard; no ETen; normal Ranger Center entry/return; empty encounter combat UI; manual image review also required','native_windows_macos':'not-run','exit_code':0,'quicksave_version':2,'map_load_restored':True,'quit_blocks_toolbar':True,'binary_sha256':hashlib.sha256(Path('/src/workplace/game-test/wasteland-ui-dev').read_bytes()).hexdigest(),'files':[{'path':f.name,'sha256':hashlib.sha256(f.read_bytes()).hexdigest()} for f in sorted(p.glob('*.png'))]}
(p/'result.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('PASS: actual native-size GUI, original/A/B, font/size, menu, journal, save/load, quit ownership and normal exit.')
PY

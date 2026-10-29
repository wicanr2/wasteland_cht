#!/bin/bash
# Inside the Go/X11 Docker image. Windows attaches to the isolated Wine server.
set -euo pipefail
test -e /.dockerenv
mode=${1:?patch or full-local}
version=${2:?full version}
platform=${3:-linux-x64}
[[ "$version" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
[[ "$mode" = patch || "$mode" = full-local ]]
[[ "$platform" = linux-x64 || "$platform" = windows-x64 ]]
if [ "$platform" = windows-x64 ]; then
    test "$mode" = full-local
    folder=full-local-windows-wine
else
    folder=$mode-linux-gui
fi
evidence=/evidence/$folder
mkdir -p "$evidence/window" /tmp/wl-package-runtime
chmod 700 /tmp/wl-package-runtime
export DISPLAY=:99 LANG=C.utf8 LC_ALL=C.utf8
export XDG_RUNTIME_DIR=/tmp/wl-package-runtime XDG_DATA_HOME=/tmp/wl-package-data XDG_CONFIG_HOME=/tmp/wl-package-config
export LIBGL_ALWAYS_SOFTWARE=1 ALSA_CONFIG_PATH=/tmp/wl-package-alsa.conf
printf 'pcm.!default { type null }\n' > "$ALSA_CONFIG_PATH"
game_pid= xvfb_pid=
cleanup() {
    if [ -n "$game_pid" ]; then kill "$game_pid" 2>/dev/null || true; wait "$game_pid" 2>/dev/null || true; fi
    if [ -n "$xvfb_pid" ]; then kill "$xvfb_pid" 2>/dev/null || true; wait "$xvfb_pid" 2>/dev/null || true; fi
}
trap cleanup EXIT INT TERM
if [ "$platform" = linux-x64 ]; then
    Xvfb :99 -screen 0 2560x1600x24 -nolisten tcp > "$evidence/xvfb.log" 2>&1 &
    xvfb_pid=$!
    for n in $(seq 1 50); do [ ! -S /tmp/.X11-unix/X99 ] || break; sleep .1; done
    package=/src/dist-all/$version/$mode/linux-x64/wasteland-cht-$version-linux-x64-$mode.AppImage
    if [ "$mode" = patch ]; then
        "$package" --appimage-extract-and-run /src/workplace/orig/wastland > "$evidence/game.log" 2>&1 &
    else
        "$package" --appimage-extract-and-run > "$evidence/game.log" 2>&1 &
    fi
    game_pid=$!
fi
window=
for n in $(seq 1 240); do
    window=$(xdotool search --onlyvisible --name "Wasteland.*$version" 2>/dev/null | head -1 || true)
    [ -z "$window" ] || break
    if [ -n "$game_pid" ]; then kill -0 "$game_pid"; fi
    [ ! -f "$evidence/exit-code.txt" ] || { cat "$evidence/game.log"; exit 1; }
    sleep .25
done
test -n "$window"
xdotool windowfocus --sync "$window"
geometry=$(xdotool getwindowgeometry --shell "$window")
window_width=$(printf '%s\n' "$geometry" | sed -n 's/^WIDTH=//p')
window_height=$(printf '%s\n' "$geometry" | sed -n 's/^HEIGHT=//p')
test "$((window_width*3))" = "$((window_height*4))"
shot() {
    import -window "$window" /tmp/wl-package-shot.png
    convert /tmp/wl-package-shot.png -filter point -resize 960x720 "$evidence/$1.png"
    case "$1" in 00-title|06-art-b-modern|09-loaded)
        cp /tmp/wl-package-shot.png "$evidence/window/$1.png" ;;
    esac
}
send() { xdotool keydown --clearmodifiers "$1"; sleep .18; xdotool keyup "$1"; sleep .4; }
click() {
    xdotool mousemove --window "$window" "$(( $1*window_width/960 ))" "$(( $2*window_height/720 ))"
    xdotool mousedown 1; sleep .18; xdotool mouseup 1; sleep .4
}
for n in $(seq 1 120); do
    shot 00-title
    ready=$(convert "$evidence/00-title.png" -format '%[fx:mean>0.002]' info:)
    [ "$ready" != 1 ] || break
    sleep .15
done
test "$ready" = 1
wait_art() {
    local previous=$1 current=$2 changed=0
    for n in $(seq 1 120); do
        shot "$current"
        changed=$(python3 - "$evidence/$previous.png" "$evidence/$current.png" <<'PY'
from PIL import Image,ImageChops
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
click 480 565
shot 01-original
click 250 22
shot 02-noto
click 340 22
shot 03-noto-16
click 420 22
click 420 22
shot 04-noto-20
click 340 22
click 250 22
click 140 22
if [ "$mode" = full-local ]; then
    wait_art 01-original 05-art-a
    click 178 22
    wait_art 05-art-a 06-art-b-modern
    click 500 22
    shot 06-art-b-modern
else
    shot 05-missing-hd2
    send Escape
fi
click 730 22
shot 07-saved
send Up
shot 08-moved
click 820 22
shot 09-loaded
send Up
send Down
send y
shot 10-ranger-center
send p
shot 11-ranger-return
send F9
send 1
shot 12-character
send Return
shot 13-items
click 38 264
shot 14-item-action
send Escape
send Escape
send p
shot 15-journal
send Escape
click 908 22
shot 16-settings
send Escape
send F10
shot 17-quit
click 250 22
click 730 22
shot 18-quit-owned
send n
send F10
send y
if [ "$platform" = linux-x64 ]; then
    for n in $(seq 1 120); do kill -0 "$game_pid" 2>/dev/null || break; sleep .1; done
    if kill -0 "$game_pid" 2>/dev/null; then echo 'Game did not exit normally'; exit 1; fi
    wait "$game_pid"
    game_pid=
else
    for n in $(seq 1 120); do [ ! -f "$evidence/exit-code.txt" ] || break; sleep .2; done
    test "$(cat "$evidence/exit-code.txt")" = 0
fi
python3 - "$evidence" "$mode" "$version" "$platform" "$window_width" "$window_height" <<'PY'
from pathlib import Path
from PIL import Image,ImageChops
import hashlib,json,sys
p=Path(sys.argv[1]);mode,version,platform=sys.argv[2:5];crop=(24,84,888,468)
images={f.stem:Image.open(f).convert('RGB') for f in p.glob('*.png')}
assert all(im.size==(960,720) for im in images.values())
assert images['07-saved'].crop(crop).tobytes()==images['09-loaded'].crop(crop).tobytes()
assert ImageChops.difference(images['07-saved'].crop(crop),images['08-moved'].crop(crop)).getbbox()
assert images['17-quit'].tobytes()==images['18-quit-owned'].tobytes()
assert images['02-noto'].crop((212,6,316,38)).tobytes()!=images['01-original'].crop((212,6,316,38)).tobytes()
assert images['03-noto-16'].crop((324,6,436,38)).tobytes()!=images['04-noto-20'].crop((324,6,436,38)).tobytes()
assert sum(sum(px)>30 for px in images['00-title'].crop((12,6,948,38)).getdata())/(936*32)>.5
visual=False
if mode=='full-local':
    for a,b in [('01-original','05-art-a'),('05-art-a','06-art-b-modern'),('01-original','06-art-b-modern')]:
        h=ImageChops.difference(images[a].crop(crop),images[b].crop(crop)).convert('L').point(lambda v:255 if v>3 else 0).histogram()
        assert h[255]/(864*384)>.25,(a,b)
    visual=True
if platform=='windows-x64':
    for name in ['wasteland','wl-setup']: assert (p/f'{name}-version.txt').read_text().strip()==version
    paths=list(Path('/wine').rglob('quicksave.wlq'));assert len(paths)==1,paths
    save=paths[0]
else:
    save=Path('/tmp/wl-package-data/wasteland-cht/quicksave.wlq')
b=save.read_bytes();assert b[:6]==b'WLQS\x02\x00'
(p/'quick.wlqs').write_bytes(b)
suffix='AppImage' if platform=='linux-x64' else 'zip'
package=Path(f'/src/dist-all/{version}/{mode}/{platform}/wasteland-cht-{version}-{platform}-{mode}.{suffix}')
result={'schema':1,'status':'passed','version':version,'platform':platform,'environment':'Wine supplementary' if platform=='windows-x64' else 'Linux Xvfb',
 'entry':'actual packaged launcher; normal title; no forced map or art mode','exit_code':0,
 'modern_ui_verified':True,'visual_modes_verified':visual,'load_restores_exact_crop':crop,
 'window_size':[int(x) for x in sys.argv[5:7]],'screenshot_scale':'nearest to logical 960x720; full-size window captures retained',
 'quick_save':{'version':2,'original_path':str(save),'sha256':hashlib.sha256(b).hexdigest()},
 'package_sha256':hashlib.sha256(package.read_bytes()).hexdigest(),
 'driver_sha256':hashlib.sha256(Path('/src/tools/check_package_gui.sh').read_bytes()).hexdigest(),
 'limits':'No native Windows/macOS hardware or human audio acceptance.',
 'screenshots':[{'path':str(f.relative_to(p)),'sha256':hashlib.sha256(f.read_bytes()).hexdigest()} for f in sorted(p.rglob('*.png'))]}
(p/'result.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('PACKAGE GUI PASS',version,mode,platform)
PY

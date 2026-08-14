#!/usr/bin/env bash
# 容器內實際執行的腳本：起 Xvfb、產生 dosbox.conf、跑 DOSBox、
# 依 timeline 送鍵／截圖、收尾。由 tools/dosbox.sh 從 host 呼叫。
#
# 用法：
#   entrypoint.sh <執行檔> <timeline> [cycles]
#
# timeline 步驟（';' 分隔，依序執行）：
#   wait:N        等 N 秒
#   key:KEYSYM    送一個按鍵（xdotool keysym：Return / space / Escape / Up …）
#   type:STRING   打一段文字（不含 Enter）
#   shot:NAME     截圖存成 /shots/NAME.png
#
# ⚠ cycles 一律寫死，**不要用 auto**——那是可重現性的敵人
# （~/.claude/knowledge-base/retro/dosbox-game-configs.md）。

set -uo pipefail

EXE="${1:?用法: entrypoint.sh <執行檔> <timeline> [cycles]}"
TIMELINE="${2:-}"
CYCLES="${3:-fixed 3000}"

export DISPLAY=:99
mkdir -p /shots/dosbox-captures

echo "[entrypoint] 啟動 Xvfb ..."
Xvfb :99 -screen 0 1024x768x24 -nolisten tcp &
XVFB_PID=$!
sleep 1

# machine=ega：原版是 mode 0Dh（320 × 200、16 色、4 平面，docs/re/04 §4）。
CONF=/tmp/dosbox-wl.conf
cat > "$CONF" << EOF
[sdl]
fullscreen=false
fulldouble=false
output=surface
autolock=false
waitonerror=false
priority=normal,normal

[dosbox]
language=
machine=ega
captures=/shots/dosbox-captures
memsize=64

[render]
frameskip=0
aspect=false
scaler=none

[cpu]
core=normal
cputype=386
cycles=${CYCLES}
cycleup=10
cycledown=20

[mixer]
nosound=true

[midi]
mididevice=none

[sblaster]
sbtype=none
oplmode=none

[gus]
gus=false

[speaker]
pcspeaker=false
tandy=off
disney=false

[joystick]
joysticktype=none

[serial]
serial1=dummy
serial2=dummy
serial3=disabled
serial4=disabled

[dos]
xms=true
ems=true
umb=true
keyboardlayout=us

[ipx]
ipx=false

[autoexec]
mount c /game
c:
${EXE}
EOF

echo "[entrypoint] 啟動 DOSBox（machine=ega, core=normal, cycles=${CYCLES}, exe=${EXE}）..."
dosbox -conf "$CONF" -userconf > /tmp/dosbox.log 2>&1 &
DOSBOX_PID=$!

WIN=""
for i in $(seq 1 30); do
    WIN=$(xdotool search --name DOSBox 2>/dev/null | head -1)
    [[ -n "$WIN" ]] && break
    sleep 0.5
done
if [[ -z "$WIN" ]]; then
    echo "[entrypoint] 錯誤：15 秒內沒等到 DOSBox 視窗" >&2
    cat /tmp/dosbox.log >&2
    kill "$DOSBOX_PID" "$XVFB_PID" 2>/dev/null
    exit 1
fi
echo "[entrypoint] DOSBox 視窗 id=$WIN"

# ⚠ 沒有 window manager，xdotool windowactivate 會失敗（不支援 _NET_ACTIVE_WINDOW）。
# 要用 windowfocus（XSetInputFocus）＋**全域** xdotool key（XTest）。
# `xdotool key --window <id>` 走 XSendEvent，SDL 預設不理合成事件——按鍵送了等於沒送。
xdotool windowfocus "$WIN"

run_timeline() {
    IFS=';' read -ra STEPS <<< "$1"
    for step in "${STEPS[@]}"; do
        [[ -z "$step" ]] && continue
        local kind="${step%%:*}" arg="${step#*:}"
        case "$kind" in
            wait) echo "[entrypoint] wait ${arg}s"; sleep "$arg" ;;
            key)  echo "[entrypoint] key $arg";  xdotool windowfocus "$WIN"; xdotool key "$arg" ;;
            type) echo "[entrypoint] type $arg"; xdotool windowfocus "$WIN"; xdotool type --delay 80 "$arg" ;;
            shot) echo "[entrypoint] shot $arg"; import -window root "/shots/${arg}.png" ;;
            *)    echo "[entrypoint] 未知步驟：$step" >&2 ;;
        esac
    done
}

if [[ -n "$TIMELINE" ]]; then
    run_timeline "$TIMELINE"
else
    sleep 6
    import -window root "/shots/default.png"
fi

echo "[entrypoint] 收尾 ..."
kill "$DOSBOX_PID" 2>/dev/null
sleep 1
kill -9 "$DOSBOX_PID" 2>/dev/null
kill "$XVFB_PID" 2>/dev/null

echo "[entrypoint] dosbox.log 最後 20 行："
tail -20 /tmp/dosbox.log
echo "[entrypoint] 完成"

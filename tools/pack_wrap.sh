#!/usr/bin/env bash
# 容器內的第二步：把舞台目錄封成各平台真正在用的格式。
#
#	linux-x64        → AppImage（單檔、雙擊即跑）
#	windows-x64      → zip（每一筆都標 UTF-8，見 tools/pack_zip.py）
#	macos-universal  → zip，裡面是 `.app`（雙擊即跑）
#
# **啟動器要處理「包是唯讀的」這件事。** AppImage 掛起來是唯讀的、`.app` 也不該
# 被寫，而這個遊戲要寫兩樣東西：合成映像（`wl-setup` 的產物）與存檔（原版資料
# 目錄的可寫副本）。所以啟動器把可寫的東西放到使用者的資料目錄，
# 唯讀的資源留在包裡，再把兩邊的路徑一起交給執行檔。
set -euo pipefail
MODE=$1; PLATFORM=$2; STAGE=$3; OUT_ROOT=$4; RUNTIME=${5:-}
ROOT=/src
PKG=$(basename "$STAGE")
OUT="$OUT_ROOT/$PLATFORM"; mkdir -p "$OUT"

# unix_launcher 產生 Linux 與 macOS 共用的那一段。
#
#	$1  包內容的根（shell 變數名已由呼叫端設成 APP）
#	$2  local-full 時原版資料的預設位置（空 ＝ 一定要玩家指定）
unix_launcher() {
    local default=$1
    cat <<EOF
mkdir -p "\$STATE"
if [ \$# -ge 1 ]; then ORIG=\$1; shift; else ORIG="$default"; fi
if [ -z "\$ORIG" ] && [ -f "\$CONF" ]; then ORIG=\$(cat "\$CONF"); fi
EOF
    cat <<'EOF'
if [ -z "$ORIG" ]; then
  ORIG=$(ask_for_data_dir)
fi
if [ -z "$ORIG" ] || [ ! -d "$ORIG" ]; then
  fail "找不到原版資料目錄。用法：$(basename "$0") <含 wl.exe 的原版資料目錄>"
fi
for f in wl.exe wla.bin game1 game2; do
  [ -f "$ORIG/$f" ] || fail "原版資料缺少 $f：$ORIG"
done
printf '%s' "$ORIG" > "$CONF"

# 合成映像：包裡有就用包裡的（完整版），沒有就在使用者的資料目錄現做一次。
IMAGE="$APP/build/wl.merged.exe"
if [ ! -f "$IMAGE" ]; then
  IMAGE="$STATE/build/wl.merged.exe"
  if [ ! -f "$IMAGE" ]; then
    "$APP/bin/wl-setup" -rom "$ORIG" -out "$IMAGE" || fail "產生合成映像失敗"
  fi
fi

# 存檔寫的是**原版資料目錄的可寫副本**，不動玩家自己那一份。
SAVE="$STATE/save-data"
if [ ! -f "$SAVE/game1" ]; then
  mkdir -p "$SAVE"
  cp -f "$ORIG"/* "$SAVE/" 2>/dev/null || true
  chmod -R u+w "$SAVE"
fi

FONT="$APP/eten"
[ -d "$FONT" ] || FONT="${WL_ETEN:-$STATE/eten}"
[ -d "$FONT" ] || FONT=""
MUSIC="$APP/music"
[ -d "$MUSIC" ] || MUSIC=""

exec "$APP/bin/wasteland" \
  -rom "$ORIG" -image "$IMAGE" -save-dir "$SAVE" \
  -lang "$APP/translations/zh-Hant.cat" \
  -paragraphs "$APP/translations/paragraphs-zh-Hant.cat" \
  -refs "$APP/docs/re/generated/paragraph-refs.tsv" \
  -font "$FONT" -music "$MUSIC" -art-root "$APP/artpacks" "$@"
EOF
}

# 開始玩.txt：兩種模式講的話不一樣，因為一種要玩家自備原版、一種不用。
player_readme() { # $1 平台上怎麼啟動
    local howto=$1
    if [ "$MODE" = local-full ]; then
        cat <<EOF
《荒野遊俠》（WASTELAND, 1988）繁體中文重製版 —— 完整版

$howto

原版資料、倚天點陣字型與背景音樂都在包裡，不必另外準備。

⚠ **這一份不可以散布。** 裡面的原版資料是 Interplay／EA 的素材、
倚天字型有授權、背景音樂的波形裡有 Roland MT-32 的 PCM 取樣。
要給別人的是公開版，那一份不含上面任何一項。

按鍵：方向鍵走路，U E O D V S R 是指令列，P 開手札。
F1 說明、F2 設定（V 切換原版／忠實高清／全面重構）、F5／F9 快速存讀檔、F10 離開（會先問，也會先存檔）。
ESC 只取消、退一層，任何一層都不會結束遊戲。

存檔與設定寫在你自己的資料目錄，不會動到包裡的東西。
EOF
    else
        cat <<EOF
《荒野遊俠》（WASTELAND, 1988）繁體中文重製版

$howto

**這一份不含任何原版素材**，要玩需要你自己準備兩樣東西：

1. 合法的原版《Wasteland》資料目錄（裡面有 wl.exe、wla.bin、game1、game2…）
2. 倚天點陣字型目錄（中文顯示要用；沒有就跑英文）

第一次啟動時指到原版資料目錄，程式會自己把 wl.exe 解包並疊上 wla.bin，
產生遊戲要用的合成映像，放在你的資料目錄裡；之後就記住了，不再問。

字型優先放 24 點（STDFONT.24 ＋ SPCFONT.24 ＋ ASCFONT.24）。
放進包旁邊的 eten/ 目錄，或用環境變數 WL_ETEN 指路。

按鍵：方向鍵走路，U E O D V S R 是指令列，P 開手札。
F1 說明、F2 設定（V 切換原版／忠實高清／全面重構）、F5／F9 快速存讀檔、F10 離開（會先問，也會先存檔）。
ESC 只取消、退一層，任何一層都不會結束遊戲。

存檔寫的是原版資料目錄的**副本**，你自己那一份不會被動到。
EOF
    fi
}

case "$PLATFORM" in
linux-x64)
    [ -f "$RUNTIME" ] || { echo "缺少 AppImage runtime：$RUNTIME" >&2; exit 1; }
    APPDIR=/tmp/AppDir; rm -rf "$APPDIR"; mkdir -p "$APPDIR/usr"
    cp -a "$STAGE/." "$APPDIR/usr/"
    # 動態相依：Ebiten 在 Linux 走 cgo（X11／GL／ALSA），把非 glibc 的那幾支帶著走。
    mkdir -p "$APPDIR/usr/lib"
    ldd "$APPDIR/usr/bin/wasteland" | awk '{for (i=1;i<=NF;i++) if ($i ~ /^\//) print $i}' |
    while read -r lib; do
        case "$(basename "$lib")" in
            libc.so.*|libm.so.*|libpthread.so.*|libdl.so.*|librt.so.*|ld-linux*) continue ;;
        esac
        cp -Ln "$lib" "$APPDIR/usr/lib/" 2>/dev/null || true
    done
    player_readme "把這個 AppImage 加上執行權限（chmod +x）之後雙擊，或在終端機直接執行它。
第一次可以在命令列給原版資料目錄：./wasteland-cht-*.AppImage /path/to/wasteland-data" \
        > "$APPDIR/usr/開始遊戲.txt"
    DEFAULT=""; [ "$MODE" = local-full ] && DEFAULT='$APP/data'
    {
        cat <<'EOF'
#!/bin/sh
# AppImage 的進入點。掛載點每次都不一樣，所以路徑一律由這支自己算。
set -eu
HERE=$(cd "$(dirname "$0")" && pwd)
APP="$HERE/usr"
STATE="${XDG_DATA_HOME:-$HOME/.local/share}/wasteland-cht"
CONF="$STATE/data-dir"
export LD_LIBRARY_PATH="$APP/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
fail() { echo "$*" >&2; exit 1; }
# 桌面環境雙擊時沒有終端機可以問，有 zenity 就用它，沒有就請玩家從命令列給。
ask_for_data_dir() {
  if command -v zenity >/dev/null 2>&1; then
    zenity --file-selection --directory \
      --title="選擇《Wasteland》原版資料目錄（裡面要有 wl.exe）" 2>/dev/null || true
  fi
}
EOF
        unix_launcher "$DEFAULT"
    } > "$APPDIR/AppRun"
    chmod +x "$APPDIR/AppRun"
    cat > "$APPDIR/wasteland-cht.desktop" <<'EOF'
[Desktop Entry]
Type=Application
Name=Wasteland 荒野遊俠 繁體中文版
Name[zh_TW]=荒野遊俠（Wasteland）繁體中文版
Comment=Interplay 1988 年作品的繁體中文 remake
Exec=AppRun
Icon=wasteland-cht
Categories=Game;RolePlaying;
Terminal=false
EOF
    # 圖示：用遊戲自己的標題畫面產生一張 PNG 太麻煩（要原版素材），
    # 這裡畫一張純色底 ＋ 字母的 SVG，不牽涉任何原版美術。
    cat > "$APPDIR/wasteland-cht.svg" <<'EOF'
<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256">
  <rect width="256" height="256" fill="#2b1d10"/>
  <rect x="16" y="16" width="224" height="224" fill="none" stroke="#c8a45c" stroke-width="6"/>
  <text x="128" y="118" font-family="monospace" font-size="72" font-weight="bold"
        text-anchor="middle" fill="#c8a45c">WL</text>
  <text x="128" y="176" font-family="monospace" font-size="34"
        text-anchor="middle" fill="#8fbf7f">1988</text>
</svg>
EOF
    ln -sf wasteland-cht.svg "$APPDIR/.DirIcon"
    SFS=/tmp/wasteland.squashfs; rm -f "$SFS"
    mksquashfs "$APPDIR" "$SFS" -root-owned -noappend -no-progress -quiet \
        -comp gzip -b 128K -mkfs-time 0 -all-time 0
    cat "$RUNTIME" "$SFS" > "$OUT/$PKG.AppImage"
    chmod +x "$OUT/$PKG.AppImage"
    rm -rf "$APPDIR" "$SFS"
    echo "[package] $OUT/$PKG.AppImage"
    ;;

windows-x64)
    player_readme "解開之後雙擊 開始遊戲.bat。" > "$STAGE/開始遊戲.txt"
    if [ "$MODE" = local-full ]; then
        WIN_ORIG='if "%~1"=="" (set "ORIG=%ROOT%data") else (set "ORIG=%~1")'
    else
        WIN_ORIG='if "%~1"=="" (if exist "%STATE%\data-dir.txt" (set /p ORIG=<"%STATE%\data-dir.txt") else (set "ORIG=")) else (set "ORIG=%~1")'
    fi
    {
        cat <<'EOF'
@echo off
setlocal enabledelayedexpansion
set "ROOT=%~dp0"
set "STATE=%LOCALAPPDATA%\wasteland-cht"
if not exist "%STATE%" mkdir "%STATE%"
EOF
        printf '%s\n' "$WIN_ORIG"
        cat <<'EOF'
if "%ORIG%"=="" (
  echo 用法：開始遊戲.bat ^<原版資料目錄^>
  echo （該目錄裡要有 wl.exe、wla.bin、game1、game2）
  pause
  exit /b 1
)
if not exist "%ORIG%\wl.exe" (echo 原版資料缺少 wl.exe：%ORIG% & pause & exit /b 1)
if not exist "%ORIG%\wla.bin" (echo 原版資料缺少 wla.bin：%ORIG% & pause & exit /b 1)
>"%STATE%\data-dir.txt" echo %ORIG%
set "IMAGE=%ROOT%build\wl.merged.exe"
if not exist "%IMAGE%" (
  set "IMAGE=%STATE%\build\wl.merged.exe"
  if not exist "!IMAGE!" "%ROOT%bin\wl-setup.exe" -rom "%ORIG%" -out "!IMAGE!"
)
set "SAVE=%STATE%\save-data"
if not exist "%SAVE%\game1" (
  if not exist "%SAVE%" mkdir "%SAVE%"
  copy /y "%ORIG%\*" "%SAVE%" >nul
)
set "FONT=%ROOT%eten"
if not exist "%FONT%" set "FONT=%ROOT%..\eten"
if not exist "%FONT%" set "FONT="
set "MUSIC=%ROOT%music"
if not exist "%MUSIC%" set "MUSIC="
"%ROOT%bin\wasteland.exe" -rom "%ORIG%" -image "!IMAGE!" -save-dir "%SAVE%" ^
  -lang "%ROOT%translations\zh-Hant.cat" ^
  -paragraphs "%ROOT%translations\paragraphs-zh-Hant.cat" ^
  -refs "%ROOT%docs\re\generated\paragraph-refs.tsv" ^
  -font "%FONT%" -music "%MUSIC%" -art-root "%ROOT%artpacks"
EOF
    } > "$STAGE/開始遊戲.bat"
    # batch 檔在 Windows 上要 CRLF，不然 `^` 續行與 `if` 區塊會被吃掉。
    python3 - "$STAGE/開始遊戲.bat" <<'PY'
import sys
p = sys.argv[1]
b = open(p, 'rb').read().replace(b'\r\n', b'\n').replace(b'\n', b'\r\n')
open(p, 'wb').write(b)
PY
    # DLL 清單由**實際的 PE 匯入表**產生，不是抄來的。
    {
        cat <<'EOF'
這個包不夾帶任何 DLL —— 下面說明為什麼，以及缺了東西會怎樣。

bin\wasteland.exe 的 PE 匯入表（由 tools/pe_imports.py 從這個包裡的執行檔實際讀出）：

EOF
        python3 "$ROOT/tools/pe_imports.py" "$STAGE/bin/wasteland.exe" | sed 's/^/  /'
        cat <<'EOF'

執行檔是純 Go 編的（CGO_ENABLED=0），沒有 C 執行期，所以不需要 MSVC
redistributable，也沒有第三方 DLL 可以夾帶。

其餘的 DLL 由 Ebiten 在執行時視情況用 LoadLibrary 載入，全部是 Windows
自己的系統元件，隨作業系統安裝：

  d3d11.dll、dxgi.dll、d3dcompiler_47.dll        DirectX 11 繪圖路徑
  opengl32.dll                                   OpenGL 繪圖路徑（DirectX 不成時的退路）
  user32.dll、gdi32.dll、imm32.dll、shcore.dll   視窗、輸入法與 DPI
  winmm.dll、ole32.dll                           音訊與 COM
  xinput1_4.dll、xinput9_1_0.dll、dinput8.dll    搖桿

其中只有 d3dcompiler_47.dll 有機會缺席（它屬微軟，隨 Windows 10／11 內建，
本專案不代為散布）。真的缺了，Ebiten 會自動退回 OpenGL，遊戲照樣跑。
EOF
    } > "$STAGE/Windows-DLL說明.txt"
    python3 "$ROOT/tools/pack_zip.py" "$STAGE" "$OUT/$PKG.zip"
    python3 "$ROOT/tools/verify_windows_zip.py" "$OUT/$PKG.zip"
    ;;

macos-universal)
    APPROOT=/tmp/macpkg; rm -rf "$APPROOT"
    BUNDLE="$APPROOT/$PKG/荒野遊俠.app"
    mkdir -p "$BUNDLE/Contents/MacOS" "$BUNDLE/Contents/Resources"
    cp -a "$STAGE/." "$BUNDLE/Contents/Resources/"
    VER=$(git -C "$ROOT" rev-list --count HEAD)
    cat > "$BUNDLE/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key><string>Wasteland CHT</string>
  <key>CFBundleDisplayName</key><string>荒野遊俠 繁體中文版</string>
  <key>CFBundleIdentifier</key><string>io.github.wicanr2.wastelandcht</string>
  <key>CFBundleExecutable</key><string>wasteland-cht</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleShortVersionString</key><string>0.$VER</string>
  <key>CFBundleVersion</key><string>$VER</string>
  <key>LSMinimumSystemVersion</key><string>11.0</string>
  <key>LSApplicationCategoryType</key><string>public.app-category.role-playing-games</string>
  <key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
EOF
    DEFAULT=""; [ "$MODE" = local-full ] && DEFAULT='$APP/data'
    {
        cat <<'EOF'
#!/bin/sh
# `.app` 的進入點。雙擊時沒有終端機，所以錯誤用對話框講，
# 資料目錄用選擇器問 —— 問過一次就記在設定檔裡，之後不再問。
set -eu
HERE=$(cd "$(dirname "$0")" && pwd)
APP="$HERE/../Resources"
STATE="$HOME/Library/Application Support/wasteland-cht"
CONF="$STATE/data-dir"
fail() {
  if [ -t 2 ]; then echo "$*" >&2; else
    osascript -e "display dialog \"$*\" buttons {\"好\"} with icon stop" >/dev/null 2>&1 || true
  fi
  exit 1
}
ask_for_data_dir() {
  osascript -e 'POSIX path of (choose folder with prompt "選擇《Wasteland》原版資料目錄（裡面要有 wl.exe）")' 2>/dev/null || true
}
EOF
        unix_launcher "$DEFAULT"
    } > "$BUNDLE/Contents/MacOS/wasteland-cht"
    chmod +x "$BUNDLE/Contents/MacOS/wasteland-cht"
    player_readme "雙擊 荒野遊俠.app。第一次會跳出資料夾選擇器（公開版），選過一次就記住了。" \
        > "$APPROOT/$PKG/開始遊戲.txt"
    cat >> "$APPROOT/$PKG/開始遊戲.txt" <<'EOF'

這個 app 沒有 Apple 開發者簽章，也沒有經過公證（notarization），第一次打開
系統會擋。兩種放行方式擇一：

  在「系統設定 → 隱私權與安全性」按「仍要打開」，或在終端機執行
  xattr -dr com.apple.quarantine 荒野遊俠.app

想從終端機啟動並自己指定資料目錄：

  ./荒野遊俠.app/Contents/MacOS/wasteland-cht /path/to/wasteland-data

執行檔是 universal（x86_64 ＋ arm64），Intel 與 Apple Silicon 都能跑。
可以用 lipo -info 荒野遊俠.app/Contents/Resources/bin/wasteland 自己確認。
EOF
    python3 "$ROOT/tools/pack_zip.py" "$APPROOT/$PKG" "$OUT/$PKG.zip"
    rm -rf "$APPROOT"
    ;;
*) echo "未知平台 $PLATFORM" >&2; exit 2 ;;
esac

# 舞台目錄裡有原版資料（local-full），封完就清掉，不要留在磁碟上。
rm -rf "$(dirname "$STAGE")"

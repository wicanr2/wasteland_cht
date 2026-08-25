#!/usr/bin/env bash
# 容器內的第一步：編出執行檔、把要帶的檔案排成一份「舞台目錄」。
#
# **這一步不產生最終檔案。** AppImage／zip 由 `tools/pack_wrap.sh` 在封裝用的
# image 內接手——理由寫在那支的開頭。
set -euo pipefail
MODE=$1; PLATFORM=$2; OUT_ROOT=$3; INPUT=${4:-}; FONT_INPUT=${5:-}; MUSIC_INPUT=${6:-}
ROOT=/src
GO=/usr/local/go/bin/go

case "$PLATFORM" in
    linux-x64)       GOOS=linux;   GOARCH=amd64; CGO=1; EXE= ;;
    windows-x64)     GOOS=windows; GOARCH=amd64; CGO=0; EXE=.exe ;;
    macos-universal) GOOS=darwin;  GOARCH=universal; CGO=1; EXE=; MAC=1 ;;
    *) echo "未知平台 $PLATFORM" >&2; exit 2 ;;
esac

STAMP=$(git -C "$ROOT" rev-parse --short=12 HEAD)
PKG="wasteland-cht-${PLATFORM}-${MODE}-${STAMP}"
STAGE="$OUT_ROOT/.stage"; D="$STAGE/$PKG"
rm -rf "$STAGE"
mkdir -p "$D/bin" "$D/translations" "$D/docs/re/generated" "$D/artpacks"

build() { # $1 輸出 $2 套件
    GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED="$CGO" \
        "$GO" build -trimpath -buildvcs=false -ldflags="-s -w" -o "$1" "$2"
}

if [ "${MAC:-0}" = 1 ]; then
    # macOS 沒有交叉編的「一個 GOARCH」——兩個架構各編一次，再 lipo 合成。
    command -v o64-clang  >/dev/null || { echo '缺少 o64-clang（osxcross）' >&2; exit 78; }
    command -v oa64-clang >/dev/null || { echo '缺少 oa64-clang（osxcross）' >&2; exit 78; }
    LIPO=$(command -v lipo || command -v x86_64-apple-darwin24.5-lipo || true)
    [ -n "$LIPO" ] || { echo '缺少 lipo' >&2; exit 78; }
    SDK=$(ls -d /osxcross/SDK/MacOSX*.sdk 2>/dev/null | head -1)
    [ -n "$SDK" ] || { echo '缺少 macOS SDK' >&2; exit 78; }
    mac_build() { # $1 GOARCH $2 CC $3 輸出 $4 套件
        GOOS=darwin GOARCH="$1" CGO_ENABLED=1 CC="$2" \
            "$GO" build -trimpath -buildvcs=false -ldflags="-s -w" -o "$3" "$4"
    }
    for pair in "wasteland ./cmd/wasteland" "wl-setup ./cmd/wl-setup"; do
        set -- $pair
        mac_build amd64 o64-clang  "/tmp/$1-amd64" "$2"
        mac_build arm64 oa64-clang "/tmp/$1-arm64" "$2"
        "$LIPO" -create -output "$D/bin/$1" "/tmp/$1-amd64" "/tmp/$1-arm64"
        INFO=$("$LIPO" -info "$D/bin/$1")
        case "$INFO" in
            *x86_64*arm64*|*arm64*x86_64*) ;;
            *) echo "$1 不是 x86_64 + arm64 universal：$INFO" >&2; exit 1 ;;
        esac
    done
else
    build "$D/bin/wasteland$EXE" ./cmd/wasteland
    build "$D/bin/wl-setup$EXE"  ./cmd/wl-setup
fi

# —— 翻譯與文件（兩種模式都帶）——————————————————————————————
cp "$ROOT/translations/zh-Hant.cat" "$ROOT/translations/paragraphs-zh-Hant.cat" "$D/translations/"
cp "$ROOT/translations/glossary.md" "$ROOT/translations/README.md" "$D/translations/"
cp "$ROOT/docs/re/generated/paragraph-refs.tsv" "$D/docs/re/generated/"
for d in manual manual-cht paragraphs walkthrough; do
    cp -r "$ROOT/docs/$d" "$D/docs/$d"
done
cp "$ROOT/README.md" "$D/docs/README-專案.md"

# —— remake 自製新版美術（兩種包都可散布）——————————————————————
# 這兩包不含玩家自備的 ROM／原版檔；啟動器以絕對路徑傳入，避免桌面啟動時
# 工作目錄不同而只能看到 original。
cp -a "$ROOT/artpacks/faithful-hd" "$ROOT/artpacks/reimagined" "$D/artpacks/"

# —— local-full：原版素材 ————————————————————————————————————
if [ "$MODE" = local-full ]; then
    [ -n "$INPUT" ] || { echo 'local-full 缺少原版資料掛載' >&2; exit 1; }
    mkdir -p "$D/data"; cp -a "$INPUT/." "$D/data/"
    # 合成映像在容器裡現做：與玩家自己跑 `wl-setup` 走的是同一支程式碼。
    mkdir -p "$D/build"
    if [ "${MAC:-0}" = 1 ] || [ "$PLATFORM" = windows-x64 ]; then
        # 目標平台的執行檔在這裡跑不動，另外編一支本機用的。
        GOOS=linux GOARCH=amd64 CGO_ENABLED=0 "$GO" build -trimpath -buildvcs=false \
            -o /tmp/wl-setup-host ./cmd/wl-setup
        /tmp/wl-setup-host -rom "$D/data" -out "$D/build/wl.merged.exe"
    else
        "$D/bin/wl-setup" -rom "$D/data" -out "$D/build/wl.merged.exe"
    fi
    if [ -n "$FONT_INPUT" ]; then mkdir -p "$D/eten"; cp -a "$FONT_INPUT/." "$D/eten/"; fi
	if [ -n "$MUSIC_INPUT" ]; then
		mkdir -p "$D/music"
		# 正式包只保留明確的 retro/、modern/，避免舊平面檔讓驗收誤判成
		# 單一配樂。只有舊工作區沒有 retro/ 時才把平面檔遷入 retro/。
		for variant in retro modern; do
			if [ -d "$MUSIC_INPUT/$variant" ]; then
				mkdir -p "$D/music/$variant"
				cp -a "$MUSIC_INPUT/$variant/." "$D/music/$variant/"
			fi
		done
		if [ ! -d "$D/music/retro" ]; then
			mkdir -p "$D/music/retro"
			find "$MUSIC_INPUT" -maxdepth 1 \( -name '*.ogg' -o -name '*.mid' \) \
				-exec cp {} "$D/music/retro/" \;
		fi
		for variant in retro modern; do
			count=$(find "$D/music/$variant" -maxdepth 1 -name '*.ogg' 2>/dev/null | wc -l)
			[ "$count" -eq 10 ] || { echo "$variant 配樂應有 10 首，實際 $count" >&2; exit 1; }
		done
		[ -f "$D/music/modern/FluidR3-GM-LICENSE.txt" ] || {
			echo 'modern 配樂缺少 FluidR3 GM 權利文件' >&2; exit 1;
		}
	fi
fi

printf 'Package: %s\nCommit: %s\nMode: %s\nPlatform: %s\nBuild image: %s\n' \
    "$PKG" "$(git -C "$ROOT" rev-parse HEAD)" "$MODE" "$PLATFORM" "${WL_BUILD_IMAGE:-unknown}" \
    > "$D/PACKAGE-MANIFEST.txt"
if [ "${MAC:-0}" = 1 ]; then
    printf 'SDK: %s\nArchitectures: x86_64 arm64 (lipo universal)\n' "$SDK" >> "$D/PACKAGE-MANIFEST.txt"
fi

# —— 公開包的守門員 ————————————————————————————————————————
#
# **靠自律不夠**：兩種模式的舞台目錄長得幾乎一樣，複製錯一次就散布出去了。
# 這裡不是抽查，是「凡是原版素材的形狀一律拒絕」。
if [ "$MODE" = public ]; then
    bad=$(find "$D" -type f \( \
        -iname 'wl*.exe' -o -iname 'wla.bin' -o -iname 'game1' -o -iname 'game2' \
        -o -iname 'allpics*' -o -iname 'allhtds*' -o -iname '*.pic' -o -iname '*.cpa' \
        -o -iname '*.fnt' -o -iname '*.wlf' -o -iname 'colorf*' -o -iname 'curs' \
        -o -iname 'transtbl' -o -iname 'paragraphs.txt' -o -iname '*.ogg' \
        -o -iname 'STDFONT.*' -o -iname 'SPCFONT.*' -o -iname 'ASCFONT.*' \) \
        -not -name 'wl-setup*' -not -name 'wasteland*' -print)
    if [ -n "$bad" ]; then
        echo '公開包裡有不可散布的東西：' >&2
        printf '%s\n' "$bad" >&2
        exit 1
    fi
    for d in data eten music build; do
        [ -e "$D/$d" ] && { echo "公開包不該有 $d/" >&2; exit 1; }
    done
fi
echo "[stage] $D"

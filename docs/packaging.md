# 三平台封裝：怎麼跑、每個決定為什麼這樣

```bash
tools/package.sh <public|local-full> <linux-x64|windows-x64|macos-universal>
```

產物落在 `dist-all/<模式>/<平台>/`（gitignore）。六個組合各跑一次就是完整一輪，
而 `tools/dist.sh` 就是「跑滿六個，再把配樂與推廣片收進同一棵樹」的那一支。

---

## 1. 兩種模式的差別只有一件事

| 模式 | 內容 | 可以給別人嗎 |
|---|---|---|
| `public` | 引擎、`wl-setup`、翻譯、文件 | **可以**——不含任何原版素材 |
| `local-full` | 上面全部 ＋ 原版資料 ＋ 倚天字型 ＋ 背景音樂 ＋ 合成映像 | **不可以**（`CLAUDE.md` §7）|

`public` 的守門員寫在 `tools/package_container.sh` 結尾：凡是原版素材的形狀
（`wl*.exe`、`game1`、`allpics*`、`*.fnt`、`*.wlf`、`STDFONT.*`、`*.ogg`…）
一律拒絕，`data/`／`eten/`／`music/`／`build/` 四個目錄存在就中止。
**靠自律不夠**——兩種模式的舞台目錄長得幾乎一樣，複製錯一次就散布出去了。

## 2. 玩家不必裝 Python：`wl-setup`

遊戲要的是**合成映像** `wl.merged.exe`（EXEPACK 解包 ＋ 疊 `wla.bin`），
字型、九張字串表與資源定址常數都在裡面（`docs/re/02`、`docs/re/03` §5）。
那是原版執行檔的衍生物，公開包不能夾帶，只能玩家自己產生。

原本這一步是兩支 Python 腳本，而 **Windows 預設沒有 Python**。
所以移植成 Go：`internal/exepack` ＋ `cmd/wl-setup`，跟著包一起走，
啟動器第一次執行時自己跑一次。

驗收條件是 **byte-exact**：`internal/exepack/exepack_test.go` 拿玩家自己那份
`wl.exe` 跑一次，SHA-256 必須等於本專案分析用的那份
（`cd5b07ea…`）。跑不出同一個雜湊就是移植錯了，不是「大概對」。

⚠ 別的發行版本雜湊本來就會不一樣，所以 `wl-setup` 對不上時**只提醒不中止**。

## 3. 三個平台各自的坑

### Linux：AppImage

- runtime 是上游那顆固定的靜態 ELF，**雜湊寫死在 `tools/package.sh`**——
  換了要有人明確改那一行，不會靜悄悄換掉。
- Ebiten 在 Linux 走 cgo（X11／GL／ALSA），所以把**非 glibc 的直接相依**
  一起帶進 `usr/lib`，`AppRun` 設 `LD_LIBRARY_PATH` 優先用私有的那份。
- 掛載點每次都不一樣，路徑一律由 `AppRun` 自己算。

### Windows：zip，而且 UTF-8 旗標要自己點亮

`zip -r` 不夠：Info-ZIP 只在檔名超出 CP437 時才打開 general purpose bit 11，
**全 ASCII 的檔名不會標**。沒標的話解壓端只能用系統預設編碼猜——
在繁中 Windows 上就是 CP950，包裡只要出現一個中文檔名就會變亂碼。
`tools/pack_zip.py` 對**每一筆**都點亮，寫完再回頭改位元組
（`ZipFile` 會在寫入時把 `flag_bits` 歸零），中央目錄與區域檔頭兩處都改，
最後自己驗一次「每一筆都有旗標」，少一筆就當場失敗。

這個包有四個中文檔名（`開始遊戲.bat`、`開始遊戲.txt`、`Windows-DLL說明.txt`…），
所以這件事不是理論問題。

- 執行檔是 `CGO_ENABLED=0` 編的，**PE 匯入表只有系統 DLL**，
  清單由 `tools/pe_imports.py` 從包裡那支執行檔實際讀出來，不是抄的。
- `.bat` 一定要 CRLF，否則 `^` 續行與 `if` 區塊會被吃掉。

### macOS：osxcross 交叉編 ＋ lipo

沒有 Mac 也出得了 macOS 版：`wolong-osxcross-go` 這個 image 裡有
o64-clang／oa64-clang 與 MacOSX SDK。**兩個架構各編一次再 `lipo` 合成**
——`GOARCH` 沒有「universal」這個值。

`.app` 的進入點是一支 shell：雙擊時沒有終端機，所以錯誤用 `osascript`
的對話框講，資料目錄用選擇器問，問過一次就記在設定檔裡。

⚠ **沒有 Apple 簽章也沒有公證**，第一次打開系統會擋；放行方式寫在包裡的
`開始遊戲.txt`（系統設定放行，或 `xattr -dr com.apple.quarantine`）。

## 4. 啟動器要處理「包是唯讀的」

AppImage 掛起來是唯讀的、`.app` 也不該被寫，而這個遊戲要寫兩樣東西：

| 要寫的 | 放哪 |
|---|---|
| 合成映像（`wl-setup` 的產物）| `<資料目錄>/build/wl.merged.exe` |
| 存檔（**原版資料目錄的可寫副本**）| `<資料目錄>/save-data/` |

`<資料目錄>` 是 `$XDG_DATA_HOME/wasteland-cht`（Linux）、
`~/Library/Application Support/wasteland-cht`（macOS）、
`%LOCALAPPDATA%\wasteland-cht`（Windows）。

**存檔寫的是副本，玩家自己那一份原版資料不會被動到**——
原版的存檔就在 `game1`／`game2` 檔尾（`docs/re/30`），直接寫回去等於改到原版檔案。

## 5. 一輪跑完長這樣

| 模式 | 平台 | 產物 | 大小 |
|---|---|---|---:|
| public | linux-x64 | `.AppImage` | 6.5 MB |
| public | windows-x64 | `.zip`（76 筆）| 4.2 MB |
| public | macos-universal | `.zip`（85 筆，內含 `.app`）| 7.8 MB |
| local-full | linux-x64 | `.AppImage` | 24 MB |
| local-full | windows-x64 | `.zip`（121 筆）| 21 MB |
| local-full | macos-universal | `.zip`（125 筆）| 25 MB |

## 6. 交付物集中在 dist-all

```
dist-all/
  public/<平台>/      可散布的三個包
  local-full/<平台>/  不可散布的三個包
  music/{midi,ogg}/   配樂（midi 可散布、ogg 不可）
  promo/              推廣片
  SHA256SUMS          全部檔案的校驗碼
```

一次跑完：`tools/dist.sh`。包已經跑過、只想重收音樂與影片：`tools/dist.sh --skip-packages`。

## 7. 要用到的 image

```bash
docker build --network host -t wasteland-go:1.24-x11  -f docker/wasteland-go.Dockerfile  docker/
docker build --network host -t wasteland-pkg:1.24-x11 -f docker/wasteland-pkg.Dockerfile docker/
# macOS 那條路另外要 osxcross ＋ Go 的 image，預設吃
# wolong-osxcross-go:20260811-event10-r4，可用 WL_OSXCROSS_IMAGE 換
```

封裝分兩步（`package_container.sh` 排舞台 → `pack_wrap.sh` 封裝）的理由：
macOS 的編譯要 osxcross 那個 image，而 AppImage 要的 `mksquashfs` 只裝在
`wasteland-pkg`，兩者不在同一個 image 裡。

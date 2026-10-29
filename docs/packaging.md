# 三平台本機封裝

版號必須完整指定為 `v.<主>.<次>.<修訂>-YYYYMMDD`。
`v.1.0.0-20261029` 是先前本機測試包。本輪依同日新提交的版號契約升為
`v.1.0.1-20261029`，加入已確認的新介面，不建立 tag，也不上傳或發布。
舊版交付目錄保留，不作為新版入口。

```bash
tools/package.sh v.1.0.1-20261029 patch linux-x64
tools/package.sh v.1.0.1-20261029 full-local windows-x64
tools/package.sh v.1.0.1-20261029 full-local macos-universal
# 依序建立全部六包，再驗證並產生交付清單
tools/dist.sh v.1.0.1-20261029
```

同版本同平台的包已存在時會拒絕覆寫。修正後另定新版號，或在尚未交付前明確保留
失敗候選；不刪除舊發行物來重跑。主機腳本只組織 Docker 呼叫，分析、編譯、封裝、
校驗與啟動驗證都在容器內進行。原版、字型、音樂與工作樹輸入採唯讀掛載，
只有明確的暫存、快取與交付目錄可寫。

## 內容與位置

| 目錄 | 內容與邊界 |
|---|---|
| `dist-all/<版本>/patch/<平台>/` | 引擎、`wl-setup`、翻譯、內建開放字型與授權；不含原版資料、倚天、音樂或 HD2 圖片 |
| `dist-all/<版本>/full-local/<平台>/` | 另含本機合法原版、倚天字型、兩套十首配樂及完整 HD2 A／B；不散布、不加入 Git |
| `dist-all/<版本>/smoke/` | 逐包內容／架構收據、啟動畫面與實測限制 |
| `dist-all/<版本>/SHA256SUMS.json` | 六包與驗證收據的大小、SHA-256、版本及權利分類 |

平台名稱為 `linux-x64`、`windows-x64`、`macos-universal`。
建置中間資料在 `dist-build/<版本>/`，不是交付入口。舊的
`dist-all/public/`、`dist-all/local-full/`、音樂與影片目錄只作歷史保留。
本輪不把歷史影片複製成新版影片，也不改變公開素材的授權範圍。

## 工具與重現

[主機入口](../tools/package.sh)分兩步呼叫
[容器建置](../tools/package_container.sh)與[容器封裝](../tools/pack_wrap.sh)。
沿用本機已有且已驗證的共用工具映像，不另建重複映像：

| 用途 | 目前映像 | 內容 |
|---|---|---|
| Linux／Windows 建置與封裝 | `san1-go:1.26.7-ebiten2.9.9-r1` | Go 1.26.7、X11／GL、squashfs、Python、Xvfb |
| macOS 交叉編譯 | `cok-osxcross:go1.24.13-sdk15.5-r1` | 專案指定的 Go 1.24.13、osxcross、MacOSX15.5 SDK、lipo，明確設定 macOS 11 |
| Windows 補充試玩 | `cok-wine-verify:noble-r1` | Wine、Xvfb；僅用本輪專屬前綴與 X11 socket |

可用 `WL_GO_IMAGE`、`WL_PKG_IMAGE`、`WL_OSXCROSS_IMAGE` 指定相容工具映像。
包內 manifest 記錄實際建置映像，交付清單另記映像 ID。所有工作均使用
`--network none`、UID/GID、資源限制與外層逾時，模組取自既有鎖版快取。

AppImage runtime 已封存在 `workplace/appimage/runtime-x86_64`，SHA-256 為
`1cc49bcf1e2ccd593c379adb17c9f85a36d619088296504de95b1d06215aebbf`。
封裝前在容器驗雜湊，缺少時停止；不在主機下載或自動改用浮動新版。
原始來源為 AppImage type2-runtime 的歷史 continuous 下載，封存副本是本輪重現輸入。

## 版本與啟動

兩個執行檔都由 linker 注入同一完整版本，`-version` 可查詢；遊戲視窗標題也會顯示。
封包檔名、包內 manifest 與交付根目錄使用相同版號。macOS 的數值欄位分別保存
版本核心與建置日期，`WastelandReleaseVersion` 保存完整版本字串。

Linux 為 AppImage，非 glibc 的直接相依放入私有 `usr/lib`。Windows ZIP 的每一筆
中央與區域檔頭都設 UTF-8 旗標，批次檔使用 CRLF。macOS 分別編出 x86_64／arm64，
再以 lipo 合成，ZIP 保留執行權限；沒有 Apple 簽章或公證。

啟動器從唯讀包定位資源，將原版存檔副本與合成映像放入使用者資料目錄：

| 平台 | 可寫資料目錄 |
|---|---|
| Linux | `$XDG_DATA_HOME/wasteland-cht`，未設時為 `~/.local/share/wasteland-cht` |
| Windows | `%LOCALAPPDATA%\wasteland-cht` |
| macOS | `~/Library/Application Support/wasteland-cht` |

`patch` 第一次啟動需要玩家的合法原版，透過包內 `wl-setup` 建立合成映像。
倚天字型與 HD2 衍生圖片不在公開包中。`full-local` 可直接使用包內來源。
兩類包均內建源泉圓體與 Noto，預設 18 像素新介面，不需要倚天。
原始字型與授權 SHA-256、包內 `UI: rounded-v1` 標記均由內容驗證器核對。

實際封包的 GUI 抽測使用 [check_package_gui.sh](../tools/check_package_gui.sh)，必須在既有
Go／X11 Docker 工具映像內執行，參數為模式、完整版本及平台。Linux 直接啟動 AppImage；
Windows 使用 [Wine 伺服器](../tools/check_package_wine_server.sh)從實際中文批次檔開場，
兩個容器只共用本輪 X11 socket，Wine 前綴隔離。全尺寸視窗與 960×720 邏輯圖同時保存，
滑鼠位置由實際視窗尺寸換算。收據綁定封包雜湊、字型工具列、A／B、存讀檔與正常退出。
證據目錄另保存當次使用的腳本，交付清單核對副本與收據中的腳本雜湊。
每次操作後先確認實際視窗更新，再送下一個輸入；避免軟體渲染下固定延遲造成漏按或截到舊畫面。
GUI 工作依序執行。容器掛載如下，來源須先確認存在：

| 容器路徑 | 來源與權限 |
|---|---|
| `/src` | 專案根目錄，唯讀 |
| `/evidence` | `dist-all/<版本>/smoke/`，UID/GID 對應的使用者可寫 |
| `/pkg` | Windows ZIP 的實際解壓根目錄，唯讀；僅 Wine 伺服器使用 |
| `/wine` | 本輪專屬 Wine 前綴，伺服器可寫、操作容器唯讀 |
| `/tmp/.X11-unix` | 本輪專屬 socket 目錄，伺服器可寫、操作容器唯讀 |

Linux 在 Go 映像內執行 `bash tools/check_package_gui.sh full-local <版本> linux-x64`，
`patch` 以同一入口另跑。Windows 先在 Wine 映像執行
`bash tools/check_package_wine_server.sh <版本>`，再於 Go 映像執行
`bash tools/check_package_gui.sh full-local <版本> windows-x64`。
兩個 Windows 容器使用相同的 `/evidence` 與專屬 socket，伺服器採 `C.utf8` locale。
容器須另加網路隔離、資源限制、UID/GID、`--rm` 及外層逾時；完成後移除專用前綴與 socket。
一般 S 與 F5／F9 的存檔入口不同，WLQS v2 的相容性見專案存檔規格。
F5／F9 沿用 `DefaultQuickSavePath`：未設 `XDG_DATA_HOME` 時使用
`<使用者家目錄>/.local/share/wasteland-cht/quicksave.wlq`，Windows 也是如此。
不要把它與上表的一般 S 存檔副本目錄混為一談。

## 驗證入口

[逐包驗證器](../tools/verify_package.py)直接解開實際產物，核對版本、ELF／PE／
Mach-O 架構、ZIP CRC／UTF-8／權限、翻譯、授權、原版雜湊、兩包各 1,157 張美術
及兩套各十首配樂。`patch` 額外拒收原版資料、PNG、音樂與非空美術包。
[交付清單工具](../tools/finalize_dist.py)重查六包並產生收據與 SHA-256 清單。

```bash
python3 tools/verify_package.py <封包路徑> --version <版本> --mode <patch或full-local> \
  --platform <平台> --receipt <收據路徑>
python3 tools/finalize_dist.py <版本>
```

上述 Python 命令在封裝容器內執行。完整建置與檢查不是原生平台試玩的替代品。
清單工具只在 GUI 收據的封包雜湊、實際圖片雜湊與模式驗證旗標相符時列為通過；
沒有收據時列為未執行。新版使用上述兩份版本參數化腳本與該版 `smoke/`；
`workplace/art-preview/v100-*.sh` 只重現先前版本，不作新版入口。
本機 Linux GUI、Wine 的 Windows 補充檢查與 macOS 結構驗證分開記錄，未執行的
原生 Windows／macOS 實機項目不得標成通過。每輪結果見 [WORKLOG](../WORKLOG.md)。

# 30：新版美術主題與動態畫布規格 — **IMPLEMENTED**

> 本規格是 remake 自訂功能，不是原版逆向結論。原版資產身分與幾何仍以
> [`01-assets-and-formats.md`](01-assets-and-formats.md) 及 `docs/re/23`、`24` 為準。

## 1. 已確認的產品決策（2026-08-21）

- 同時保留三種呈現：`original`、`faithful-hd`、`reimagined`。
- `faithful-hd` 維持原版 4:3 構圖：16×16 圖磚的邏輯格改用 48×48 資產，
  96×84 場景改用 288×252 資產；規則、視野與點擊語意不變。
- `reimagined` 使用固定 3/4 俯視鏡頭、16:9 響應式 HUD。視窗可增加可見範圍，
  但必須受 manifest 的最大視野限制；偵測、遭遇與事件仍使用原版規則範圍。
- `original` 與 `faithful-hd` 只等比例重繪並維持 4:3，不重新排版。
- 玩家可在遊戲內切換；目標包必須先完整驗證，成功才一次替換。失敗時維持
  目前模式與遊戲狀態，禁止半套資產混用。
- `faithful-hd` 使用單一隊伍圖示；`reimagined` 顯示最多七名隊員，但規則上全隊
  仍只佔一格。外觀由既有角色資料穩定導出，不擴充原版存檔 schema。
- 正式 sprite 包含 8 方向待機、行走、攻擊、受傷、倒下、互動與多武器疊層；
  忠實模式的隊伍圖示至少提供 4 方向動畫。
- 82 張 `ALLPICS` 場景使用分層資產；重要人物、出口、物件與事件提示不可改變。
  重構模式可增加純裝飾內容，但不得暗示不存在的互動。
- 新版素材全程由 AI 產生及整理。Git 保存提示詞、工具／模型、生成日期、雜湊、
  manifest 與驗收紀錄；大型來源圖及中間產物不入版控。

## 2. 邊界

```text
原版 ROM ── internal/assets ── 原版索引資產 ── original renderer

artpacks/<id>/manifest.json ── internal/artpack ── 已驗證的完整 Bundle
                                             ├── faithful-hd renderer
                                             └── reimagined renderer

game.World / Save ───────────────────────────┴──（三模式共用，不複製）
```

`internal/artpack` 只驗證檔案、尺寸、雜湊與資產覆蓋率，不認識 Ebiten 或遊戲規則。
呈現層只能持有一個已驗證 `Bundle`；切換不能逐檔修改現有 renderer。

## 3. Manifest 最小契約

- `schema`：目前固定為 `1`。
- `id` 與 `mode`：`mode` 只能是 `faithful-hd` 或 `reimagined`；`original` 不使用外部包。
- `canvas`：基準寬高、是否響應式、最大可見格數。
- `assets[]`：穩定 ID、種類、相對路徑、像素尺寸、SHA-256。
- 路徑必須是包目錄內的相對路徑，不接受絕對路徑或 `..`。
- PNG 必須能解碼，實際尺寸必須與 manifest 相同，雜湊必須逐檔吻合。
- 同一資產 ID 不得重複；任何錯誤都讓整包載入失敗。

完整性 gate 已由 checked-in bundle 測試覆蓋：

- 9 組 tileset 的實際索引覆蓋率；
- 10 個疊圖、遮罩及 animation set；
- 82 張場景、標題與結局；
- 分層場景的必需 layer 及循環長度；
- 角色方向、動作、影格與武器 layer 的笛卡兒積。

## 4. 動態畫布契約

- `original`／`faithful-hd`：邏輯比例固定 4:3；輸出使用 contain，不裁切。
- `reimagined`：`Layout(outsideWidth, outsideHeight)` 依實際視窗重算 16:9 安全區、HUD、
  地圖 viewport 與輸入逆轉換；不得只放大上一幀 framebuffer。
- 地圖 viewport 可增加但不得超過 manifest 的 `max_view_cols`／`max_view_rows`。
- 小視窗低於最小可讀尺寸時縮放完整 HUD，不隱藏遊戲必要控制。
- 模式切換不重建 `game.World`、RNG、戰鬥、設施、時鐘或存檔物件。

## 5. AI 來源紀錄

每個正式資產 ID 都要能追到：提示詞檔、產生工具／模型、UTC 時間、來源圖雜湊、
後處理版本、輸出 PNG 雜湊及人工／自動驗收結果。禁止在 metadata 宣稱模仿特定
在世藝術家；禁止混入原版逐像素放大後冒充新繪素材。

## 6. 完整驗收

1. 合法最小 bundle 可在無頭測試中載入。
2. 路徑穿越、錯誤尺寸、錯誤雜湊、重複 ID、未知 mode 或非 PNG 都失敗。
3. 載入失敗不改變呼叫端目前 bundle（由切換端以 prepare／commit 實作）。
4. `original` 不依賴任何新版素材即可啟動。
5. 可在同一存檔、座標與狀態切換三模式；新版畫面不能作為原版 parity 證據。
6. `faithful-hd` 共有 1,157 個 manifest 資產；`reimagined` 共有 1,905 個，包含
   九組地形、82 場景、標題／結局、七人八方向全動作與四類武器圖層。
7. `reimagined` 於 960×540、1280×720、1600×900、1920×1080 都重新合成畫面；
   地圖可見範圍隨畫布增加並封頂 25×15，規則掃描範圍保持原版。
8. F2 設定的 `V` 依 `original → faithful-hd → reimagined → original` 循環；任一
   bundle 驗證或解碼失敗均不提交新模式。

## 7. 現行實作與證據

- loader／驗證：`internal/artpack`。
- 固定 4:3 與全面重構 renderer：`internal/play/art_mode.go`、
  `internal/play/reimagined.go`。
- 動態視窗：`internal/ui.Game.Layout`；僅 `reimagined` 依外部尺寸改變邏輯畫布。
- 可重現生成：`tools/art/build_reimagined.py`；輸出逐檔雜湊在
  `artpacks/reimagined/manifest.json`，生成契約在
  `artwork/manifests/reimagined-build.json`。
- 玩家路徑截圖：`workplace/art-preview/reimagined-live-1280x720.png`、
  `reimagined-live-1600x900.png`、`reimagined-title-1280x720.png`。

## 8. 大地圖合成與可讀性修正（2026-08-21）

推廣片抽樣證實第一版把 64×48 的 3/4 俯視 tile 當成無重疊正交格排列，造成逐列
黑縫、方塊色差與「素材拼貼」感；暗色七人隊伍縮進影片後也幾乎無法辨識。這是
remake 呈現缺陷，不是原版資料或規則未知。

- tile 使用小於素材尺寸的 X／Y 步距，重疊區以 Alpha 漸層合成，不改 tile ID、
  地圖座標、碰撞、事件或可見範圍上限。
- 隊伍 sprite 放大並加低彩度細定位環；定位環不得遮住地貌或暗示可互動格。
- HUD 只擷取原版內容安全區，不縮放原版外框；新版框線與繁中指令至少保留
  12 px 內距，最長指令不得蓋住邊框。
- 驗收必須包含 1280×720 同座標截圖，以及縮進推廣片版面後仍能辨識隊伍的實機片段。

## 9. 忠實高畫質的中文安全矩形與切換穩定性（2026-08-22）

- 設施圖下方標題只有原版 12 格寬；帶英文原名的中文地名必須先保留中文主名，
  再截限於該安全矩形，不能讓文字畫進右側選單。
- 中文名單使用單層上緣時，左右側框必須從緊接的第一筆資料列開始；不可沿用
  英文雙層 banner 的起始列而留下黑色缺口。
- `faithful-hd` 與 `reimagined` 使用作業系統游標，不放大疊加原版 16×16 `CURS`。
- 實體 `Y/N` 鍵必須在輸入法沒有送出 ASCII rune 時仍能回答確認框。
- `reimagined` 響應式 backing image 切回固定畫布前必須同步重建；連續
  `original → faithful-hd → reimagined → original` 不得因像素長度不符 panic。

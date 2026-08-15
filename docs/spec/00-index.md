# 00：規格索引與閘門狀態

這個目錄是 **G2**（規格收攏）的產物：把 `docs/re/` 的逆向結論整理成可實作的規格。
`CLAUDE.md` §0 的規定是「**只有標 READY 的規格可以動手寫引擎**」，
所以這份索引同時是「`internal/` 底下哪些東西可以開始寫」的白名單。

日期：2026-08-15 ｜ 逆向現況見 [`docs/re/00-remake-knowledge-gaps.md`](../re/00-remake-knowledge-gaps.md)

---

## 1. 目前的閘門狀態

| 規格 | 狀態 | 對應的 `internal/` | 擋住它的東西 |
|---|---|---|---|
| [`01-assets-and-formats.md`](01-assets-and-formats.md) | **READY**（已實作） | `internal/assets` ✅ | — |
| [`02-rng-and-dice.md`](02-rng-and-dice.md) | **READY**（已實作） | `internal/game/rng` ✅ | — |
| [`03-screen-and-text.md`](03-screen-and-text.md) | **READY**（已實作） | `internal/textlayout` ✅、`internal/render` ✅、`internal/ui` ✅ | — |
| [`04-movement-and-clock.md`](04-movement-and-clock.md) | **READY**（已實作） | `internal/game` ✅ | — |
| [`05-character-and-save.md`](05-character-and-save.md) | **READY**（已實作） | `internal/assets` ✅、`internal/game` ✅ | — |
| [`06-combat.md`](06-combat.md) | **READY**（已實作） | `internal/game` ✅ | — |
| [`07-world-events.md`](07-world-events.md) | **READY**（已實作） | `internal/assets` ✅、`internal/game` ✅ | — |
| [`09-facilities.md`](09-facilities.md) | **READY**（已實作） | `internal/game` ✅ | — |
| [`10-cjk-layout.md`](10-cjk-layout.md) | **READY**（已實作） | `internal/assets` ✅、`internal/render` ✅ | — |
| [`11-translation-catalogue.md`](11-translation-catalogue.md) | **READY**（已實作） | `internal/lang` ✅ | — |
| [`12-combat-rounds.md`](12-combat-rounds.md) | **READY**（回合骨架） | `internal/game` ✅ | 逃跑與隊形未解 |
| [`13-encounters-and-spawning.md`](13-encounters-and-spawning.md) | **READY**（已實作） | `internal/game` ✅ | 遭遇佇列是誰填的未解 |
| [`14-combat-commands.md`](14-combat-commands.md) | **READY**（已實作） | `internal/game` ✅ | Hire／Weapon／Use／Load 四支處理程式未解 |
| [`15-encounter-scan.md`](15-encounter-scan.md) | **READY**（已實作） | `internal/game` ✅ | 決定接戰值 15／254 的那條路未解 |
| [`16-combat-screen.md`](16-combat-screen.md) | **READY**（已實作） | `internal/play` ✅ | 命中訊息的字串編號未確認 |
| [`17-command-handlers.md`](17-command-handlers.md) | **READY**（已實作） | `internal/game` ✅ | 可雇用對象的算法、Use 那個 byte 的欄位配置未解 |
| [`18-facility-loops.md`](18-facility-loops.md) | **READY**（已實作） | `internal/game` ✅ | 賣價公式、清單框架未解 |
| [`19-paragraph-journal.md`](19-paragraph-journal.md) | **READY**（已實作） | `internal/game` ✅、`internal/play` ✅ | — |
| [`20-mouse-input.md`](20-mouse-input.md) | **READY**（已實作） | `internal/input` ✅ | — |
| [`08-audio.md`](08-audio.md) | **READY**（已實作） | `internal/audio` ✅ | 3／6／8 **沒有呼叫端**（已確認，`docs/re/44` §6）|

**二十一份 READY 規格涵蓋資產層、呈現層、亂數、規則層十一塊、戰鬥畫面、段落手札、中文排版與翻譯管線**——
走一步、遊戲時鐘、體力隨時間恢復、事件分派，以及**從視野掃出遭遇、生怪、下指令、
排行動順序到打完一場戰鬥或逃走**的整條路。
**規格全部 READY，沒有未寫的了。**

## 2. 為什麼這些現在就能寫

閘門是**逐機制**的，不是「整個逆向做完才能開始」。每一份底下的每一條敘述都：

- 在 IDA 裡讀到程式碼，並寫進 `docs/re/NN-*.md`（G1）
- 有可重跑的工具重現過（`tools/` 底下），或有 42/42、9/9 這類全量驗證
- 未解的部分在規格裡**明寫成未解**，不用猜測填空

## 3. 規格的體例

每份規格固定四段：

1. **依據**——引用 `docs/re/NN` 與推論等級，讓實作者知道每條規則的來源強度
2. **規格本體**——資料結構、演算法、邊界條件，寫成可直接翻成 Go 的形式
3. **未解與邊界**——實作時遇到會撞牆的地方，以及暫代方案（標明是暫代）
4. **驗收條件**——**對原版行為驗收**，不是只跑單元測試（`CLAUDE.md` §4）

## 4. 實作順序建議

```
1. internal/assets     ← 規格 01；**已完成**（9 個測試全綠，含 round-trip）
2. internal/textlayout ← 規格 03；**已完成**（控制碼與組行，無相依）
3. internal/render     ← 規格 03；**已完成**（合成索引畫面，幾何逐像素驗過）
4. internal/game/rng   ← 規格 02；**已完成**（驗收數列與分佈全過）
5. internal/input      ← 規格 03；**已完成**（與函式庫無關的按鍵模型）
6. internal/ui         ← 規格 03；**已完成**（Ebiten：上色 ＋ 送圖 ＋ 收鍵）
7. internal/game       ← 規格 04；**已完成**（走一步、時鐘、體力處理、事件分派骨架）
8. 存檔與角色記錄      ← 規格 05；**已完成**（round-trip byte-for-byte、升級與技能公式）
9. 世界事件與檢定      ← 規格 07；**已完成**（section 定址、條件串列、檢定與練等、腳本直譯器）
10. 戰鬥               ← 規格 06；**已完成**（命中、兩種傷害、護甲、傷勢、擊殺經驗值）
11. 設施               ← 規格 09；**已完成**（商店價格、醫生三種收費、訓練師）
12. 中文排版           ← 規格 10；**已完成**（640×400 畫布、倚天 16×15、索引 oracle 過關）
13. 翻譯目錄           ← 規格 11；**管線已通**（抽 4,827 條、編譯擋三種錯、端到端出中文）
14. 戰鬥回合           ← 規格 12；**骨架完成**（三組×10、行動旗標、行動順序表）
15. 音效               ← 規格 08；等 F2 的位元組碼指令集解出來
```

相依取得方式（2026-08-15 定案）：**唯讀掛載本機模組快取當 file proxy**，
`tools/go.sh` 仍然 `--network none`。Ebiten 的視窗層走 cgo，要 X11／GL 標頭，
所以另建 `docker/wasteland-go.Dockerfile`（明確版本、可重現），不是臨時裝套件。

⚠ **Ebiten 在沒有 DISPLAY 的環境會在 package init 就 panic**，所以無頭用得到的
東西一律不放 `internal/ui`：按鍵模型在 `internal/input`、畫面合成在 `internal/render`、
檢視器場景在 `internal/viewer`，三個都測得到。

## 5. 兩支指令

| 指令 | 用途 | 需要 X |
|---|---|---|
| `cmd/wl-shot` | 把一幀寫成 PNG，**與 DOSBox 的原版截圖對拍用**（`-mode play` 也可以） | 否 |
| `cmd/wasteland` | 開視窗；`-mode play` 是遊戲、`-mode map/title/pic` 是資產檢視器 | 是 |

`-mode play`（`internal/play`）**是遊戲**：從出廠存檔開場（挑序號大的那一份、
讀出四個 Ranger、時鐘 01:00、座標 (55, 62)），方向鍵走的是規則層——
四道閘、時鐘推進、體力處理、遭遇擲骰、事件分派。

`-mode map` 仍然是**純檢視器**，一條規則都沒有，留著對拍用。
兩者分成 `internal/play` 與 `internal/viewer` 兩個套件，就是為了讓
「有規則」與「沒規則」永遠分得開。

`internal/assets` 不認識 Ebiten，`internal/game` 不認識畫面（`CLAUDE.md` §4）。
規格 01 與 03 的分界就照這條線切：**解碼回 `image.RGBA` 屬於 01，畫到畫面屬於 03**。

# 00：規格索引與閘門狀態

這個目錄是 **G2**（規格收攏）的產物：把 `docs/re/` 的逆向結論整理成可實作的規格。
`CLAUDE.md` §0 的規定是「**只有標 READY 的規格可以動手寫引擎**」，
所以這份索引同時是「`internal/` 底下哪些東西可以開始寫」的白名單。

日期：2026-08-14 ｜ 逆向現況見 [`docs/re/00-remake-knowledge-gaps.md`](../re/00-remake-knowledge-gaps.md)

---

## 1. 目前的閘門狀態

| 規格 | 狀態 | 對應的 `internal/` | 擋住它的東西 |
|---|---|---|---|
| [`01-assets-and-formats.md`](01-assets-and-formats.md) | **READY**（已實作） | `internal/assets` ✅ | — |
| [`02-rng-and-dice.md`](02-rng-and-dice.md) | **READY**（已實作） | `internal/game/rng` ✅ | — |
| [`03-screen-and-text.md`](03-screen-and-text.md) | **READY**（`textlayout`／`render` 已實作） | `internal/textlayout` ✅、`internal/render` ✅、`internal/ui` | Ebiten 那層還沒做 |
| `04-movement-and-clock.md` | 未寫 | `internal/game` | 事件處理函式 5／8／9 未讀（`docs/re/26` §8） |
| `05-character-and-save.md` | 未寫 | `internal/game` | **C2 存檔欄位未解**、C4 隊伍未解 |
| `06-combat.md` | 未寫 | `internal/game` | D1 技能數值、D5 經驗值與升級未解 |
| `07-world-events.md` | 未寫 | `internal/game` | E2 各處理函式只到入口、E3 段落編號未解 |
| `08-audio.md` | 未寫 | `internal/audio` | F2 位元組碼指令集與曲目資料未解 |

**三份 READY 規格涵蓋的是「讀得懂原版資料、畫得出畫面、擲得出骰子」**，
也就是資產層與呈現層。遊戲規則層一條都還不能寫。

## 2. 為什麼現在就能寫這三份

閘門是**逐機制**的，不是「整個逆向做完才能開始」。這三份底下的每一條敘述都：

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
5. internal/ui         ← 規格 03 的 Ebiten 那層：上色 ＋ 送圖 ＋ 收鍵
6. 其餘                ← 等對應規格 READY
```

`internal/ui` 還沒做的理由是**建置環境**：`tools/go.sh` 刻意 `--network none`，
而 Ebiten 還沒進本專案的模組快取。要做之前先決定相依的取得方式
（vendor 進版控，或用唯讀掛載的本機快取當 file proxy），不要為了方便把建置開網路。

`internal/assets` 不認識 Ebiten，`internal/game` 不認識畫面（`CLAUDE.md` §4）。
規格 01 與 03 的分界就照這條線切：**解碼回 `image.RGBA` 屬於 01，畫到畫面屬於 03**。

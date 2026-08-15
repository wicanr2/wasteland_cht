# 25：設施的選單互動

狀態：**READY** ｜ 日期：2026-08-15 ｜ 對應 `internal/play`

規格 18 是設施的**規則**（庫存、賣、逐點治療、解毒），規格 23 是**場景**
（進得去、看得到、離得開）。這一份補中間的**互動**：按哪個鍵、走到哪一步。

---

## 1. 依據

| 規格內容 | 來源 | 推論等級 |
|---|---|---|
| 商店主迴圈的按鍵與字串 | [`re/42`](../re/42-facility-loops.md) §1 | 已確認 |
| 買／賣的入口條件 | [`re/42`](../re/42-facility-loops.md) §2、§3 | 已確認 |
| 每一列怎麼畫、裝備中要標記 | [`re/42`](../re/42-facility-loops.md) §3.1 | 已確認 |
| 醫生的逐點治療迴圈 | [`re/42`](../re/42-facility-loops.md) §5 | 已確認 |
| 解毒的入口 | [`re/42`](../re/42-facility-loops.md) §6 | 已確認 |

## 2. 商店

```
進場 → 招呼字串（地圖記錄 +0x05）
      隊伍只有一人 → 直接用他
      否則「Who wants to enter?」選人
      選到的人不能行動 → 「… can't buy anything.」，重選
主迴圈：「Do you want to:  Buy / Sell」＋「You have $<餘額>」
    ESC → 離開
    B   → 買：物品陣列滿了就「Your inventory is full.」回主迴圈
    S   → 賣：身上沒有賣得掉的就「You don't have anything they want!」回主迴圈
    P   → 換下一個人
    其他 → 繼續等
```

⚠ **兩個「回主迴圈」不是錯誤分支**，是原版的正常路徑——
按了 B 但背包滿了，畫面回到 Buy／Sell 而不是離開商店。

## 3. 醫生

```
主迴圈：「還差 N points. You can:  Heal 1 point / Continue」＋餘額
    ESC／C → 結束
    H      → 治一點：扣「每點價格」→ CON +1 → **重算差額**
              錢不夠 → 「You don't have enough money.」並跳出
    P      → 換下一個人
```

⚠ **一次按鍵只治一點、只扣一次錢，而且每一輪重算差額**（`re/42` §5）。
差額算在迴圈外會變成「錢不夠就一點都不治」，而原版是治到沒錢就停在那裡、
已治的點數留著。

## 4. Go 介面

```go
// internal/play

// FacilityStep 是設施選單的一個狀態。
type FacilityStep int

const (
    StepMain FacilityStep = iota // 主選單
    StepBuy                      // 買的清單
    StepSell                     // 賣的清單
    StepPick                     // 選人
)

// Key 送一個按鍵給設施選單。回傳 false 表示要離開設施。
func (f *FacilityScene) Key(k byte, p *game.Party, items game.ItemTable) bool
```

## 5. 未解與邊界

- **賣價公式（`sub_1C1C2`）未解**（`re/42` §7）。這一版賣價用買價
  （`ShopPrice`），並在畫面與註解裡標明是暫代——**不要當成解出來的**。
- 清單框架（`sub_16DB4`／`sub_16D34`）的每列回呼介面未解，所以清單用
  重製版自己的排版，不假裝重現原版的框架。
- 技能訓練師的流程未解（`re/42` §7），這一版不接。
- 解毒的疾病編號對應已在規格 09；這一版接主迴圈的入口，選單細節下一輪。

## 6. 驗收條件

1. **背包滿了按 B**：回主迴圈，不離開商店，不當成錯誤。
2. **身上沒有賣得掉的東西按 S**：同上。
3. **賣一件**：槽清成 0、錢增加、店家庫存 +1（`0xFF` 不變）。
4. **治療**：錢只夠治 2 點時，CON 只增加 2，而且**已治的點數留著**。
5. **不能行動的人**（CON ≤ 0）不能被選進商店。
6. ESC 在任何一層都回得去：清單 → 主迴圈 → 離開設施。

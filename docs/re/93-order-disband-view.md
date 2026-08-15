# 93：`ORDER`／`DISBAND`／`VIEW` —— 兩支要多隊伍，一支不用

日期：2026-08-15 ｜ 接 `docs/re/91`（指令列的七個入口）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

指令列剩下的四支裡有三支形狀不長。讀完之後分成兩類：
`ORDER` 只動一支隊伍內部，`DISBAND` 與 `VIEW` 要有**多隊伍**才有意義。

---

## 1. `ORDER`（`0x12AE0`）：重排隊伍槽表

```
0x12AE0  dl ← ds:4653h（人數）；dec dl ＝ 0 → retn    ; **一個人不用排**
loc_12AE9:
  bl ← 7
  迴圈（bl 從 7 數到 0）:
      al ← [ds:46B7h + bl]      ; 隊伍槽表（docs/re/15）
      [0x46CF + bl] ← al        ; 抄到暫存區
      [ds:46B7h + bl] ← 0       ; 原槽清空
  push ds:4653h
  ds:A4C0h ← 0；ds:4653h ← 0    ; 人數暫時歸零
  sub_1728C                     ; 切畫面
  pop → ds:4653h
  al ← 0x0F；sub_16CB2          ; 印訊息 15
  sub_12B8A                     ; ← 逐一重新指定順序
```

**八個槽**（`bl` 從 7 到 0）——隊伍上限是 7 人 ＋ 1 格。
做法是「整份抄到暫存、原表清空、再逐一放回去」，
所以中途取消會留下空表，`sub_12B8A` 必須把每一格都填回去。

排完之後的互動在 `0x12B18`：

```
loc_12B18:
  al ← 1；sub_1721B            ; 挑一個隊員（1–9）
      ＝ 0 → 取消
  al ← [0x46CF + bl]           ; 暫存區那一格
      ＝ 0 → 回去重問           ; **已經放過的選不了**
  [0x46CF + bl] ← 0            ; 從暫存區移除
  ds:A4C0h++                   ; 新順序的位置
  [ds:46B7h + ds:A4C0h] ← al   ; 放進槽表
  sub_17033                    ; 印
  ds:A4C0h + 1 ≠ ds:4653h → 繼續
```

`sub_12B8A`（`0x12B15`）是「把暫存區裡還沒放回去的人列出來」——
逐格 `sub_16EE8` ＋ `sub_171B9`（印名字，`docs/re/15`），不是重排本身。

推論等級：**已確認**（迴圈邊界、暫存區位址、挑人與寫回都讀到）。

## 2. `DISBAND`（`0x15E77`）：分出一支新隊伍，上限四組

```
0x15E77  al ← ds:4653h；＝ 1 → retn      ; **一個人不能分**
loc_15E7F:
  sub_19727
  ds:A6C8h ← 0                            ; 數能行動的人
  ds:A6C6h ← 1
  迴圈到 ds:4656h:
      loc_19624；loc_172BE                ; CON ≤ 0 的不算（docs/re/89 §2）
      能行動 → ds:A6C8h++
  al ← ds:4657h（目前有幾組）
      ＝ 3 → 印訊息 0x15；回地圖          ; **已經有 4 組（0–3）就不能再分**
```

`ds:4657h` 是隊伍組數上限的指標，與 `docs/re/91` 的 `ds:CF86h`（`< 4`）一致——
**原版最多四支隊伍**。

## 3. `VIEW`（`0x160A8`）：切到下一支隊伍

```
0x160A8  al ← ds:4654h（目前是哪一組）
         ds:A6C3h ← al；ds:A6C2h ← al      ; 起點記兩份
loc_160B1:
  bl ← ds:A6C3h
      ＝ ds:4657h → bl ← 0                 ; 環繞
      否則 bl++
  ds:A6C3h ← bl
      ＝ ds:A6C2h（起點）→ loc_160D6        ; **繞回原點 ＝ 沒有別組**
  al ← bl；sub_160E1                        ; 試著切過去
      失敗（CF）→ 回 loc_160B1              ; 換下一組再試
  jmp sub_163C1                             ; 成功 → 重畫
loc_160D6:
  sub_1676A；sub_17DF1
  al ← ds:A6C2h；→ 切回原本那組
```

**只有一支隊伍時它會繞一圈回到原點，然後把畫面切回原本那組**——
也就是「什麼都沒發生」，不是錯誤路徑。

## 4. remake 這一側

| 指令 | 前置 | 狀態 |
|---|---|---|
| `ORDER` | 無（只動一支隊伍內部） | **已接**（`internal/play/order.go`）|
| `DISBAND` | **多隊伍**（最多 4 組） | 缺架構 |
| `VIEW` | **多隊伍** | 單組時照原版「繞一圈回原點」＝ 什麼都不做，這一版就照這條路做 |

remake 目前只有第 0 組（`docs/spec/21` §4）。三支裡兩支卡在同一件事，
所以**多隊伍是一個前置工項**，不是三個獨立的指令各缺一塊。

## 5. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/cmd3.json 0x12AE0 0x15E77 0x160A8
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/cmd4.json 0x12AE9 0x15E7F

tools/go.sh test ./internal/play/ -run TestViewWithSingleParty -v
```

## 6. 這一輪學到的（寫成規則）

- **三個缺口可能是同一個缺口。** `DISBAND` 與 `VIEW` 各自看起來是一支待實作的
  指令，實際上兩支都只是「多隊伍」這個前置的介面。
  **把工項排進清單之前先問「它們的前置是不是同一個」**，
  否則會做出兩個各缺一半的半成品。
- **「什麼都不做」也要照原版的路走。** `VIEW` 在單組時不是特例分支，
  是「繞一圈回到起點」的自然結果。照著做就不必為「只有一組」寫特例，
  將來接上多隊伍時這條路也不用改。

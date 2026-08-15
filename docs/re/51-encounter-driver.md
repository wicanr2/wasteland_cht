# 51：遭遇驅動器——地圖與戰鬥之間那一層

日期：2026-08-15 ｜ 對應盤點 **D2**（戰鬥流程）、**E1**（遭遇）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

---

`docs/re/36`–`40` 解的是「戰鬥裡面怎麼跑」，`docs/re/39` 解的是「遭遇怎麼冒出來」。
中間少一層：**誰決定要打、什麼時候切畫面、打完怎麼回地圖**。那一層是
`sub_11CD0`（678 bytes，249 條指令），本份把它讀完。

## 1. 入口只有一個

```
sub_163C1 (0x163C1)  jmp sub_11CD0
```

三個呼叫端：`sub_12760`、`sub_160A8`、以及**地圖主迴圈的 `0x16BA8`**
（節拍計數器滿 `0x400` 那條路，`docs/re/26` §1.2）。走一步之後的那一次由
`loc_163A0` 那條線進來。

## 2. 開頭：掃描 → 要不要打

```
0x11CD0  ds:46DCh ← ds:4655h      ; 記住「現在這張地圖」
0x11CD6  call sub_14664           ; 遭遇掃描（docs/re/39），CF 回報
0x11CD9  lahf                     ; 把 CF 收起來
0x11CDA  ds:46DFh ≠ 0 → retn      ; 有旗標就不打
0x11CE4  CF ＝ 1 → retn           ; 掃描說「沒有可打的」就不打
```

⚠ **`lahf`／`sahf` 那一對很容易讀漏。** 中間夾了一個旗標檢查，
CF 是先存起來、檢查完再放回去比的。把它當成兩個獨立的 `retn` 會少一個條件。

## 3. 打之前先把每個人的經驗值抄一份

```
0x11CF1  ds:71B7h ← ds:4656h              ; 角色數
0x11CF7  dl ← 1
0x11CFF  迴圈（ds:A435h ＝ 1 … 7）：
           loc_19624 / sub_19B67           ; 選到第 n 個角色
           [71B7h + dl+0] ← ds:466Bh       ; 經驗值 24-bit 的三個 byte
           [71B7h + dl+1] ← ds:466Ch
           [71B7h + dl+2] ← ds:466Dh
           dl += 3
```

**這份快照的唯一用途是戰鬥結束後報「誰拿了多少經驗值」**（§6）。
原版沒有「這一場拿了多少」的累加器，它是用**前後相減**算出來的。

## 4. 主體：四個隊伍組各跑一次

`ds:A435h` 從 0 數到 3（`ds:4657h` 是上限），對每一組：

| 位址 | 做什麼 |
|---|---|
| `0x11D67` | `sub_16149(組)`——視窗原點對到這一組 |
| `0x11D72` | 這一組在不在「打過的地圖」清單（`ds:A9B0h`，4 bytes）裡 |
| `0x11D89` | `sub_19D0E`：這一組有沒有敵人 |
| `0x11D8E` | **`sub_1728C`：畫面模式 ← 1**（地圖視窗換成隊伍名單，`docs/re/40` §1）|
| `0x11D97` | 印字串 **54** ＋ **76**：`This party isn't on this map and isn't in battle.` ＋ `Do you want them to execute a battle round?` |
| `0x11DA1` | `sub_12619`：Yes／No，選 No 就跳過這一組 |
| `0x11DAF` | `sub_14480`：依佇列建敵方記錄（`docs/re/37` §3）|
| `0x11DD8` | **`sub_11F76`：指令階段**（§5）|
| `0x11E27`–`0x11E42` | 結算：`sub_12551`／`sub_1354F`／`sub_13580`／`sub_15036`／`sub_19EFC` |
| `0x11E45` | `sub_1465F` 重掃，決定要不要再來一輪 |

**一輪 ＝ 四組都問完指令，再一起結算。** 不是「一組打完換下一組」——
`ds:A437h` 記「這一輪有沒有人下過指令」，全部沒有就結束（`0x11E08`）。

## 5. `sub_11F76`：一組的指令階段

```
0x11F8C  sub_19D0E ≠ 0 ？        ; 沒有敵人 → clc, retn（這組不打）
0x11FB3  三組敵人的數量都是 0 ？   ; → 印字串 20 ＋ 76（This party is not
                                  ;   being attacked. / 要不要照樣打一輪）
0x11FF8  對每個成員（1 … ds:4653h）：
           sub_172BB     ; 不能行動就跳過（CON ≤ 0）
           sub_171B9     ; 印名字
           字串 55       ; `, choose:` ＋ 七個選項
           ds:4680h ← 0xA44Dh   ; 熱鍵表 " HAWRELU"（docs/re/38 §2）
           sub_173B0     ; 選單，回指令碼
           ds:46D8h[成員] ← 指令碼
           ds:4661h ← ds:A43Bh[指令碼 × 2]   ; 指令處理程式
           sub_19EB4     ; 間接呼叫它（docs/re/41）
           ds:46DAh[成員] ← 處理程式回的參數
0x12085  處理程式回 CF ＝ 1（取消）→ 退回上一個成員重問
```

⚠ **指令與參數是分開存的兩張表**：`ds:46D8h` 存指令碼、`ds:46DAh` 存參數。
處理程式只負責「檢查 ＋ 選參數」，不執行動作（`docs/re/41` 早就讀出這個形狀）——
這裡看到的是它的另一半：**存哪裡**。

指令 4（Run）而且處理程式回負值 → 跳 `loc_11F93`：整組每個人的
`ds:46D8h` 設成 **8**、`ds:46DAh` 設成回傳值的低 7 位——**全隊一起逃**。

## 6. 收尾：經驗值前後相減

```
0x11ED7  對每個角色（1 … ds:71B7h）：
           loc_19624 / sub_19B67                ; 選到第 n 個
           ds:466Bh −= [71B7h + n+0]  （24-bit 借位相減）
           ds:466Ch -= …（sbb）
           ds:466Dh -= …（sbb）
           差值 ≠ 0 → 印字串 39 ＋ 數字 ＋ 字串 40
                      （`\x0b gains ` ／ ` experience.`）
0x11F33  sub_1272E
0x11F36  sub_19E53(6)                            ; 訊息視窗收尾
0x11F3B  sub_125B7
0x11F41  sub_18350                                ; 重新載入地圖
0x11F44  jmp sub_163C4                            ; 回地圖重畫
```

⚠ **`ds:466Bh`–`466Dh` 是經驗值不是金錢。** 兩個都是 24-bit、都在角色記錄裡，
判準是這裡印的字串 39／40 明寫 `gains … experience.`（`docs/re/40` §3 那張表
是獨立來源）。

## 7. 中文化要注意的兩處

- 字串 **39** 開頭是 `\x0b`（名字佔位），**40** 結尾是兩個 `\r`。
  中文譯文要保留同樣數量的控制碼，否則名字會印不出來或版面錯位（`docs/re/28`）。
- 字串 **55** 的七個選項各自帶 `\x10`，熱鍵比對走靜態字母表不跟著翻譯走
  （`docs/re/40` §4）——中文寫成 `A 攻擊` 這種形式。

## 8. 還沒讀的

- `sub_12619`（Yes／No 提示）的按鍵比對細節。
- 結算那一串六支裡的五支（`sub_12551`／`sub_1354F`／`sub_13580`／
  `sub_19EFC`／`sub_1272E`）各自做什麼——`docs/re/36` 解過回合結構，
  但沒把它們對到回合的哪一段。**`sub_15036` 已解**：
  敵人在地圖上移動（`docs/re/87`）。
- `ds:46DFh`（開頭那個「有旗標就不打」）是誰設的。

## 9. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/combatentry.json 0x11CD0 --callers

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/battleloop.json 0x11F76

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_call_context.py \
  workplace/analysis/dumps/scan-callers.json 0x14664 0x14480 --before 14
```

## 10. 這一輪學到的（寫成規則）

- **「哪一層做決定」要單獨找。** 規則層（怎麼打）與資料層（誰冒出來）都解完了，
  中間那層驅動器還是缺的——它不在任何一邊的筆記裡，因為它兩邊都不是。
  重製版接不起來的時候，缺的常常是這種**中間層**。
- **`lahf`／`sahf` 夾著別的檢查時，條件不只一個。** 把它讀成兩個獨立出口
  會漏掉一個條件，而漏掉的那個只在特定狀態下才看得出來。
- **原版沒有的欄位，它用前後相減算。** 「這一場拿了多少經驗值」沒有累加器，
  是打之前抄一份、打完相減。重製版照做就不必多一個狀態。

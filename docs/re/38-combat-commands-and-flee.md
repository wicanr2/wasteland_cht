# 38：戰鬥的指令階段與逃跑

日期：2026-08-15 ｜ 對應盤點 **D2**（戰鬥流程）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

---

## 1. `sub_11F76`：一回合開始時逐人下令

```
sub_13924(ds:4654h)                     ; 準備這一組隊伍
ds:46D8h[1..隊伍人數] ← 0                ; 每個人的指令歸零
sub_19D0E                                ; 還有沒有敵人？沒有就 clc; retn
若三個敵方記錄槽都沒有活的遭遇：
    印字串 20 ＋ 76                       ; 「This party is not being attacked.
                                         ;   Do you want them to execute a battle round?」
逐一問每個成員（ds:0A432h 從 1 數到 ds:4653h）：
    sub_172BB(成員)  → CF 設就跳過        ; 倒下的人不下令（§4）
    印字串 55                             ; 「, choose: Run Use Hire Evade Attack Weapon Load/unjam」
    ds:4680h ← 0A44Dh                     ; 熱鍵字母表
    sub_173B0                             ; 等按鍵 → 字母表索引（§2）
    ds:46D8h[成員] ← 指令碼
    ds:4661h ← ds:A43Bh[指令碼 × 2]       ; 該指令的處理程式位址
    sub_19EB4                             ; ＝ jmp word ptr ds:4661h
    CF 設 → 這個人重問
    ds:46DAh[成員] ← 處理程式的回傳值      ; 指令的參數
```

所以：

| 全域 | 內容 |
|---|---|
| `ds:46D8h + 成員` | **這回合的指令碼** |
| `ds:46DAh + 成員` | **指令的參數**（逃跑方向、攻擊目標…） |

這解掉 `docs/re/20` §6 的「`ds:46D8h` 那個決定基礎值 40／50／60 的欄位是什麼」——
是**這個人這回合選了什麼**（§3）。

推論等級：**已確認**（整段逐指令讀完，跳表與字母表都倒出來對過）。

## 2. 指令碼 ＝ 熱鍵字母表的索引

`sub_173B0` 只做一件事：等一個按鍵，然後拿 `ds:4667h`（大寫化的按鍵碼）
線性掃 `ds:4680h` 指到的字母表，回傳**索引**；掃不到或 ESC 回 `0xFF`。

`ds:A44Dh` 那張表是 `20 48 41 57 52 45 4c 55`：

| 指令碼 | 字母 | 意思 |
|---|---|---|
| 0 | `' '` | 沒下令 |
| 1 | `H` | Hire（雇用） |
| 2 | `A` | Attack（攻擊） |
| 3 | `W` | Weapon（換武器） |
| 4 | `R` | Run（逃跑） |
| 5 | `E` | Evade（迴避） |
| 6 | `L` | Load/unjam（裝填／排除卡彈） |
| 7 | `U` | Use（使用） |

⚠ **選單顯示的順序不是指令碼。** 字串 55 印的是
`Run / Use / Hire / Evade / Attack / Weapon / Load-unjam`，
和字母表的順序完全不同——照選單順序編號會把每一條規則都對錯人。
指令碼一律以字母表為準。

`ds:A43Bh` 是 10 筆的處理程式跳表，索引就是指令碼：

| 指令碼 | 處理程式 | 備註 |
|---|---|---|
| 0 | `0x1208D` | 沒下令：問方向、切回地圖、`stc`（＝ 這個人重問） |
| 1 Hire | `0x1236F` | |
| 2 Attack | `0x120C1` | 取裝備武器（`sub_196C9`），沒武器就印字串 56 並 `stc` |
| 3 Weapon | `0x1234E` | |
| 4 Run | `0x123A6` | §3 |
| 5 Evade | `0x12078` | **`clc; retn`——什麼都不做** |
| 6 Load | `0x123E2` | 裝填，成功印 `A`／`B` 兩種訊息 |
| 7 Use | `0x1240B` | |
| 8 | `0x120B2` | 不是選單選得到的：把 `ds:0A457h` 清 0 再當成指令 2 走 |
| 9 | `0x14820` | 未追 |

**Evade 的處理程式是空的**——迴避完全靠指令碼本身：敵方算命中門檻時
看到 `ds:46D8h` ＝ 5 就用基礎值 60（`docs/re/20` §1.1），
攻擊（2）用 50、其餘用 40。**選了迴避就更難被打中**，這樣就說得通了。

## 3. 逃跑（`0x123A6`）

```
隊伍只有一個人 → 直接跑（sub_123D1）
否則印字串 36：「Who should run?  Party / Single player」
    等按鍵：
        ESC  → stc; retn          ; 取消，這個人重問
        'S'  → sub_123D1          ; 只有自己跑
        其他 → sub_123D1，回傳值 |= 0x80   ; 整隊跑
```

`sub_123D1` 就是逃跑本身：

```
sub_163C4     ; 重建地圖／區段狀態，並用目前資源重跑一次遭遇同步（sub_14664）
sub_125BD     ; 「Which way?」——方向選單，回傳 0–4
sub_1728C     ; ds:46B9h ← 1，畫面切回地圖視窗
sub_190B8
```

方向選單的字母表是 `ds:A45Ch`：`49 4B 4A 4C 20 C8 D0 CB CD 00`
＝ `I K J L ' '` ＋ 上／下／左／右四個方向鍵的掃描碼，
所以 `sub_125BD` 收到索引 > 4 就減 5——**兩組按鍵對到同五個方向**。
提示字串在 `ds:A469h`（`Which way?`）。

⚠ **逃跑沒有擲骰，也沒有失敗分支。** 選好方向就走，
成功率不是規則的一部分。

整隊逃跑會回到 `loc_11F93`：

```
al &= 0x7F                       ; 去掉「整隊」旗標，剩下方向
ds:46DAh[每個成員] ← 方向
ds:46D8h[每個成員] ← 8           ; 指令 8
clc; retn                        ; 指令階段結束
```

## 4. 誰不能下令：`sub_172BB`

```
sub_19614(成員)            ; ds:46B5h ← 該成員的角色記錄
記錄 +0x1E 為負            → stc（不能下令）
記錄 +0x1D/+0x1E ＝ 0      → stc
否則                        clc
```

也就是 **CON ≤ 0 的人不下令**——與 `docs/re/15` 的 CON 欄位一致。

## 5. 還沒解的

- **隊形**。目前只知道射程相關的是 `ds:46CCh`／`ds:46CDh`（`sub_15755`），
  還沒讀到「站位」這種東西存在的證據——不要先假設有。
- 指令 1（Hire）、3（Weapon）、7（Use）三支處理程式的內容。
- 跳表索引 9（`0x14820`）是誰用的。
- `sub_190B8`。

## 6. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/menu.json \
  0x173B0 0x19EB4 0x172BB 0x16CB2 --callers

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/flee2.json \
  0x125BD 0x163C4 0x1728C 0x190B8 0x125B7 0x12636 --callers

python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xA43B 10 --code
python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xA44D 16 --bytes
```

`0x11F76`–`0x1242x` 大半沒被 IDA 建成函式，要用 `export_listing.py` 的
逐指令 JSON 讀（`workplace/analysis/dumps/listing.json`）。

## 7. 這一輪學到的（寫成規則）

- **選單的顯示順序與內部編號可以完全不同。** 這裡選單印
  `Run Use Hire Evade Attack…`，指令碼卻是熱鍵字母表 `' ' H A W R E L U`
  的索引。看到「選單」就照顯示順序編號，會把每一條規則都對錯人——
  要找的是**按鍵怎麼變成數字**那一支（這裡是 `sub_173B0`）。
- **處理程式是 `clc; retn` 不代表那個指令沒作用。** Evade 的處理程式是空的，
  作用全在別處讀 `ds:46D8h` 的那一行。空函式是「效果在別處」的線索，不是死碼。
- **一張跳表就把整個階段的骨架給出來了。** 與其逐一追呼叫端，
  不如先把 `ds:A43Bh` 那 10 個位址倒出來——每一支的開頭就說明了它是哪個指令。

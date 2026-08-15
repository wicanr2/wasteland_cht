# 94：`ENC` —— 它不是新指令，是自動遭遇的手動入口

日期：2026-08-15 ｜ 接 `docs/re/91`（指令列的七個入口）、`docs/re/51`（遭遇驅動器）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

指令表 `ds:AB1Ch` 的第 1 項寫的是 `0x11CE7`，`docs/re/91` 因此把它記成
「七項裡最長的一支，還沒讀」。**其實它早就解完了**——
`0x11CE7` 不是函式起點，是 `sub_11CD0` 的第二個入口，
而 `sub_11CD0` 就是 `docs/re/51` 整份在講的那支遭遇驅動器。

---

## 1. 兩個入口，一支程式碼

```
sub_11CD0:                       ; ← 走一步之後的自動入口
  ds:46DCh ← ds:4655h            ; 記下目前地圖
  sub_14664                      ; 掃描：要不要打
  lahf                           ; **把掃描的 CF 存起來**
  ds:46DFh ≠ 0 → retn            ; 另一道旗標（誰設的未解）
  al ← 1; sahf                   ; 還原 CF
  jnb loc_11CE7                  ; 不打 → retn
locret_11CE6:
  retn

loc_11CE7:                       ; ← **指令表指到這裡**
  ds:A436h ← 0                   ; 「這一輪真的打過」旗標
  ds:CA64h ← 0
  ds:71B7h ← ds:4656h            ; 人數，等一下當經驗值快照的迴圈上界
  …
```

按 `E` 就是**跳過掃描直接進驅動器**。所以 `ENC` 沒有自己的規則，
它與自動遭遇共用同一條路——`docs/re/51` §3–§6 描述的就是它。

`lahf`／`sahf` 夾著 `ds:46DFh` 的檢查，是兩個獨立條件（`docs/re/51` §10 已記）。

推論等級：**已確認**（兩個入口的位址與指令流都讀到；
指令表的值是從 `ds:AB1Ch` 直接倒出來的）。

## 2. 不在這張地圖的隊伍要先問一句

外層對四支隊伍各跑一次（`loc_11D56` … `loc_11F73`）。每一組進去之前：

```
0x11D6A  ds:46DCh ← ds:4655h        ; 這一組在哪張地圖
0x11D70  bl ← 3
loc_11D72: [ds:A9B0h + bl] ＝ ds:4655h → loc_11DA9   ; 已在交戰清單 → 直接打
           bl−−，jns loc_11D72
0x11D80  ds:46DCh ＝ ds:46E0h → loc_11DA9             ; 是主地圖 → 直接打
0x11D89  sub_19D0E；ZF → 跳過這一組
0x11D97  al ← 36h；sub_16CB2        ; "This party isn't on this map and isn't in battle. "
0x11D9C  al ← 4Ch；sub_16CB2        ; "Do you want them to execute a battle round?"
0x11DA1  sub_12619                  ; 等 Y／N
0x11DA4  jnb loc_11DA9              ; Y → 打
0x11DA6  jmp loc_11E8A              ; N → 跳過，繼續掃下一組
```

**`ds:A9B0h` 是四格的交戰地圖清單**，`0xFF` ＝ 空格：

| 位置 | 動作 |
|---|---|
| `0x11D44` | `[A9B0 + 0] ← ds:4655h`，其餘由後面填 |
| `0x11E4A`（打完，NC） | 找到 `ds:46DCh` → 什麼都不做；找不到 → 填進第一個 `0xFF` 格 |
| `0x11E72`（打完，CF） | 找到 `ds:46DCh` → 設回 `0xFF`（**這張地圖的戰鬥結束了**）|
| `0x11EB5` | 全部都是 `0xFF` → 收尾；還有非 `0xFF` → 繼續下一組 |

也就是說**外層迴圈的終止條件是「所有地圖上的戰鬥都結束」**，不是「四組都跑過一次」。
`ds:720Eh` 的四格是同一組資料的另一面：`0x11E8A` 那段把「與目前地圖相同」的組標成 1，
`0x11D5A` 讀它決定要不要跳過。

推論等級：**已確認**（三處讀寫與終止條件都讀到）／
**未解**（`sub_19D0E` 判的是什麼、`ds:46DFh` 誰設的）。

## 3. 收尾：經驗值前後相減

`0x11ED7` 起逐人做 24-bit 借位相減（`sub` ＋ 兩個 `sbb`），
差值不為零就印字串 `0x27` ＋ 數字 ＋ `0x28`：

```
exe[1][0x27] = "\v gains "
exe[1][0x28] = " experience.\r\r"
```

**這兩句是把 `ds:466Bh`–`466Dh` 判成經驗值而不是金錢的直接證據**——
它們就印在那個減法的後面。`docs/re/51` §6 已經寫過，這裡是字串本身的佐證。

迴圈上界是 `ds:71B7h`，也就是開頭抄進去的人數：
快照表的版面是 `[0]` ＝ 人數、之後每人 3 bytes。

## 4. remake 這一側

**已接**：`Scene.cmdEnc`（`internal/play/enc.go`）→ `StartEncounter`，
與自動遭遇同一支，照原版「兩個入口一支程式碼」。
不在同一張地圖的隊伍走 Y／N 問句（`updateEncAsk`），答 `N` 就跳過。

**沒做的一件事**：原版的外層迴圈會**在四支隊伍之間輪流跑回合**，
直到 `ds:A9B0h` 全空。重製版的戰鬥是單一場景，一次只驅動一支隊伍——
答 `Y` 會切過去讓那一組打，但不會自動輪回來。
`ds:A9B0h` 在重製版沒有對應物，所以「不在同一張地圖」就是唯一的問句條件。

## 5. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/enc.json 0x11CE7

tools/go.sh test ./internal/play/ -run TestEnc -v
```

字串的取法（`ExeStrings()` 的第 1 張表）：

```go
tables, _ := rom.ExeStrings()
tables[1][0x27] // "\v gains "
tables[1][0x36] // "\aThis party isn't on this map and isn't in battle. "
```

## 6. 這一輪學到的（寫成規則）

- **待辦清單上的「未解」要對照函式索引再相信一次。** `WORKLIST` 把 `Enc`
  列成「七項裡最長的一支」，`docs/re/51` 卻已經把同一支函式整份解完了——
  差別只在一個記的是**指令表的入口位址**（`0x11CE7`），
  一個記的是**函式起點**（`0x11CD0`）。
  **中途入口會讓同一支函式在筆記裡長出兩個身分。**
- **字串是型別判定的一手證據。** `ds:466Bh` 是經驗值還是金錢，
  靠欄位寬度與位置永遠只能推；`" gains "` ／ `" experience."`
  就印在那個減法後面，一句話定案。
  **看到「印字串 N」就把 N 倒出來**，別留到之後再查。
- **迴圈的終止條件不一定是迴圈變數。** 這裡外層看起來是「跑過四組」，
  實際是「`ds:A9B0h` 全空」。照迴圈變數實作會在多組交戰時提早結束，
  而且症狀只是「有一組沒打完」。

# 69：條件閘的四個旗標（記錄 `+0x00` 的低位）

日期：2026-08-15 ｜ 接 `docs/re/67`（獎懲參數）、`docs/re/68`（改寫地圖格）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 見 `docs/re/01`。

`sub_13EC9` 的收尾段有四條 `call sub_142E2` ＋ `and al, n` 的分支。
`sub_142E2` 讀的就是**地圖記錄的 `+0x00`**，所以那四條是記錄第一個 byte
低位的四個旗標。這一份把四條都讀完並接上。

---

## 1. 四個旗標

| bit | 遮罩 | 檢查點 | 語意 | 出貨資料 |
|---:|---:|---|---|---:|
| 2 | `0x04` | `0x1401D`（收尾） | 只要**有人通過**（`ds:A5D0h` ≠ 0）就算過 | 76 |
| 3 | `0x08` | `0x14052`（有人通過時） | **立刻收尾**，改寫位移由 `sub_142B1` 依「通過的是哪一條條件」算 | 61 |
| 4 | `0x10` | `0x13FB5`（有人失敗時） | **立刻收尾**，懲罰改成 `sub_14296`（**全隊每個人各套一次**） | 15 |
| 5 | `0x20` | `0x13FE6`／`0x14004`／`0x14063` | **逐個角色跑**；沒設就只處理第一個能行動的人 | 139 |

格數是 42 張地圖裡 nibble 2 指到的 **424 筆相異記錄**的統計
（`TestGateFlagsAppearInShippedData`）。**四條都有資料會走到，沒有死碼。**

## 2. `sub_13EC9` 的控制流

```
0x13ED4  A5D0h（通過人數）← 0；A5D4h（失敗人數）← 0
         A5D2h（受罰人數）← 0；A5D1h（回傳值）← 0
0x13EE4  sub_16D1A(bl ＝ 1)          ; 印記錄 +0x01
0x13EE7  記錄 +0x00 shl 1；jns → 回傳 1、CF ＝ 0（bit6 沒設 ＝ 放行）

每個角色（A5D3h ＝ 1..ds:4653h）：
  sub_172BB CF ＝ 1 → 換下一個（不能行動的不算沒過）
  逐條跑 +0x0A 起的條件串列
    通過 → 0x1404B
    0xFF → 0x13FB1（這個人沒過）

0x1404B（通過）  inc A5D0h
  & 8  → sub_14175（印 +0x02）；sub_142B1(A5CFh) → 改寫位移；收尾
  & 20 → 換下一個角色
  否則 → 0x1406A：sub_14175；改寫位移 ＝ 4；收尾（放行）

0x13FB1（沒過）  inc A5D4h
  & 10 → 0x14010：inc A5D2h；sub_142ED；sub_14296（全隊受罰）→ 落到 0x1401A
  +0x08 ＝ 0 → 不罰（連 A5D2h 都不加）→ & 20 決定換人或收尾
  否則 → sub_14193(A5D3h)（罰這個人）；inc A5D2h
         & 20 → 換下一個角色；沒設 → 0x14028

迴圈跑完（0x13FFB）  A5D4h ＝ 0 → 0x1406A（放行）；≠ 0 → 0x1401A

0x1401A（收尾）  & 4 且 A5D0h ≠ 0 → 0x1406A（放行）
0x14028          A5D2h ＝ 0 → sub_1417A（印 +0x03）
                 否則欄位 ＝ 0x1D → sub_19E53(6)（重畫）
0x14041          inc A5D1h（＝ 擋住）；改寫位移 ＝ 6
0x1406F          sub_17CFF(al, A5D6h, A5D5h)；回傳 A5D1h
```

回傳值 `ds:A5D1h` 只在 `0x14041` 加一次，所以**擋住的唯一條件是走到 `0x14028`**。

`0x14045` 的 `mov al, 6 / test al, al / jnz` 恆成立，是編譯器產物。

## 3. 兩支印訊息的 helper

```
sub_14175:  bl ← 2；jmp sub_16D1A     ; 印記錄 +0x02
sub_1417A:  bl ← 3；jmp sub_16D1A     ; 印記錄 +0x03
```

所以一筆條件記錄的前四個 byte 各是一句話：

| 位移 | 什麼時候印 |
|---:|---|
| `+0x00` | 被第四道閘擋住時（`docs/re/62`），同時是這四個旗標的載體 |
| `+0x01` | 進 `sub_13EC9` 就印（`docs/re/66`） |
| `+0x02` | 通過 |
| `+0x03` | 沒過**且一個人都沒受罰** |

有人受罰時不印 `+0x03`——受罰本身已經有 `" gets hurt for "` 那句。

## 4. `sub_142B1`：條件串列後面接一張改寫表

```
0x142B1  A5C0h ← bl                  ; 通過那條條件的位移
0x142B5  掃到 0xFF（每次 +2）
0x142C5  al ← A5C0h − 0Ah            ; ＝ 2n（通過的是第 n 條）
0x142CA  inc bl                      ; 0xFF 的位置 + 1
0x142D0  回傳 bl ＝ (0xFF 位置 + 1) + 2n
```

**條件串列的 `0xFF` 之後接著一張逐條件的改寫表**，每條條件對應一對 byte
（`sub_17CFF` 要讀 `[位移]` 與 `[位移+1]`）。通過哪一條，這一格就變成哪一種東西。

### 4.1 資料驗證

判準是**取到的值合不合法**：改寫對的第一個 byte 只能是 bit7 設（不改／沿用）
或 nibble ≤ `0x0F`；`0x10`–`0x7F` 是不可能的值，佔隨機資料的 44%。

| 取樣 | 對數 | 不合法 |
|---|---:|---:|
| 有 bit3，位移照公式 | 599 | **0（0%）** |
| 沒有 bit3，同樣位移 | 983 | 240（24%） |
| 有 bit3，位移 −1 | 599 | 49（8%） |
| 有 bit3，位移 +1 | 599 | 60（10%） |

`TestCondPatchTableIsRealData`。公式在出貨資料上完全自洽，錯開一格就出現不可能的值。

⚠ **「記錄長度放得下」不能當判準。** `SectionRecord` 回傳的切片一路延伸到
區段結尾，任何位移都放得下——用那個判準測，424 筆全過，什麼都證明不了。

### 4.2 實跑：一格三段的伏擊

地圖 4 的 (1, 2)，nibble 2 記錄 0：

```
c8 02 00 03 | ff ff | 00 00 | 00 00 | 60 01 60 02 60 03 60 04 | ff | 0c 00 0c 00 02 01 02 01
+0x00 ＝ 0xC8：bit7 移動閘放行、bit6 跑條件、bit3 改寫位移依條件
+0x01 ＝ 2   ：You see three outlaws with their backs turned catching a few puffs.
條件         ：型別 3（比隊伍人數）＝ 1／2／3／4，**必定通過其中一條**
改寫表       ：1–2 人 → nibble 12 記錄 0；3–4 人 → nibble 2 記錄 1
```

兩兩成對，而且對到的東西在故事上接得起來：

| 目標 | 內容 |
|---|---|
| nibble 12 記錄 0 | `+0x00` ＝ `They hear you stumble through the door and attack.` |
| nibble 2 記錄 1 | `+0x02` ＝ `While they aren't looking you ambush them and knock them out.`，條件是四條屬性／技能檢定 |

**人少（1–2）直接被發現開打，人多（3–4）還有一次伏擊的機會。**

出廠存檔 4 人走進去：

```
tools/go.sh run ./cmd/wl-play -script "map=4:2:2,left,right,left,right,left" -trace
   2 left   (1, 2)  You see three outlaws with their backs turned catching a few puffs.
   4 left   (1, 2)  While they aren't looking you ambush them and knock them out.
   6 left   (1, 2)
```

三段：看到 → 伏擊成功（記錄 1 通過，用位移 4 改寫成 `00 00` ＝ 空地）→ 什麼都沒有。
**位移公式錯一格，第二次踏上去就不會印那句話。**
`TestCondPatchRewritesCellByPartySize` 驗確定性的那一段（改寫後是不是 nibble 2 記錄 1）。

## 5. `sub_14296` 與 `sub_142ED`

```
sub_14296:  bl ＝ 1..ds:4653h：sub_14193(bl)     ; **全隊每個人各套一次懲罰**
sub_142ED:  sub_19720
            sub_14314(ds:A5C5h)                 ; 暫存 465Bh／46F7h，換成 A5C5h
            sub_178A0(記錄 +0x03)               ; 在那個時間值底下顯示
            sub_18DB4(dl ＝ 4)
            465Bh ← A5E6h；46F7h ← A5E7h        ; **還原**
```

`sub_142ED` 是「在一個暫時換掉的時鐘值底下顯示記錄 `+0x03`」，
不是永久改時間——`sub_14314` 先存後還。

## 6. remake 這一側

`internal/game/gates.go` 的 `EvalGate` 照 §2 的控制流重寫，四個旗標各自具名：

```go
GateAnyPass    = 0x04  // 有人過就算過
GateCondPatch  = 0x08  // 有人過就收尾，改寫位移依條件算
GateWholeParty = 0x10  // 有人沒過就收尾，全隊各罰一次
GateEachMember = 0x20  // 逐個角色跑
```

一併接上的還有：

- **收尾訊息**（`GateResult.Message`）：通過印 `+0x02`、沒過且沒人受罰印 `+0x03`。
- **型別 4（比金錢）失敗直接判這個人沒過**：`0x13F50` 是 `jb loc_13FB1`，
  不是 `jb loc_13FA4`——它不往下一條條件試。其餘型別失敗才試下一條。

還沒接：`sub_142ED` 的「暫時換掉時鐘值」那層顯示效果（remake 目前只印訊息）、
`0x13FC8`–`0x13FD9` 那段「第一個受罰的人才跑一次」的欄位前置處理。

## 7. `0xFE`／`0xFD`：沿用上一次的值

`sub_17CFF` 讀到記錄裡的改寫值是 `0xFE` 或 `0xFD` 時走特例：

```
0x17D11  ＝ 0FEh → sub_17D34
0x17D15  ＝ 0FDh → loc_17D42

sub_17D34:  46B3h ← ds:46FCh；46B4h ← ds:46FDh；jmp loc_17D21
loc_17D42:  call sub_17D34；clc；retn
```

`loc_17D21` 是主流程裡第一個 `call sub_17D47`，所以 **`0xFE` 是「拿上一次的值，
然後照常寫下去」**。`0xFD` 用 `call` 進去，主流程走到 `jmp sub_17D50` 之後那個
`retn` 就返回到 `0x17D43` 的 `clc; retn`——**一樣會寫，只是把回傳的 CF 壓成 0**
（回報「沒改」）。

暫存本身在主流程裡填：

```
0x17D24  call sub_17CD2         ; al ← 這一格目前的 nibble、dl ← 目前的記錄
0x17D27  ds:46FCh ← al
0x17D2A  ds:46FDh ← dl
```

**存的是「這一格被改寫**之前**是什麼」**，而且在寫入之前就取好。
全檔只有四處碰這兩個 byte：`sub_17CFF` 寫、`sub_17D34` 讀，
另外 `0x12BDB`／`0x12C50` 是一對 `push`／`pop` 的先存後還。
所以它是**全域、跨格、跨地圖都不清空**的單一暫存；
它住在資料段不在存檔裡，重新載入就回到零值。

`internal/game/world.go` 的 `carryTerrain`／`carryRecord` 照這個形狀實作。

### 7.1 出貨資料裡用得很多

逐 section 記錄掃過真正會被當成改寫位移的組合，共 **153 處**，
最大的用戶是設施記錄的 `+0x01`／`+0x02` ＝ `fd fd`（跑完把這一格改回原樣）。
分布與掃描方法見 `docs/re/70` §3。

⚠ **逐格掃會得到零。** 沒有格子指到的記錄一樣會被執行——
改寫可以把任何一格變成任何一筆記錄。統計要逐 section 走。

## 8. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/gate-flags.json 0x142B1 0x142D7 0x142ED 0x14296 0x14314

tools/go.sh test ./internal/play/ -run TestGateFlagsAppearInShippedData -v
tools/go.sh test ./internal/play/ -run TestCondPatchTableIsRealData -v
tools/go.sh test ./internal/play/ -run TestCondPatchRewritesCellByPartySize -v
tools/go.sh run ./cmd/wl-play -script "map=4:2:2,left,right,left,right,left" -trace

# 0xFE／0xFD 的四個引用
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_range_refs.py \
  workplace/analysis/dumps/reuse-slot.json 0x46FC 0x46FD
```

## 9. 這一輪學到的（寫成規則）

- **恆真的判準等於沒有判準。** 「改寫表放得下記錄嗎」測出來 424 筆全過，
  看起來是強證據，其實 `SectionRecord` 的切片延伸到區段結尾，
  任何位移都會過。**下結論前先問：這個判準有沒有可能失敗？**
  換成「取到的值合不合法」＋位移 ±1 的負對照，鑑別力才出現。
- **旗標的語意要從它跳到哪裡讀，不是從它旁邊的呼叫讀。** `& 0x10` 的分支
  裡有 `sub_14296`，看起來像「換一種懲罰」；但那條路 `jnz` 之後
  **落進迴圈的收尾段**，所以真正的語意是「有人沒過就整個結束」。
  只看分支裡呼叫了什麼會把控制流讀反。
- **假陽性和假零一樣會誤導，而且更難察覺。** 掃 `0xFE`／`0xFD` 時把
  「條件閘用的位移」套在每一種 nibble 上，撈出 14 筆看起來很像的東西；
  限定成「這個 nibble 真的會用這個位移」之後是零。
  **零命中要問過濾器有沒有洞，有命中要問位置對不對。**
- **一筆記錄的前四個 byte 是四句話。** `+0x00` 擋住時印、`+0x01` 進來就印、
  `+0x02` 通過印、`+0x03` 沒過印。找到其中一個之後，
  **順著 `sub_16D1A(bl)` 的 bl 值把整組找齊**比逐個追呼叫端快。

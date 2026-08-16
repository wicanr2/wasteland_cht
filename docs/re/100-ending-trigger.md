# 100：結局的觸發點 —— 不在資料裡，在主迴圈裡

日期：2026-08-16 ｜ 接 `docs/re/96`（結局的播放序列）、`docs/re/98` §5（範圍縮到一句話）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 的 SHA-256 見 `docs/re/01`。

`docs/re/96` 把結局那一段（`0x1B4F0`）定位在設施跳表 `ds:A4E0h` 的第 4 格，
而 42 張地圖裡**沒有一筆記錄指到第 4 格**。那不是漏掃——那是設計。
第 4 格是程式自己合成出來的。

---

## 1. 跳表有兩個消費點，都不會產生索引 4

全檔掃 `0xA4D0`–`0xA560` 的定址存取，指到那張表的只有兩條指令：

| 位址 | 指令 | 索引從哪來 |
|---|---|---|
| `0x12C8C` | `mov cx, [bx+0A4E0h]` | 地圖記錄 `+0x00 & 0x7F`（bit7 設起來的那一種） |
| `0x12CA6` | `mov cx, [bx+0A4EAh]` | section `0x10` 取出的 opcode ＋ 5（bit7 沒設的那一種） |

資料側兩條都掃過：

- bit7 設的 32 筆記錄用到的索引是 `0 1 2 3 17 23 29 40 42 48 92`——**第 4 格零筆**
  （`TestEndingHasNoTriggerInData`）。
- bit7 沒設的那些，取出的 opcode 是 0–43（另有四筆超出範圍的值），
  `opcode + 5 == 4` 要 opcode ＝ `0xFFFF`，一筆都沒有。

所以**踩地圖踩不到結局**。

## 2. 第三個入口：`sub_1CB30`

```
sub_1CB30（0x1CB30–0x1CB66，唯一呼叫端 0x16C28）：
  al ← ds:722Ah
  al ＝ 0 → retn                          ; 自毀沒啟動就什麼都不做

  ; 24-bit 減法：目前時鐘 − 啟動當下的時鐘
  dl ← ds:4650h − ds:722Bh
  bl ← ds:4651h − ds:722Ch  (sbb)
  al ← ds:4652h − ds:722Dh  (sbb)

  cmp dl, 0F0h ；al ← bl；sbb al, 0；sbb al, 0；cmc
  jnb → retn                              ; 差 ＜ 0xF0 → 還沒到

  call sub_183B1
  al ← 84h                                ; ← **bit7 設、索引 4**
  jmp sub_12C80                           ; ← 合成一次設施分派 ＝ 結局
```

`0x16C28` 在**地圖主迴圈**裡，位置就在全隊陣亡那道檢查（`docs/re/99`）前面：

```
0x16C28  call sub_1CB30        ; ← 倒數
0x16C2B  al ← ds:4656h ；全隊陣亡的檢查
0x16C6C  jmp loc_16B49         ; 迴圈
```

**一輪一次**，所以玩家每做一個動作就檢查一次。

推論等級：**已確認**（兩支程式碼逐條讀完，`0xA4E0` 的引用點做過全檔掃描）。

## 3. 誰啟動倒數：腳本 opcode 35

寫 `ds:722Ah` 的地方只有兩處，一處是結局自己的收尾（`0x1B6B0` 寫 0），
另一處是 `0x1AB0E` ＝ 跳表第 40 格 ＝ **腳本 opcode 35**：

```
0x1AB0E  ds:722Bh ← ds:4650h        ; 記下現在的時鐘（24-bit）
         ds:722Ch ← ds:4651h
         ds:722Dh ← ds:4652h
         ds:722Ah ← 1               ; 倒數開始
         [記錄區標頭 +0x2F] ← 0      ; **遭遇分母歸零**——之後不再擲遭遇
         call sub_1142B; jmp sub_1142B   ; 音效 7 播兩次
```

`+0x2F` 歸零這一手很說明問題：**自毀啟動之後基地裡不再有隨機遭遇**，
給玩家一條逃出去的路。

全 42 張地圖只有**一筆**記錄用得到 opcode 35：**資源 20 記錄 4**
（`TestSelfDestructScriptIsUniqueInData`）。資源 20 是科奇斯基地的反應爐層。

### 3.1 240 刻是多久

倒數的門檻是 `0xF0` ＝ 240 刻。資源 20 的記錄區標頭 `+0x36`（每步幾刻）是 **1**，
`+0x34/+0x35`（每步幾分鐘）是 0.25——所以是 **240 步、遊戲內一小時**。

## 4. 那一格怎麼變出來的：反應爐層的完整序列

資源 20 記錄 4 在出貨資料裡**沒有格子指到**。它是一連串改寫的終點，
而那一串就是整個遊戲的最後一道謎題。

四個角 (1,1)、(30,1)、(30,30)、(1,30)：

| 步驟 | 這一格 | 做什麼 | 改寫成 |
|---|---|---|---|
| 1 | (1,1) nibble 2 記錄 16 | 黑色圓柱，`USE` **黑星鑰匙** | nibble 12 記錄 0 |
| 2 | ↑ | 批次：(30,1) ← 黃色圓柱（記錄 17） | 自己 ← nibble 1 記錄 21 |
| 3 | (30,1) 記錄 17 | `USE` **新星鑰匙** | nibble 12 記錄 1 →(30,30) ← 記錄 18 |
| 4 | (30,30) 記錄 18 | `USE` **脈衝星鑰匙** | nibble 12 記錄 2 →(1,30) ← 記錄 19 |
| 5 | (1,30) 記錄 19 | `USE` **類星體鑰匙** | nibble 1 記錄 24「第 4 站已啟動。」→ nibble 12 記錄 6 |
| 6 | 四個角 | ← nibble 8 記錄 0 | 「按下按鈕以啟動安全程序 #1342-666。」 |
| 7 | 任一角 | 按 `R` | nibble 12 記錄 5：四個出口變牆、(1,1) ← 面板（記錄 2） |
| 8 | (1,1) | 面板答 **R** | nibble 1 記錄 26「階段 1 站已啟動。」→ 記錄 7 →(30,30) ← 面板 3 |
| 9 | (30,30) | 答 **Y** | 階段 2 → 記錄 15 →(1,30) ← 面板 4 |
| 10 | (1,30) | 答 **G** | 階段 3 → 記錄 16 →(30,1) ← 面板 5 |
| 11 | (30,1) | 答 **B** | 階段 4 ＋「警報！自毀程序已啟動。」→ nibble 12 記錄 3 |
| 12 | (30,1) | 記錄 3 的批次：四個出口重新打開、(16,9) 變傳送格 | **自己 ← nibble 6 記錄 4** |
| 13 | (30,1) | 踩上去 → **opcode 35** | 倒數開始 |

答錯任何一站就落到 nibble 1 記錄 25「啟動序列錯誤。程序中止。」，
那一格改寫成 nibble 12 記錄 6 ——四個角退回按鈕狀態，整段重來。

第 12 步的「自己 ← nibble 6 記錄 4」不是批次表裡的一筆，是 nibble 12 的**收尾**：
批次表跑完之後用「最後一筆的位移 ＋ 5」去改寫腳下那一格（`0x12C48` 的
`al ← bl + 5`），而記錄 3 那個位置的兩個 byte 正好是 `06 04`。

⚠ **每一步都要重踩一次那一格。** 地圖格就是這台直譯器的程式計數器
（`docs/re/71` §5.1）：一次踩踏跑一格指令，改寫完就結束這一步。
所以同一個角落要來回踩好幾次才走得完。

推論等級：**已確認**（資料逐筆解出，remake 走完整條鏈子驗過，
`TestCochiseEndgame`）。

## 5. remake 這一側：三段都是斷的

倒數這件事本身沒實作，而它前面還有兩段也沒接——**三段任何一段斷掉，
玩家都走不到結局**，而且症狀都不像壞掉：

| 斷點 | 症狀 |
|---|---|
| `USE` 成功之後沒有改寫地圖格 | 黑色圓柱說 `It works!`，但下一根圓柱永遠不會出現 |
| nibble 8 沒有呈現層 | 踩上控制面板只印一句 `CHOOSE.`，答不了任何一題 |
| opcode 35 沒實作 | 踩上去回報「這個指令還沒做」，倒數不會開始 |

三段都補上了：

| 位置 | 做了什麼 |
|---|---|
| `internal/game/selfdestruct.go` | `ArmSelfDestruct`／`SelfDestructDue`／`DisarmSelfDestruct` |
| `internal/game/script.go` | `OpStartTimer` |
| `internal/game/gates.go` | `Party.UseOn`：`USE` 的收尾（位移 4／條件式／6）與懲罰 |
| `internal/play/question.go` | nibble 8 的呈現層：單鍵與打字兩種模式、比對、分支改寫 |
| `internal/play/play.go` | 主迴圈每一輪檢查 `SelfDestructDue`（對應 `0x16C28`） |

**`ds:722Ah`–`ds:722Dh` 不進存檔**：它們是執行檔資料段的全域，不在存檔的
MSQ 資源裡。倒數中存檔再讀回來，倒數就沒了——原版就是這樣，照做。

## 6. 可重跑的完整指令

```bash
tools/go.sh test ./internal/play/ -run TestSelfDestructScriptIsUniqueInData -v
tools/go.sh test ./internal/play/ -run TestSelfDestructCountdownEndsTheGame -v
tools/go.sh test ./internal/play/ -run TestCochiseEndgame -v
tools/go.sh test ./internal/play/ -run TestEndingHasNoTriggerInData -v

# 問答的呈現層（單鍵與打字各一）
tools/go.sh run ./cmd/wl-play -script "map=1:3:2,down,S" -trace
tools/go.sh run ./cmd/wl-play -script "map=49:16:30,up,R,E,D,H,A,W,K,enter" -trace

# 跳表本體
python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xA4E0 60 --code
```

## 7. 這一輪學到的（寫成規則）

- **「資料裡沒有觸發點」不等於「觸發點還沒找到」。** 找了三輪都在資料側找，
  是因為前兩個消費點都吃資料。第三個消費點根本不看資料——它自己造一個索引。
  **要找的不是「哪一筆記錄是 kind 4」，而是「`al` 還能從哪裡變成 `0x84`」**：
  掃常數 `84h` 的寫入點比掃資料快得多。
- **一張表的消費點要全檔掃過再說「只有這幾個」。** 這一份的關鍵一步是
  掃 `ds:A4E0h` 一帶的所有定址存取，掃出第二個消費點（`ds:A4EAh` ＝ 表 ＋ 10），
  然後才在**別的地方**找到第三條路。
- **「這個機制沒接上」的症狀是別的機制看起來壞掉。** `USE` 沒有收尾改寫，
  看起來像「鑰匙沒用」；nibble 8 沒有呈現層，看起來像「這一格沒東西」。
  三個斷點沒有一個會在原地報錯——只有從頭走一遍才發現得了。
- **負面事實的門檻要成對。** `TestEndingHasNoTriggerInData`（資料裡沒有第 4 格）
  現在配上 `TestSelfDestructScriptIsUniqueInData`（啟動倒數的記錄只有一筆）。
  只留前者，下一輪會讀成「結局還沒接」；只留後者，會忘記第 4 格不能從資料來。

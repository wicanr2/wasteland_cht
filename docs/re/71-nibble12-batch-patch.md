# 71：nibble 12 是遠端批次改寫器

日期：2026-08-15 ｜ 接 `docs/re/70`（nibble 1 與商店入口的排除）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 見 `docs/re/01`。

`docs/re/26` §5 把 nibble 12 記成「讀記錄 `+0x00` → `sub_17920`（印訊息那一族）」。
那只是第一行。整支讀完之後它是這個遊戲裡**唯一會改寫腳下以外格子**的機制。

---

## 1. `0x12BD0` 全文

```
0x12BD0  bl ← 0；al ← 記錄[0]；sub_17920      ; 印訊息
0x12BDB  push ds:46FCh；push ds:46FDh          ; 保存沿用暫存（docs/re/69 §7）
0x12BE3  bl ← 1；ds:A4D1h ← bl                ; 批次表從 +0x01 起

loc_12BE5:
  al ← 記錄[ds:A4D1h] & 1
  ＝ 0 → 基準 (0, 0)                          ; 相對地圖原點
  ≠ 0 → 基準 (ds:46A6h, ds:46A7h)             ; 相對隊伍座標

loc_12C02:
  bl ← ds:A4D1h + 1
  dl ← 記錄[bl] + 基準x                       ; 目標 x
  bl++；bl ← 記錄[bl] + 基準y                 ; 目標 y
  al ← ds:A4D1h + 3
  call sub_17CFF(al, dl, bl)                  ; ← **改寫那一格**
  al ← 記錄[ds:A4D1h]
  bit7 設 → loc_12C48
  ds:A4D1h += 5；jmp loc_12BE5                ; 下一筆

loc_12C48:
  pop → ds:46FDh；pop → ds:46FCh              ; 還原沿用暫存
  al ← ds:A4D1h + 5
  ds:A4D0h 的 bit6 沒設 → call loc_163CA      ; 重畫
  jmp sub_169B1                               ; 最後改寫**自己這一格**
```

## 2. 批次表的格式

從記錄 `+0x01` 起，每筆 **5 bytes**：

| 位移 | 內容 |
|---:|---|
| `+0` | 旗標：**bit0** ＝ 座標相對隊伍（否則相對地圖原點）、**bit7** ＝ 最後一筆、bit6 ＝ 跳過重畫 |
| `+1` | 目標 x（加上基準） |
| `+2` | 目標 y（加上基準） |
| `+3` | 新的第 1 層 |
| `+4` | 新的第 2 層 |

`+3`／`+4` 走的是 `sub_17CFF`，所以 bit7 ＝ 不改、`0xFE`／`0xFD` ＝ 沿用
（`docs/re/69` §7）通通適用。**批次跑完之後暫存會被還原**——
`ds:46FCh`／`46FDh` 在進來時 push、離開時 pop，所以整批改寫不會污染下一次的沿用值。

批次結束的位移同時就是「改寫自己這一格」要用的位移（`al ← ds:A4D1h + 5`）。

出貨資料：42 張地圖共 **2,450 筆**批次改寫，每筆記錄多半 1–8 筆
（`TestNibble12BatchPatch` 涵蓋解析）。

## 3. 實跑

地圖 3 的 (28,16) 是 nibble 12 記錄 14，批次表第一筆把 (28,17)
從 nibble 1 記錄 0 改成 nibble 3 記錄 16：

```
tools/go.sh test ./internal/play/ -run TestNibble12BatchPatch -v
```

走上 (28,16) 之後 (28,17) 確實變了。資源 49 的記錄 5 更明顯——
一次把六格改成 nibble 0 與 nibble 4，形狀像「牆倒了、路通了」。

## 4. 對 remake 的意思

`internal/game/world.go` 的 `applyBatchPatch` 照 §2 實作，在 nibble 12 的
事件收尾跑，跑完再用 `batchPatchEnd` 算出的位移改寫腳下那一格。

在此之前 remake 對 nibble 12 只印一句訊息，**2,450 筆遠端改寫一筆都沒發生**——
症狀是「有些門永遠不會開」，而且不會有任何錯誤訊息。

## 5. 商店入口：資料側的四條路也掃完了

`docs/re/70` §4.2 確認程式碼那一側封閉之後，把「哪些改寫的目標是
(nibble 6, 設施記錄)」掃過四種來源：

| 改寫來源 | 位移算法 | 指到設施 |
|---|---|---:|
| nibble 1 的串列尾 | 訊息條數（`docs/re/70` §1） | 0 |
| nibble 2 的 `+0x04`／`+0x06` 與逐條件改寫表 | `docs/re/69` §4 | 0 |
| nibble 8 答對之後 | **`3 + 答案數 + 2 × 答對序號`**（`0x1522F` 直讀） | 0 |
| nibble 12 的批次表 | 每筆 `+3`／`+4` | 0 |

另外靜態掃「任何 section 記錄的改寫對指到 nibble 6 ＋ 設施編號」只有 2 處
（資源 12 的 `Vegas Lib.`、資源 34 的一筆未命名）。

⚠ nibble 8 的位移算法本身**已由程式碼確認**（`loc_1522F` 的
`al ← 3; adc al, ds:A653h; al ← ds:A651h << 1 + …`），
`docs/re/70` §4.3 裡「指到區塊外」的那幾筆是因為掃了**沒有格子指到**的記錄——
限縮到有格子的之後，174 處改寫裡指到 nibble 6 的 8 處，設施 0 處。

### 5.1 腳本的程式計數器就是地圖格

`docs/re/34` §1 說 nibble 6 腳本記錄的 `+0x01`／`+0x02` 是**下一步要跳到哪**，
而 `0x12C74` 的 `mov al, 1` → `sub_169B1` 是**用位移 1 改寫這一格**。
兩者是同一件事：**跑完一個指令就把這一格改寫成下一個指令的 (nibble, 記錄)**。

所以這台直譯器沒有另外的程式計數器——地圖格本身就是。
分支（op 1／3／15 把 `+0x03/+0x04` 或 `+0x05/+0x06` 搬進 `+0x01/+0x02`）
改的是「這一格接下來會變成什麼」。

這也解釋了設施記錄的 `+0x01`／`+0x02` ＝ `fd fd`：設施的「下一步」是
**自己**（沿用改寫前的值），所以離開商店之後那一格還是商店。

### 5.2 `SectionCount` 沒有低估 section 6

逐一列出 section 6 指標陣列的原始 word：資源 9 的 `base = 0x906`，
`[0][1][2]` ＝ `0x90c`／`0x90f`／`0x912` 連續，第 4 個 word 已經是記錄本體的內容。
`SectionCount` 算出的 3 是對的，`AG. store` 確實是記錄 2，而格子只指到 0 與 1。

⚠ **`SectionCount` 對 section `0x10` 不適用**：那個陣列存的是 opcode 不是位移
（`docs/re/34` §1），套「第一個非空指標指到哪」的推論會算出**負數**。
資源 9 的 `SectionCount(0x10)` ＝ −1159。要讀它得用 `SectionPointers`。

### 5.3 `+0x00` 是 section `0x10` 的索引，不是 opcode 編號

`sub_12C80` 的腳本路徑先 `sub_17CB1(bl ＝ 0x10, al ＝ 記錄 +0x00)`，
把 `ds:46AEh` 換成 section `0x10` 的第 al 個 word，**那個 word 才是 opcode**，
再拿它索引 `ds:A4EAh`。資源 9 的兩筆腳本記錄 `+0x00` 都是 0，
對到的 opcode 是 **26**（播音效，`docs/re/60` §8 已排除）。

### 5.4 遭遇那條也掃完了

`sub_13762` 的改寫位移由 `0x13633` 給定為 **`0x0A`**，來源是 `ds:46C6h` 指的
敵方記錄的 `+0x0A`／`+0x0B`，而且 `+0x0A` 的 bit7 設起來就不做
（`docs/re/68` §1 記的「敵方記錄的前兩個 byte」不精確）。

掃 42 張地圖的 `EnemyData`（94 bytes 一筆 ＝ 4 標頭 ＋ 3 組 × 30，`docs/re/37`）
共 336 筆：改寫成 nibble 6 的只有 1 筆（資源 29），不是設施。

**五條資料側路徑全部掃完，指到設施記錄的總共 2 處。**

### 5.5 強假說：商店與醫生在設施 3 的選單裡

`Ranger Ctr.`（資源 0 的 section 6 記錄 0，**有格子指到**，在 (55,62)）
的 `+0x00` ＝ `0x83` → 跳表索引 **3** ＝ `0x1A2C0`：

```
0x1A2C0  sub_18801
         bx ← 3；si ← 0
         [si + 0x7201] ← 記錄[bx]；si++；bx++；直到 si ＝ 0x0D
                                     ; 記錄 +0x03 起 13 bytes ＝ **設施名稱**
         sub_1A308                   ; 選單：ds:4689h ← 2（**兩個選項**）
         ds:CE49h ← ds:46FAh；ds:CE4Ah ← ds:46FBh
         al ← 3；sub_190A6            ; 畫面
         sub_1728C；sub_19727
         sub_18E90                   ; 讀鍵
         sub_17574                   ; 清單選單的按鍵處理（docs/re/53）
         jmp loc_1A2F4               ; 迴圈
```

所以**設施 3 是「有選單的設施」**，而不是某一種店。
這正好解釋為什麼商店與醫生沒有自己的格子——它們是**從設施 3 的選單進去的**，
跳表索引 0／1／2（醫生／商店／圖書館）由選單分派，不是由格子分派。

推論等級：**強證據**（畫面與選單迴圈是直讀的；「選單項就是醫生與商店」還沒讀到
分派的那一行）。下一步是 `sub_18801` 與 `sub_1A308` 設的兩張表
（`ds:468Dh`／`468Fh`）——`sub_1A308` 填的是 `ds:CE42h`／`ds:CE44h`，
而那兩個位址在開機時裝的是開場選單（`ds:CE12h` ＝ `CREATE DELETE PLAY`），
所以設施畫面一定在別處覆寫過它們。**先找覆寫點，再談選單項是什麼。**

## 6. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/n12.json 0x12BFE 0x12C02 0x12C2B 0x12C40

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/ans4.json 0x1522F 0x15242

tools/go.sh test ./internal/play/ -run TestNibble12BatchPatch -v
```

## 7. 這一輪學到的（寫成規則）

- **「讀了第一行」的筆記最危險。** nibble 12 在 `docs/re/26` 記成「印訊息那一族」，
  完全正確，也完全不夠——真正的內容在後面 120 個 byte。
  **一句話的註解要標明它涵蓋到哪一行。**
- **改寫不一定作用在腳下。** 先前掃「哪些改寫指到設施」時，
  所有來源都假設目標是隊伍所在的格子；nibble 12 的目標是任意座標。
  **列舉一種機制的變化形時，先確認它的作用對象是不是固定的。**
- **限縮取樣範圍會讓假陽性消失，但要說明為什麼那個範圍才對。**
  nibble 8 掃全部 section 記錄會撈出指到區塊外的目標，
  限縮到「有格子指到的記錄」之後乾淨——因為沒有格子指到的那些
  多半根本不是有效的 nibble 8 記錄（`SectionCount` 可能高估）。

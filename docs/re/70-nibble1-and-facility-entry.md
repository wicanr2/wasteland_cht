# 70：nibble 1 的氛圍敘述，與商店入口的排除紀錄

日期：2026-08-15 ｜ 接 `docs/re/69`（沿用暫存）、`docs/re/60` §8（商店入口）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 見 `docs/re/01`。

---

## 1. `sub_16CD0`：nibble 1 是一串訊息 ＋ 一次改寫

```
0x16CD0  sub_17DC7                 ; dl／bl ← 這一格的地圖座標
0x16CD3  sub_17CD2                 ; al ← 目前 nibble、dl ← 目前記錄
0x16CD6  bl ← 1
0x16CD8  al ＝ ds:4716h 且 dl ＝ ds:4717h → bl ← 0   ; 同一種格子 ＝ 不印
0x16CE6  ds:B1C1h ← bl（要不要印）；ds:B1C0h ← 0（計數器）

loc_16CEC（迴圈）:
  ds:B1C0h ← bl
  al ← 記錄[bl]
  push ax；al &= 7Fh
  ds:B1C1h ≠ 0 → sub_17920(al)     ; 印這一條
  bl ← ds:B1C0h + 1；pop ax
  al 的 bit7 沒設 → 回圈

0x16D0F  al ← bl                   ; ＝ 訊息條數
0x16D11  jmp sub_169B1             ; → sub_17CFF(al, 隊伍座標)
```

兩件事：

- **訊息可以有很多條**：從記錄 `+0x00` 起逐 byte，**bit7 設的那一條是最後一條**。
- **訊息串列的長度就是收尾改寫要用的位移**——串列後面緊接著一對改寫 byte。
  這與條件閘「`0xFF` 之後接改寫表」（`docs/re/69` §4）是同一種佈局慣例。

`sub_169B1` 只有三行：`dl ← ds:46A6h`、`bl ← ds:46A7h`、`jmp sub_17CFF`，
也就是**改寫隊伍腳下那一格**。它是 nibble 1／4／6／12 與傳送共用的收尾。

### 1.1 `ds:4716h`／`4717h` ＝ 上一次觸發的 (nibble, 記錄)

比對的不是座標。全檔三處碰它：`0x13D3A`（`sub_13C58`）與 `0x169FE`
（移動主流程 `sub_163F6`）寫，`0x16CD8` 讀。所以**連續踩過同一種格子只印一次**，
走過別的格子之後再回來會重印。

### 1.2 實跑

地圖 2 的 (24,9)–(28,9) 連續五格都是 nibble 1 記錄 6，那筆記錄有兩條訊息：

```
tools/go.sh run ./cmd/wl-play -script "map=2:23:9,right,right,right,right,right" -trace
   2 right  (24, 9)  You are beside a low stage.Several shapely dancers prance out onto the stage.
   3 right  (25, 9)
   4 right  (26, 9)
   5 right  (27, 9)
   6 right  (28, 9)
```

兩條一起印、後面四格靜音，與資料一致（`TestNibble1MessageList`）。

remake 這一側原本把 nibble 1 當成 `EventNone`（依據是「遠看才顯示」的推測），
**整個廢棄建築群一句敘述都沒有**。現在照 `sub_16CD0` 實作。

## 2. nibble 6 與 nibble 12 的收尾

| nibble | 訊息 | 收尾改寫的位移 |
|---:|---|---|
| 1 | 記錄 `+0x00` 起的串列，bit7 結束 | 訊息條數 |
| 6 | 由 `sub_12C80` 分派（設施或腳本） | **1**（`0x12C74` 的 `mov al, 1`） |
| 12 | 記錄 **`+0x00`**（`0x12BD0` 的 `mov bl, 0` → `sub_17920`） | 還沒讀完 |

nibble 12 的字串編號是記錄 `+0x00`，**不是這一格的第 2 層值**——remake 原本拿錯。

## 3. `0xFE`／`0xFD` 在出貨資料裡用得很多

`docs/re/69` §7 解出「沿用上一格改寫前的值」之後，把它掃過真正會被當成改寫位移
的組合，**逐 section 記錄**統計：

| 用在哪 | 處數 |
|---|---:|
| nibble 1（位移 ＝ 訊息條數） | 89 |
| nibble 6（位移 1） | 28 |
| nibble 2 的 `+0x04`／`+0x06` | 13 ／ 13 |
| nibble 4／9／12 的 `+0x01`／`+0x02` | 10 |

最大的用戶是**設施記錄**：30 筆設施幾乎都是 `+0x01`／`+0x02` ＝ `fd fd`，
意思是「跑完之後把這一格改回原樣，並回報沒改」。設施要能重複進出，
這是唯一的做法——原版沒有旗標，離開商店之後那一格必須還是商店。

⚠ 只掃「有格子指到的記錄」會得到零。**沒有格子指到的記錄一樣會被執行**
（改寫可以把一格變成任何一筆記錄），所以資料面的統計要逐 section 走，
不是逐格走。

## 4. 商店與醫生的入口：程式碼那一側已經封閉

32 筆設施記錄（`+0x00` 的 bit7 設）裡**只有 2 筆有 nibble 6 的格子指到**
（`Ranger Ctr.`、`Chopper Sim.`）。這一輪新排除的：

| 路 | 結果 |
|---|---|
| Quartz（**地圖 1**）的 144 個傳送格 | 目標是地圖 2／3／4／6，以及 20 個 bit7 編號——**後者全部映射到資源 5** |
| 資源 5（建築內部） | 沒有一格 nibble 6；section 6 的 4 筆記錄 bit7 全沒設。**它是「廢棄建築」的通用內部地圖** |
| 資源 9 的 nibble 6 格子（指到記錄 0／1） | 那兩筆腳本跑完改寫成 **nibble 2 記錄 3／4** ＝ 天上掉巨大番茄的陷阱，不是商店 |
| `sub_12C80` 的第二個呼叫端 `sub_1CB30` | 遊戲時間減 `ds:722Bh` 起算超過 `0xF0` → `sub_12C80(al ＝ 0x84)`，**寫死的設施 4**，不是通用入口 |

`AG. store` 是資源 9 的 section 6 記錄 2，而資源 9 的 nibble 6 格子只指到 0 與 1。

### 4.1 設施跳表與腳本跳表是同一張表，差 5 個 word

`sub_12C80` 的兩條路讀的是 `ds:A4E0h`（bit7 設）與 `ds:A4EAh`（bit7 沒設），
而 `0xA4EA − 0xA4E0 ＝ 10` ＝ **5 個 word**。把 `ds:A4E0h` dump 出來：

| 索引 | 位址 | |
|---:|---|---|
| 0 | `0x1C260` | 醫生（`80` 那一族：`Infirmary`、`Patch em' up`、`Old Doc Bobs`、`NUCLEAR AID`） |
| 1 | `0x1BE50` | 商店（`81`：`AG. store`、`Store`、`Market`、`Blackmarket`） |
| 2 | `0x1BBA0` | 圖書館／訓練（`82`：`Library`、`New Thoughts`、`HOLY KNOWING`） |
| 3 | `0x1A2C0` | ？ |
| 4 | `0x1B4F0` | `sub_1CB30` 計時到期寫死呼叫的那一個 |
| 5 以上 | `0x1A470`、`0x1A4F4`… | **就是腳本 opcode 0、1、…**（`0x1A699` ＝ `docs/re/34` 的 opcode 7） |

所以「44 個腳本 opcode 沒有一個開設施」這條排除仍然成立——
opcode 就住在同一張表的後半，而**真正的設施只有索引 0–4**。

### 4.2 程式碼那一側已經封閉

五個設施函式的 16-bit 位移（`0xC260`、`0xBE50`、`0xBBA0`、`0xA2C0`、`0xB4F0`）
掃全檔指令**全部零命中**——它們只透過跳表被呼叫，而跳表只有 `sub_12C80` 讀，
`sub_12C80` 只有兩個呼叫端（nibble 6 的格子、`sub_1CB30` 的計時）。

**所以錯的一定在資料那一側**：某些格子的 (nibble 6, 設施編號) 組合
不在初始地圖上，是遊戲中才被改寫出來的。

### 4.3 改寫來源掃完的結果

掃過所有靜態改寫來源，目標是「nibble 6 ＋ 這張地圖的設施記錄編號」的只有 **2 處**
（資源 12 的 nibble 9 記錄 1 → `Vegas Lib.`、資源 34 的 nibble 12 記錄 9）。

nibble 8（問答）答對之後的改寫（位移 ＝ `0x03 + 答案數 + 2n`，`docs/re/46` §4.1）
指到 nibble 6 的有 14 處，但目標編號（10–14、42、192…）不是設施：
資源 33 的 10–14 確實存在但是腳本，42 與 192 則指到區塊外——
**那幾筆的位移算法還沒對，不能當成證據**。

### 4.4 還沒走的路

- `sub_13762`／`sub_13C58` 那條改寫來源（敵方記錄的前兩個 byte）。
- nibble 12 的處理 `0x12BD0` 只讀到 `0x12C02`，它的改寫位移還沒確定
  （`sub_17CFF` 的呼叫端 `0x12C2B` 就在那一段裡）。
- nibble 8 答案後續動作的位移算法要對著 `sub_15160` 重讀一次，
  現在的算法會撈出指到區塊外的目標。
- **實機**：`docs/re/60` §8 寫的那條路仍然有效，而且現在目標更明確——
  走進一家商店之後，用 remake 載入同一張地圖同一格，看它的 (nibble, 記錄)
  與初始資料差在哪。差異本身就會指出是誰改寫的。

## 5. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/nibble1b.json 0x16CD0

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_range_refs.py \
  workplace/analysis/dumps/lastcell.json 0x4716 0x4717

tools/go.sh test ./internal/play/ -run TestNibble1MessageList -v
tools/go.sh run ./cmd/wl-play -script "map=2:23:9,right,right,right,right,right" -trace
```

## 6. 這一輪學到的（寫成規則）

- **「沒有格子指到」不等於「到不了」。** 這個遊戲的狀態就是改寫格子，
  所以一筆記錄可以在遊戲中途才出現在地圖上。統計可達性時要把**改寫的目標**
  一起算進去，只數初始格子會少算一整類內容。
- **同一個問題換一個統計單位，答案會相反。** `0xFE`／`0xFD` 逐格掃是 0、
  逐記錄掃是 153。兩個數字都對，錯的是把其中一個當成「這個機制沒用到」。
- **「還沒讀完的函式」與「讀了前幾行的函式」要分開記。** nibble 1 的處理在
  `docs/re/26` §5 只有一句「比對上一次的位置，避免重複觸發」——
  那句話讓它看起來已經解了，實際上整個訊息串列與收尾改寫都沒讀到。
  **筆記要標推論等級，「入口與前幾行」不算解過。**

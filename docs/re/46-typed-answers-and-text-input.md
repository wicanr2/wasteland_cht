# 46：打字回答、文字輸入與字串比對

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`CONTEXT.md` §7.2 掛著「打字型密語的比對程式碼還沒定位」。找到了，而且順帶
把「角色名字怎麼輸入」與 `docs/re/29` §3 對 nibble 8 的描述一起補正。

---

## 1. 結論

| 事實 | 等級 |
|---|---|
| 文字輸入本體是 `loc_17750`，緩衝區與長度上限都是全域變數（§2） | 已確認 |
| **每個按鍵在讀進來的當下就轉大寫**（`0x18FF0`） | 已確認 |
| 字串比對是 `sub_18D8E`，逐 byte 全等（§3） | 已確認 |
| nibble 8 是**問答**，記錄 `+0x00` 的 bit7 決定「按一個鍵」還是「打一行字」（§4） | 已確認 |
| 打字模式的上限是 **16 bytes**，角色名字是 **13 bytes**（§2） | 已確認 |
| `sub_18D8E` 的呼叫端 IDA 只認出一個，**全檔掃描有兩個**（§3） | 已確認 |

`docs/re/29` §3 把 nibble 8 寫成「多選一的選單」，只讀到單鍵那一半，
而且把 `sub_1721B`（挑隊員 1–9）誤當成選項的範圍檢查——那支根本不在這條路上。
該節已改寫。

## 2. 文字輸入：`sub_17748` / `loc_17750`

兩個入口共用同一支：

| 入口 | `ds:467Bh` | `ds:0B264h` |
|---|---|---|
| `sub_17748`（`0x17748`） | 0 | `0xFF` ＝ 也收滑鼠 |
| `loc_17750`（`0x17750`） | 0 | 0 ＝ 只收鍵盤 |

`ds:467Bh` 非 0 時**只收 `'0'`–`'9'`**（`0x177B9`），是數字專用模式。

進來的參數全部是全域變數：

| 變數 | 意思 |
|---|---|
| `ds:4680h` | 緩衝區位址 |
| `ds:4684h` | 長度上限（含結尾的 NUL 位置） |
| `ds:4682h` | 目前長度（進來時被清成 0） |
| `ds:4672h` | 顏色／欄位 |

按鍵處理（`0x1779C` 起）：

```
0x1B  ESC        → al ← 0xFF 回去（取消）
0x08 / 0xFF      → 刪一個字（長度為 0 時不動）
0x0D  Enter      → 從尾巴往回吃掉空白，然後回 al ← 0（成功）
< 0x20           → 丟掉
其他             → al &= 0x7F，寫進緩衝區；長度到 ds:4684h 就不再收
```

⚠ **Enter 會 trim 尾端空白**（`0x17809`–`0x1781B`）。前導空白不會被吃掉。

### 2.1 按鍵在讀進來的當下就轉大寫

`sub_18EFE`（鍵盤讀取本體，`sub_18E90` 包它）的**最後一段**：

```
0x18FF0  cmp  al, 61h        ; 'a'
0x18FF2  jb   short loc_18FFC
0x18FF4  cmp  al, 7Bh        ; '{'
0x18FF6  jnb  short loc_18FFC
0x18FF8  sub  al, ds:0C058h  ; ← 0x18EFE 開頭就把它設成 0x20
0x18FFC  mov  ds:4667h, al
```

所以**整個遊戲拿到的字母永遠是大寫**，比對時不需要（也沒有）另外轉換。
密語在資料裡也是大寫（`BIRD`、`DIPSTICK`、`MUERTE`），兩邊對得起來。

## 3. 字串比對：`sub_18D8E`

```
0x18D8E  bl ← 0
0x18D90  di ← ds:4680h ; al ← [bx+di]
0x18D96  di ← ds:4665h ; al |= [bx+di]
0x18D9C  jz  locret          ; 兩邊同時到 NUL → al ＝ 0（相等）
0x18D9E  di ← ds:4680h ; al ← [bx+di]
0x18DA4  di ← ds:4665h ; cmp al, [bx+di]
0x18DAA  jnz loc_18DB0
0x18DAC  inc bl ; jnz 迴圈
0x18DB0  al ← 0xFF           ; 不相等
```

比 `ds:4680h` 與 `ds:4665h` 兩個 NUL 結尾字串，**逐 byte 全等**才回 0。
沒有大小寫折疊、沒有長度容忍、沒有前後空白處理（空白在輸入端就被 trim 掉了）。

⚠ **IDA 的 xref 只認出一個呼叫端**（`0x1CA3E`）。全檔掃描找到兩個：

```
$ python3 tools/scan_callers.py workplace/analysis/unpacked/wl.merged.exe 1000:8D8E
同段 near call：2 筆
  檔案 0x052c2  線性 0x15212   ← 密語比對，xref 圖上看不到
  檔案 0x0caee  線性 0x1ca3e   ← 角色重名檢查
```

漏掉的那個正是本文要找的東西。**「只有一個呼叫端」這種話沒有全檔掃描佐證就不要寫。**

## 4. nibble 8 ＝ 問答（`0x15160`）

地圖第 1 層 nibble 8 的處理程式。**不是選單，是問答**，而且有兩種模式。

```
0x15160  bl ← 3；dl ← 0
0x15164  迴圈：dl++，讀記錄[bl]，bit7 沒設就 bl++ 繼續    ; 數出有幾個「答案」
0x15174  ds:0A653h ← dl                                   ; 答案數
0x15178  call sub_19727                                   ; 開文字框（控制碼 0x07）
0x1517D  al ← 記錄[+0x00]；push
0x15184  sub_178A0(al & 0x7F)                             ; 印題目
0x1518A  test al, al ／ jns 0x151B6                       ; ← bit7 決定模式

; ── 模式 A：bit7 設 → 按一個鍵 ──
0x1518E  ds:7DF3h ← 0x0404 ／ sub_18E90                   ; 等一個鍵
0x1519D  Enter 或 ESC → 0x1524B（離開）
0x151A8  ds:7931h ← 鍵 & 0x7F ；ds:7932h ← 0              ; 當成一個字的答案

; ── 模式 B：bit7 沒設 → 打一行字 ──
0x151B6  sub_1A0C5 ；印 '>'（sub_19DC3）
0x151BE  ds:4680h ← 0x7931                                ; 輸入緩衝區
0x151C4  ds:4684h ← 0x10                                  ; ← 上限 16 bytes
0x151C9  sub_17748                                        ; 讀一行
0x151CC  回傳非 0（ESC）→ 0x1524B（離開）

; ── 比對 ──
0x151D0  緩衝區第一個 byte 是 0 → 離開
0x151D7  ds:0A651h ← 0                                    ; 命中索引
0x151DC  bl ← 3
0x151DE  ds:0A650h ← bl
0x151E2  al ← 記錄[bl] & 0x7F ；sub_16D14(al)             ; 選第 al 條字串
0x151ED  逐字元 sub_17B8F → 存進 ds:79B1h（到 0 為止）      ; 解出這個答案
0x15206  ds:4680h ← 0x79B1 ；ds:4665h ← 0x7931
0x15212  sub_18D8E                                        ; 比！
0x15215  相等 → 0x1522F
0x15219  ds:0A651h++ ；記錄[ds:0A650h] 的 bit7 還設著就試下一個

; ── 分支 ──
0x1522F  ds:0A650h ← 3 + ds:0A653h                        ; 答案清單之後
0x15239  al ← ds:0A651h × 2 + ds:0A650h
0x15242  sub_169B1(al)                                    ; 依「第幾個答案」決定後續
```

**記錄的形狀**：

| 位移 | 內容 |
|---|---|
| `+0x00` | 題目的字串編號；**bit7 ＝ 用單鍵模式** |
| `+0x03` 起 | 一串答案的字串編號，最後一個的 bit7 是結束標記 |
| `+0x03 + 答案數` 起 | 每個答案兩個 byte 的後續動作 |

比對是**照順序試，第一個相等的贏**；全部不中就落到「答案數」那一格
（`ds:0A651h` ＝ 答案數），也就是「答錯」的那一支。

### 4.1 分支要做什麼：`sub_169B1` ＝ 改寫腳下那一格

`sub_169B1(al ＝ 記錄裡的位移)` 只是把玩家所在的座標（`ds:46A6h`／`46A7h`）
填好，轉給 `sub_17CFF`——**那是六個地方共用的「改寫地圖格」機制**
（條件串列、腳本、寶箱都走它）。

```
0x17CFF  ds:0BA0Ch/0Dh ← 座標
0x17D09  al ← 記錄[al]                ; 第一個 byte
0x17D0F  == 0xFE → sub_17D34          ; 改用上一次算出來的 ds:46FCh/46FDh
0x17D13  == 0xFD → sub_17D34 後 clc   ; 同上，但回報「沒改」
0x17D17  ds:46B3h ← al                ; 新的第 1 層 nibble
0x17D1A  ds:46B4h ← 記錄[al+1]        ; 新的第 2 層記錄編號
0x17D21  → sub_17D47 / sub_17CD2 / sub_17D50
```

`sub_17D50` 才是真的動手的那一支：

```
0x17D50  al ← ds:46B3h
0x17D53  bit7 設 → clc; retn          ; ← 這一格不改
0x17D5F  sub_17D7A(al)                ; 寫第 1 層的 nibble（讀出、清舊、or 新值、寫回）
0x17D66  sub_17C72(bl)                ; 算第 2 層那一列的位址 → ds:46ACh
0x17D74  [ds:46ACh + x] ← ds:46B4h    ; 寫第 2 層的記錄編號
0x17D76  stc; retn
```

**所以「答對第 N 個」的兩個 byte 是（新的地形 nibble、新的記錄編號）。**
守衛問密語、答對之後那一格從「擋路」變成「可以走的通道」，就是這樣做的——
不是設一個旗標，是**直接改地圖**。

第一個 byte 的 bit7 設起來 ＝ 這一支什麼都不改（例如「答錯只印一句話」）。

## 5. 角色名字：同一支輸入

`sub_1CAA9`（`0x1CAA9`）：

```
0x1CAA9  sub_1786E              ; 印提示（ds:4680h 先設好）
0x1CAAC  ds:4680h ← 0x7931      ; 同一個輸入緩衝區
0x1CAB2  ds:4684h ← 0x0D        ; ← 名字上限 13 bytes
0x1CAB7  ds:4672h ← 0x0F
0x1CABD  jmp loc_17750          ; 只收鍵盤
```

`sub_1CAC0` 把 `ds:7931h` 的 14 bytes 抄進角色記錄 `+0x00`。
`sub_1CA14` 拿 `sub_18D8E` 逐一比隊伍裡其他人的名字，**重名就要求重打**。

## 6. 對中文化的意義（硬約束）

- **密語答案不能翻譯。** 玩家打的字經過 `and al, 0x7F` 與大寫轉換，
  是純 ASCII；比對又是逐 byte 全等。答案字串一旦變成中文，玩家永遠打不出來。
  要嘛保留英文原文（題目可以翻，答案留原字），要嘛改輸入層——
  這是**取捨，不是自由選擇**，要寫進規格。
- **打字上限 16 bytes**。Big5 一個中文字兩個 byte，等於 8 個中文字；
  角色名字 13 bytes ＝ 6 個中文字（第 13 byte 放不下半個字，會截斷成亂碼）。
  重製版要改成**按字數而不是按 byte 數**限制，並記下與原版的差異。
- **`and al, 0x7F` 會把 Big5 的高位 byte 打壞**（`0xA4` → `0x24`）。
  中文輸入不能走原版這條路，要在輸入層另外處理。
- 單鍵模式的鍵與 `\x10` 的列熱鍵表（`docs/re/43` §2）是**兩個不同機制**，
  不要混在一起改。

## 6.1 實際掃出來的數字

`tools/summarize_questions.py` 掃 42 個區塊的 section 8：

| | 題數 |
|---|---:|
| 單鍵（記錄 `+0x00` bit7 設） | **117** |
| 打字 | **68** |
| 標成打字、但沒有一條答案打得出來 → 不是問答 | 28 |

「打得出來」不是猜的判準，是原版自己的上限：輸入緩衝區 16 bytes，
所以超過 15 個字元的字串永遠比不中；輸入層丟掉 `< 0x20` 的字元，
所以含控制碼的也一樣。用這條過濾之後，剩下的答案長這樣：

```
MUERTE  UGLY  SQUINT  MULEFOOT  THANATOS  KAPUT  DIPSTICK  PROTEUS
ACAPULCO  ROSEBUD  MOTEKIM  RUN  TOAST  THE LETTER R  11-16-27  …
```

——密語、暗號、謎題答案，全部是大寫英文。完整清單在
[`generated/ida94/questions.md`](generated/ida94/questions.md)。

**這 126 條字串已經接進翻譯的建置守則。** `tools/summarize_questions.py`
另外產生 `translations/must-not-translate.tsv`，`tools/build_lang.py` 讀它，
譯文與原文不同就整個建置失敗。實測過現有的 126 條**全部維持英文原字**，
守則也實測會咬（把 `MUERTE` 改成「死亡」就擋下來）。

⚠ 這種錯**不會有任何徵兆**——遊戲照跑、畫面照顯示，只有那道關卡永遠過不了。
所以守則的清單檔不存在時 `build_lang.py` 直接失敗，不靜靜跳過。

## 7. 還沒解的

- `sub_18EFE` 裡 `ds:CA62h`–`ds:CA64h` 那組按鍵錄放（巨集）機制。

## 8. 可重跑的完整指令

```bash
# sub_18D8E 的呼叫端（xref 圖漏掉一個，要用全檔掃描）
python3 tools/scan_callers.py workplace/analysis/unpacked/wl.merged.exe 1000:8D8E

# 本文引用的函式
WL_IDA_TARGET=$PWD/workplace/analysis/ida94/wl.merged.exe \
  tools/ida.sh run tools/ida/export_function.py <輸出>/f.json \
  0x18D8E 0x18E90 0x18EFE 0x1CAA9 0x1CAC0 0x1CA14 --callers

# 42 個區塊的問答清單 ＋ 產生翻譯的守則清單
python3 tools/summarize_questions.py workplace/analysis/unpacked/wl.unpacked.exe \
  workplace/orig/wastland/game1 workplace/orig/wastland/game2 \
  docs/re/generated/ida94/block-strings.json docs/re/generated/ida94/questions.md

# nibble 8 的處理程式與文字輸入本體 IDA 沒建成函式，要強制分析
WL_IDA_TARGET=$PWD/workplace/analysis/ida94/wl.merged.exe \
  tools/ida.sh run tools/ida/export_forced.py <輸出>/g.json \
  0x15160 0x151A8 0x151D0 0x17748 0x1779C 0x17806 0x17809
```

## 9. 這一輪學到的（寫成規則）

- **「這支只有一個呼叫端」要用全檔掃描證明，不能用 IDA 的 caller 清單。**
  `sub_18D8E` 的第二個呼叫端就是這一題的答案，而它不在 xref 圖上。
  這條在 `CLAUDE.md` 已經寫過（「唯一／只有一處沒有全檔掃描佐證就不要寫」），
  這次是它第一次真的擋下錯誤結論。
- **讀到一半就命名，會把另一半藏起來。** nibble 8 被寫成「多選一選單」，
  是因為只讀到 `0x1518E` 的單鍵分支就停了；`jns` 的另一邊整個打字流程
  在文件裡等於不存在，連帶讓「密語比對沒定位」多掛了好幾輪。
  **看到條件跳躍，兩邊都要走完再命名。**
- **找不到 X 時，先找「X 必須用到的東西」。** 直接找密語比對沒有頭緒，
  但「比對兩個字串」這件事一定要有一支字串比較函式——先掃 `cmpsb`／`scasb`
  （零命中，證明是手寫迴圈），再從已知的重名檢查回頭找那支手寫迴圈，
  它的另一個呼叫端就是答案。

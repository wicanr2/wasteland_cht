# 96：結局 —— 它掛在設施跳表的第 4 格

日期：2026-08-16 ｜ 接 `docs/re/23` §9（`END.CPA` 的畫面）、`docs/re/79`（設施跳表）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`END.CPA` 的第一段是畫面（`docs/re/23` §9）。這一份解開**第二段**、
播放器、四段敘述文字，以及最後一個問題：**誰觸發結局**。

---

## 1. 觸發：走進一格「設施」

`sub_1B7FE` 的唯一呼叫端是 `0x1B51D`，而 `0x1B4F0`（那一段的開頭）
**沒有任何 near call 或 jmp 指到它**——自己掃 `E8`／`E9`／`EB` 全檔零命中，
而同一支腳本對 `0x1B735` 掃出 11 筆，所以不是掃描的問題。

它是**表裡的一個 word**：`0xB4F0` 在全檔只出現一次，在 `ds:A4E8h`。
往前退到那張表的頭是 **`ds:A4E0h`**，消費點在 `0x12C8E`：

```
0x12C8C  shl bx, 1
0x12C8E  mov cx, [bx+0A4E0h]      ; ← 索引跳表
0x12C92  xor bx, bx
```

`0x12C80` 就是 `sub_12C80`（`docs/re/75` §2 的腳本／設施分派）。
表的前 5 格是**設施種類**、第 5 格起才是腳本 opcode
（`FacilityKindCount ＝ 5`）——所以：

| 索引 | 值 | 是什麼 |
|---:|---|---|
| 0 | `0xC260` | 醫院 |
| 1 | `0xBE50` | 商店 |
| 2 | `0xBBA0` | 訓練師 |
| 3 | `0xA2C0` | Ranger Center |
| **4** | **`0xB4F0`** | **結局** |

remake 先前把第 4 種叫 `FacilityUnknown`「身分未定」，圖片編號 `-1`
——看起來只是一種沒有店面的設施，走進去會出現一個空畫面。
它其實是遊戲的終點：**走進那一格，遊戲就結束了**。

推論等級：**已確認**（跳表的位址是全檔掃描 ＋ 消費點的定址方式一起定的）。

## 2. 播放序列（`0x1B4F0`）

```
逐人：角色記錄 +0x4B 的 bit0 ← 1          ; 參與過摧毀 Base Cochise，見 §5
sub_1B7FE                                  ; 載 END.CPA 兩段
ds:D168h ← 0
dx ← 3Ch；sub_1B735                        ; 播 60 tick，**按鍵不中斷**
sub_1142B
dx ← 3Ch；sub_1B735
ds:D168h ← 1                               ; 之後按鍵可以跳過
四次：sub_162C7；sub_1B7B7(al ＝ 1…4)；dx ← 96h；sub_1B735
```

`sub_1B7B7` → `sub_1B7C9`：把字串表基址 `ds:46B0h` 暫時換成 `ds:D173h`
（那個 word 裡放的是 `0xD18E`）再叫 `sub_17920(al)`，印完換回來。
`ds:D18Eh` 就是 `ExeStrings()` 的第 **4** 張表「結局敘述」，第 1–4 條：

```
[1] "Shuddering explosions rock the base,\rfire blossoms throughout every\rdoorway.\r\r"
[2] "Everywhere walls and supports buckle\rand crumble in the explosions.\r\r\r"
[3] "Debris and shrapnel fly everywhere\rkilling everything it touches.\r\r\r"
[4] "You can almost imagine robots\rscreaming as they realize they, too,\rare mortal.\r\r"
```

同一張表的第 5 條是 `"\v was killed in the blast.\r\r\r"`——**結局會點名死者**，見 §2.1。

### 2.1 清算：Base Cochise 裡的人全部死在爆炸裡

四段敘述之後（`0x1B571`）逐組清算：

```
ds:D171h ← ds:4654h              ; 記住目前是哪一組
逐組（ds:4654h 從 0 到 ds:4657h）：
  al ← [ds:46B7h + 0Ah]          ; **那一組的地圖編號**（槽表 +0x0A）
  al ＜ 10h → 跳過
  al ≥ 15h → 跳過
loc_1B595:                       ; 這一組在 Base Cochise 裡
  ds:4653h ← 那組人數（sub_171E3）
  loc_1B5A6（逐人）:
      al ← 34h；sub_190A8        ; 載圖 52
      sub_19614(1)               ; 選第 1 個人
      [記錄 +1Dh] ← 0；[+1Eh] ← 0 ; **CON 歸零**
      sub_1B7B7(5)               ; 「\v was killed in the blast.」
      dx ← 69h；sub_1B735        ; 停 105 tick
      sub_172D4(1)               ; 把他移出隊伍
      ds:4653h ≠ 0 → 再來一個
  sub_1B7EC                      ; 槽表 14 bytes 全歸零
  ds:4657h−−                     ; 少一組
```

**判準是地圖編號 `0x10` ≤ 地圖 ＜ `0x15`**（`0x1B58A` 的兩個 `cmp`）——
Base Cochise 的五張地圖。留在外面的隊伍活下來。

⚠ **不是「全隊一律陣亡」。** 依地圖判定這件事看起來像細節，
但它正是 §5 那個 `+0x4B` 旗標存在的理由：活著回去的人才有人可以表揚。

刪掉一組之後 `0x1B5F2` 會把後面的組往前搬（`bl` 從 `0x1B` 數到 `0x0D`，
15 bytes 一組），並修正 `ds:D171h`——**記住的那一組可能已經被往前搬過**。

## 3. 第二段 ＝ 動畫腳本

`sub_1B7FE` 把第二段解到 `seg003:0x5120`，接著設好游標：

```
ds:D160h ← 5122h        ; **跳過開頭那個長度 word**
ds:D162h ← 0            ; 計時器
ds:D166h ← 0            ; 迴圈記憶點（還沒設）
ds:D165h ← 0Ch          ; 播滿 12 格才記迴圈點
```

然後 `si ＝ 920h`／`di ＝ si + 90h`／`cx ＝ 23B8h` 的
`lodsw; xor [di], ax` 迴圈——**列間 XOR delta 就地解開**，
`0x90` ＝ 144 ＝ 這張圖的 stride，與 `docs/re/23` §9 對上。

`sub_1B735(dx ＝ tick 數)` 是播放器：

```
每次 BIOS 0040:006C 變動（一個 tick）：
  word es:[si] ≠ 計時器 → 計時器 ++，這一 tick 什麼都不做
  相等 → 播這一格：
      si += 2
      while word es:[si] > 0:  sub_1004B（畫一個元素）; si += 6
      計時器 ← 0；si += 2
      word es:[si] < 0 → si ← ds:D166h        ; 迴圈
ds:D168h ≠ 0 且有按鍵 → 直接返回
```

### 3.1 一個元素 6 bytes

`sub_1004B` 是 overlay slot 25（`0x10FA7`，`docs/re/04`）：

```
lodsw → di ← ax + 141h            ; **螢幕**位移，0x141 ＝ 8 × 40 ＋ 1 ＝ (8, 8)
sub_10FD3: 讀 4 bytes ＝ 8 個 nibble，位切片成四個平面
           bl／bh／cl／ch ＝ 平面 0／1／2／3，各自寫一次
```

所以元素 ＝ `word 螢幕位移` ＋ **4 bytes ＝ 8 個像素**（一個 byte 兩個像素、
高 nibble 在左，與全遊戲一致）。

⚠ **位移的列寬是螢幕的 40 bytes，不是這張圖的 36。**
兩個都除得出「看起來合理」的列號，差別只在畫面右緣——
用 40 算，全部 2,433 個元素都落在這張圖的 288 像素內；用 36 算會有元素跑到圖外。

### 3.2 版面對得起來

整份 14,664 bytes 解析後剛好用完：

```
2（長度 word）＋ 15 格 × 4（延遲 ＋ 終止字）＋ 2,433 元素 × 6 ＝ 14,660
```

剩下 4 bytes 是結束標記與尾巴。**這個等式就是格式正確的證據**——
少讀一個欄位或多讀一個 byte 都會讓它對不齊。

推論等級：**已確認**（播放器與元素解碼逐指令讀完，版面等式在實檔上成立）。

## 4. remake 這一側

**已接**：`Rom.EndAnim`（`internal/assets/endanim.go`）＋
`Scene.BeginEnding`／`TickEnding`（`internal/play/ending.go`），
`FacilityEnding`（原本叫 `FacilityUnknown`）走進去直接進結局。

清算照原版依地圖判定（`Scene.collectToll`），旗標寫進角色記錄
（`Mission`／`Praised`，`StoreTo` **只動 bit0**，其餘七位未解不碰）。

⚠ **BIOS 一個 tick 是 1/18.2 秒，Ebiten 一幀是 1/60。** 每幀推一個 tick
會讓結局用三倍速播完——而畫面看起來只是「動畫比較快」，不像壞掉。
這裡每 3 幀推一個 tick。

**兩處與原版不同，都是簡化不是未解**：

| 原版 | 這一版 |
|---|---|
| 清算逐人停 105 tick，每個死者載一次肖像（圖 52） | 一次算完，名單走 `Killed()` |
| 刪掉一組之後把後面的組往前搬（15 bytes 一組） | 直接把那一組的槽表歸零，不搬 |

## 5. `+0x4B`／`+0x4C`：任務旗標與表揚旗標

`docs/re/91` §2.1 把這兩個 byte 記成未解。兩邊的用法擺在一起就看得出來：

| 欄位 | 誰設 | 誰讀 | 是什麼 |
|---|---|---|---|
| `+0x4B` bit0 | **結局**（`0x1B4FB`，逐人） | Radio 第一輪（`0x1B8D4`） | **參與過摧毀 Base Cochise** |
| `+0x4C` bit0 | Radio 第一輪 | Radio 第一輪 | **總部已經表揚過**（所以只聽得到一次）|

決定性的證據是 Radio 第一輪印的那條字串。`sub_1BB5D` 印之前會
`ds:4692h ← 0D622h`——**換成階級表**（`ExeStrings()` 的第 5 張），
第 `0x3D` 條是：

```
"Congratulations Rangers on a mission well done."
"Since you embarked on your mission we had reports from other rangers
 about the horrible strength contained in Base Cochise." …
```

⚠ **拿第 1 張表去查 `0x3D` 會取到 `"That doesn't seem to work."`**——
那句話出現在無線電畫面上完全說得通，而它會讓整段結論反過來。
`sub_1B7B7` 換 `ds:46B0h`、`sub_1BB5D` 換 `ds:4692h`，**兩支換的是不同的全域**；
看到「印字串 N」先確認那一刻的表基址是哪一個。

推論等級：**已確認**（兩處的讀寫點與字串內容都讀到）。

remake 因此把 Radio 的兩輪都接上了（`internal/play/command.go`）：
第一輪表揚、第二輪升級。先前只做第二輪，理由就是這兩個 byte 未解。

## 6. 三支輔助函式與那張肖像

| 位址 | 是什麼 |
|---|---|
| `sub_1142B` | `sub_1CBD3(ax ＝ 7)` ＝ **播音效 7**（`docs/re/44`）|
| `sub_162C7` | `ds:46FEh ← 0` 之後 `jmp sub_17A6B` ＝ 清訊息視窗 |
| `sub_1785E` | 一行：`ds:B265h ← 9E53h`，20 個呼叫端共用的輸出鉤子 |
| 圖 `0x34`（52） | **巡守員肖像**：紅底、戴帽、扛槍的人形，96 × 84。清算時每個死者載一次 |

## 7. 後日談那四條沒有人印

結局表（`ds:D18Eh`）第 6–9 條是 *The History of the Rangers, Vol. II*
的獻詞——「高踞鄰山，巡守員看著烈火吞沒 Base Cochise」那一段。
**遊戲裡走不到它們**：

- `sub_1B7B7` 在結局流程裡只被叫了 1–5（四段敘述 ＋ 死訊）。
- `ds:D18Eh` 這個立即數在全檔只出現一次，在 `0x1B7C0`——
  而那一行屬於 **`0x1B7BA` 這個沒有人呼叫的區塊**：

```
0x1B7B7  e9 0f 00   jmp sub_1B7C9      ; ← **跳過下面這一段**
0x1B7BA  50         push ax            ; 沒有呼叫端；0xB7BA 這個 word
0x1B7BB  e8 …       call sub_1785E     ;   在全檔一次都沒出現
0x1B7BF  b9 8e d1   mov cx, 0D18Eh     ; 換 ds:4692h（另一個表基址全域）
0x1B7C2  89 0e 92 46
0x1B7C6  e9 …       jmp sub_178A3
0x1B7C9  50         push ax            ; ← 活著的那一支（換 ds:46B0h）
```

`sub_1B7B7` 的 `jmp` 正好跨過整段——**兩支功能相同、只差換哪個全域的印字函式，
新的那支寫在舊的後面，入口用一個 jmp 跳過去**。這是改版留下的形狀。

那四條文字也不在 `paragraphs.txt` 或 `manual.txt` 裡（都 grep 過）。
**它是只存在於執行檔、玩不到的文字**——保存價值高於實作價值。

**重製版把它收進遊戲內手札**（使用者定案 2026-08-16）：手札在段落書的
162 段之後多四頁，自成一區 `SectionEpilogue`。
⚠ **`ParagraphCount` 維持 162**——那是紙本那本書的頁數，是一手事實；
後日談不在那本書上，混進 1–162 會讓那個數字失真。
正文走一般的執行檔字串翻譯（`exe:4:6`–`9`，四條都已經翻好），
查不到中文就顯示英文原文——**這一區不會有空白頁**，原文一定在執行檔裡拿得到。

## 8. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/endflow.json 0x1B7FE --callers
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/endanim.json 0x1B735
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/slot25.json 0x10FA7 0x10FD3
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/endtrig.json 0x1B4F0

# 觸發點：0xB4F0 在全檔只出現一次
python3 -c "
d=open('workplace/analysis/unpacked/wl.merged.exe','rb').read()
i=-1
while True:
    i=d.find(b'\xf0\xb4',i+1)
    if i<0: break
    print(hex(i+0xFF50))"

tools/go.sh test ./internal/assets/ -run TestEndAnimation -v
tools/go.sh test ./internal/play/ -run 'TestEnding|TestFacilityFourIsTheEnding' -v
tools/go.sh run ./cmd/wl-shot -mode pic -pic 52 -out workplace/shots/pic52.png

# 後日談的印字函式沒有人呼叫：位址與立即數兩種形式都掃過
python3 -c "
d=open('workplace/analysis/unpacked/wl.merged.exe','rb').read()
for w in (b'\xba\xb7', b'\x8e\xd1'):
    i=-1
    while True:
        i=d.find(w,i+1)
        if i<0: break
        print(w.hex(), hex(i+0xFF50))"
```

## 9. 這一輪學到的（寫成規則）

- **「零呼叫端」有第三種可能：它在表裡。** 先掃 `E8`／`E9`／`EB` 全空、
  正對照又證明掃描沒問題，剩下的就是**位址以資料的形式存在**。
  這時候要掃的是 `word` 本身，不是指令。
- **`Unknown` 這個名字會凍結進度。** 第 4 種設施叫 `FacilityUnknown`、
  圖片編號 `-1`、測試寫著「這一種原版沒有載圖，不要猜一個」——
  三處都正確、都誠實，而且**看起來像已經處理過了**，於是沒有人回頭問它是什麼。
  它是整個遊戲的終點。**把未解的東西命名成「未解」，會讓它從缺口變成一個條目。**
- **兩支「印字串」的函式可能換的是不同的表基址。** `sub_1B7B7` 換
  `ds:46B0h`、`sub_1BB5D` 換 `ds:4692h`，而拿錯表查出來的句子
  （`"That doesn't seem to work."`）在畫面上完全說得通。
  **看到「印字串 N」先問「這一刻的表基址是哪一個」**，不要預設是第 1 張。
- **一個欄位的語意，常常要兩個使用點才看得出來。** `+0x4B` 單看結局只知道
  「結局會設它」，單看 Radio 只知道「它擋著一段賀詞」；
  兩邊擺在一起才是「參與過摧毀 Base Cochise」。
  **未解的欄位要找齊讀寫點再下結論，不要在第一個使用點就命名。**
- **一個 `jmp` 跳過一整段，多半是改版留下的。** `0x1B7B7` 跳過
  `0x1B7BA`–`0x1B7C8` 直接到 `0x1B7C9`，而被跳過的那段是同一支函式的舊版本。
  **看到「跳過去的距離剛好等於下一個區塊」，先假設那是死碼再驗**——
  驗法是掃它的位址有沒有以 word 的形式出現在任何地方。
- **版面對不對，用等式驗不要用抽樣。** 15 格 × 4 ＋ 2,433 × 6 ＋ 2 剛好等於
  14,660，這比「前幾筆看起來合理」強得多——欄位少讀一個就對不齊。

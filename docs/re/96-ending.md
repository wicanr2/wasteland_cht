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
逐人：角色記錄 +0x4B 的 bit0 ← 1          ; 語意未解，見 §5
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

同一張表的第 5 條是 `"\v was killed in the blast.\r\r\r"`——**結局會點名死者**，
`sub_1B735` 在 `0x1B5C3` 的第五個呼叫端就在那條路上。

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

⚠ **BIOS 一個 tick 是 1/18.2 秒，Ebiten 一幀是 1/60。** 每幀推一個 tick
會讓結局用三倍速播完——而畫面看起來只是「動畫比較快」，不像壞掉。
這裡每 3 幀推一個 tick。

## 5. 還沒解的

- **角色記錄 `+0x4B` 的 bit0**：結局逐人把它設起來，Radio 的第一輪讀它
  （`docs/re/91` §2.1）。兩處都碰到同一個 byte，語意仍未解——
  角色結構沒有對應欄位，remake 兩邊都不做。
  下一個入口：全檔掃 `+0x4B` 的其他讀寫點。
- `0x1B571` 之後那一大段（`ds:4654h` 換組、隊伍槽表 `+0x0A` 比 `0x10`／`0x15`）
  ——結局似乎會依隊伍所在位置分支，還沒讀。
- `sub_1142B`、`sub_162C7`、`sub_1785E` 三支輔助函式。

## 6. 可重跑的完整指令

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
```

## 7. 這一輪學到的（寫成規則）

- **「零呼叫端」有第三種可能：它在表裡。** 先掃 `E8`／`E9`／`EB` 全空、
  正對照又證明掃描沒問題，剩下的就是**位址以資料的形式存在**。
  這時候要掃的是 `word` 本身，不是指令。
- **`Unknown` 這個名字會凍結進度。** 第 4 種設施叫 `FacilityUnknown`、
  圖片編號 `-1`、測試寫著「這一種原版沒有載圖，不要猜一個」——
  三處都正確、都誠實，而且**看起來像已經處理過了**，於是沒有人回頭問它是什麼。
  它是整個遊戲的終點。**把未解的東西命名成「未解」，會讓它從缺口變成一個條目。**
- **版面對不對，用等式驗不要用抽樣。** 15 格 × 4 ＋ 2,433 × 6 ＋ 2 剛好等於
  14,660，這比「前幾筆看起來合理」強得多——欄位少讀一個就對不齊。

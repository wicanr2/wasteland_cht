# 112：游標圖形對應哪個狀態 —— `ds:8DCDh` 是索引，`0x10D4D` 是繪製常式

日期：2026-08-18 ｜ 接 `docs/re/57`（`CURS` 的版面）、`docs/re/43`（滑鼠熱區）、
`docs/re/56` §4（`int 33h` 初始化）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`curs` 的 SHA-256 見 `docs/re/01`。

`docs/re/57` 解出 `CURS` 是 8 個 32 × 16 的圖形（左半遮罩、右半圖形），
但「哪個圖形對應哪個狀態」記成未解，理由是**找不到讀 `CURS` 的程式碼**。
讀它的程式碼在 `0x10D4D`；找不到是因為掃描的方式漏掉了負位移那一種寫法。

---

## 1. 為什麼上一輪掃不到

`CURS` 載到 `seg002:0x7E0B`（`docs/re/03` §5）。拿字串 `7E0B` 去 grep 反組譯，
會漏掉 IDA 把同一個位址寫成**負位移**的那些指令——`ds:A4E0h` 就是這樣，
全檔唯一的引用長成 `mov cx, [bx-5B20h]`。

正確的做法是把反組譯裡每一個十六進位字面值**帶正負號地正規化成 16 位元**
再比範圍：

```python
lit = re.compile(r'([+-])?\b([0-9A-F][0-9A-Fa-f]*)h')
def vals(s):
    out = []
    for sign, h in lit.findall(s):
        v = int(h, 16)
        if sign == '-':
            v = (-v) & 0xFFFF
        out.append(v & 0xFFFF)
    return out
```

正對照用 `ds:A4E0h`：正規化前 0 命中、正規化後 1 命中（`0x12C8C`）。
**正對照失敗的那一次才是真正的產出**——它證明的不是「A4E0 沒人用」，
是「這支掃描器有洞」。

`[0x7E0B, 0x860B)` 正規化後 8 命中，扣掉六筆「被當成立即數存進
`ds:7DF3h` 的值」（`0x8404`／`0x8434`，那是值不是位址），剩兩筆：

| 位址 | 指令 | 角色 |
|---|---|---|
| `0x111DD` | `mov dx, 7E0Bh` | 開機載入（`docs/re/03` §5）|
| `0x10D86` | `add bx, 7E0Bh` | **繪製游標** |

推論等級：**已確認**。

## 2. `0x10D4D`：一支標準的遮罩疊圖游標

```
0x10D4D  ax ← ds:7DF9h；ax >>= 3            ; 游標 x（像素）→ byte 欄
0x10D58  ax ← ds:7DFBh；ax <<= 3            ; y × 8
0x10D61  si += ax；ax <<= 2；si += ax        ; ＋ y × 32 ＝ y × 40（每列 40 bytes）
0x10D69  ds:7D31h ← si                      ; 螢幕位移
0x10D72  out 3CEh/4、out 3C4h/2              ; EGA 讀寫平面
0x10D7E  bx ← ds:8DCDh
0x10D82  bh ← bl；bl ← 0                    ; **bx ＝ 索引 × 256**
0x10D86  bx += 7E0Bh                        ; ← CURS ＋ 索引 × 256
0x10D8A  cl ← ds:7DF9h & 7                  ; 位元內的位移量
loc_10DAE:
  ax ← [si]                                 ; 背景
  dx ← es:[bx]；bx += 2；dx >>= cl；dx ← ~dx
  ax &= dx                                  ; 先照遮罩挖洞
  dx ← es:[bx]；dx >>= cl；ax |= dx          ; 再疊圖形
  [si] ← ax
```

一個圖形 256 bytes，所以 `索引 × 256` 就是第幾個圖形——**`ds:8DCDh` 是游標索引**。
遮罩與圖形交錯讀（先遮罩後圖形），與 `docs/re/57` §1 的「左 16 遮罩、右 16 圖形」
是同一件事的兩種看法：一列 4 個 word，前兩個是遮罩、後兩個是圖形。

推論等級：**已確認**。

## 3. 誰寫 `ds:8DCDh`

全檔 11 個引用（同一支正規化掃描），一個讀（上面那個）、十個寫：

| 位址 | 值 | 情境 |
|---|---:|---|
| `0x174BE` | 1 | `sub_174B7`：開始掃熱區標籤表 |
| `0x1750D` | 0 | 同一支：掃完沒命中 |
| `0x1891D` | 0 | 放開按鍵後回到預設 |
| `0x18BFF` | 1 | 清單列的熱區命中 |
| `0x18C2E` | 0 | 沒命中 |
| `0x18C4C` | 1 | **指到名片列**（`dx ≥ 0x7D`，列號 ≤ `ds:4653h` ＝ 隊伍人數），點下去送 `'0' + 列號` |
| `0x18C80` | 6 | **指到地圖正中央那 16 × 16 的方框**（x ∈ [0x88, 0x98)、y ∈ [0x38, 0x48)），點下去送 `0x1B`（ESC）|
| `0x18CC3` | 2–5 | **地圖視窗的四個楔形**，見 §4 |
| `0x18CE7` | 0 | 回到預設 |
| `0x18D04` | 1 | 另一張列表（`ds:8DDCh`）命中 |

**沒有任何一處寫 7。** 第 8 個圖形（斜紋填滿的塊）在這份 DOS 版裡選不到。

推論等級：**已確認**（十個寫入點逐一讀過）；
「7 號完全沒用到」是**強證據**——掃描比對的是位址字面值，
若有人先把 `8DCDh` 載進暫存器再間接寫就掃不到，全檔沒有這種形式。

## 4. 2–5 是四個方向，而且方向鍵就是 `I J K L`

```
0x18C93  ax ← dx + 50h；cx < ax → loc_18CB1
0x18C9C  ax ← 0D0h − dx；cx < ax → bx ← 2，否則 bx ← 5
loc_18CB1:
0x18CB1  ax ← 0E0h − dx；cx < ax → bx ← 4，否則 bx ← 3
0x18CC3  ds:8DCDh ← bx
0x18CCF  bx −= 2；al ← [bx-3FA3h]           ; ← ds:C05Dh[方向]
```

兩條判準線是 `x − y = 0x50` 與 `x + y = 0xD0`，交點 `(144, 64)`
——**正好是 §3 那個中央方框的中心**。兩條 45° 線把地圖視窗切成四個楔形：

| `bx` | 楔形 | `ds:C05Dh` 送出的鍵 |
|---:|---|---|
| 2 | 上 | `0x49` ＝ `I` |
| 3 | 下 | `0x4B` ＝ `K` |
| 4 | 左 | `0x4A` ＝ `J` |
| 5 | 右 | `0x4C` ＝ `L` |

```bash
python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xC05D 4 --bytes
# ds:C05Dh  49 4b 4a 4c
```

四條線各自獨立而且互相對上：

1. **幾何**：楔形的方位（上／下／左／右）由兩條判準線算出來。
2. **鍵表**：`ds:C05Dh` 的四個 byte 是 `I K J L`。
3. **另一份鍵表**：`docs/re/72` §4 早就從地圖指令列的字串
   （`"…Radio\0IKJL…"`）讀出「I 上、K 下、J 左、L 右」。
   這裡是**第二個獨立來源**，兩份的順序完全一致。
4. **圖形**：`docs/re/57` §3 讀出 2／3 是上下雙箭頭、4／5 是左右鏡像；
   `generated/ida94/curs.md` 的遮罩形狀確認 2 的尖端在頂列、3 在底列、
   4 的尖端在左、5 的尖端在右。

`docs/re/57` §3 把 0／1、2／3、4／5 記成「同形狀的兩個變體，像動畫的兩幀」。
**那三對不是動畫**：0／1 是「一般／指到可點的東西」，2／3 是上／下，4／5 是左／右。

推論等級：**已確認**。

## 5. 所以八個圖形的狀態表

| # | 形狀 | 什麼時候用 |
|---:|---|---|
| 0 | 實心橫塊（紅）| 預設；熱區沒命中 |
| 1 | 實心橫塊（亮紅，中間一道缺口）| 指到可點的字：指令列標籤、清單列、名片列 |
| 2 | 上箭頭 | 地圖視窗的上楔形 → `I` |
| 3 | 下箭頭 | 地圖視窗的下楔形 → `K` |
| 4 | 左箭頭 | 地圖視窗的左楔形 → `J` |
| 5 | 右箭頭 | 地圖視窗的右楔形 → `L` |
| 6 | 十字 | 地圖正中央 16 × 16 的方框 → ESC |
| 7 | 斜紋塊 | **選不到**（沒有寫入點）|

## 6. 對 remake 的意思

`internal/play/play.go` 的 `drawCursor` 原本固定用第 0 個，理由是
「哪個圖形對應哪個狀態沒有解」。解出來之後那個理由不成立了。

重製版的熱區與原版不同（`docs/spec/29`：畫面已經是 960 × 600，訊息視窗重排過），
所以接法是**把狀態對應套在重製版自己的熱區上**，不是照抄座標：

| 重製版的情況 | 圖形 |
|---|---|
| 指標下有可點的 ASCII 字（`charAt`）| 1 |
| 在地圖視窗且 `viewDirection` 回一個方向 | 2／3／4／5 |
| 在地圖視窗但點在隊伍自己那一格 | 6 |
| 其他 | 0 |

`I`／`J`／`K`／`L` 這一組同時解釋了另一件事：清單框架用 `I`／`K` 翻頁
（`docs/re/53` §3）不是隨便選的鍵，**它們就是原版的上下方向鍵**
（`docs/re/72` §4）。`USE` 的清單因此改成 `I`／`K` 翻頁
（`internal/play/use.go`，`docs/re/92` §3.2）。

## 7. 可重跑的完整指令

```bash
# 正規化過的位址掃描（正對照 ds:A4E0h 要有 1 命中）
docker run --rm --network none --log-opt max-size=10m --log-opt max-file=3 \
  -v "$PWD":/w -w /w -u "$(id -u):$(id -g)" python:3.12-slim \
  python3 tools/scan_addr_refs.py workplace/analysis/dumps/listing.json 0x7E0B 0x860B
docker run --rm --network none --log-opt max-size=10m --log-opt max-file=3 \
  -v "$PWD":/w -w /w -u "$(id -u):$(id -g)" python:3.12-slim \
  python3 tools/scan_addr_refs.py workplace/analysis/dumps/listing.json 0x8DCD 0x8DCF

python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xC05D 4 --bytes
python3 tools/dump_curs.py workplace/orig/wastland/curs docs/re/generated/ida94/curs.md
```

## 8. 這一輪學到的（寫成規則）

- **「找不到讀它的程式碼」要先驗掃描器，再下結論。** 拿一個**已知會命中**的
  位址跑同一支掃描器；正對照回零就代表掃描器有洞，而不是目標沒人用。
  這一格卡了三份筆記（`57`、`24`、`00-remake-knowledge-gaps`），
  而洞只是「IDA 把位址寫成負位移」。
- **16 位元的位址在反組譯裡有兩種寫法。** `[bx+7E0Bh]` 與 `[bx-81F5h]`
  是同一個位址。比對位址一律先 `& 0xFFFF` 正規化——這與
  `get_operand_value()` 的符號擴展（`docs/re/03` §7）是同一個坑的兩種長相。
- **成對的美術不一定是動畫。** 「兩個圖形長得像，所以是兩幀」是看圖說故事；
  真正的答案在寫索引的那十行程式碼裡。**先問誰寫這個索引**，
  再回頭看美術對不對得上。

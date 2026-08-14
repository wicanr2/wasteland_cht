# 04：`wla.bin` overlay 與它提供的繪圖層

日期：2026-08-13 ｜ 輸入：`wl.merged.exe`（本專案合成）
SHA-256 `cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`

## 1. 為什麼要另外做一份映像

`start` 開機時把 `wla.bin`（4,209 bytes）讀進 `CS:0000`（`docs/re/03` §5）。
直接分析 `wl.unpacked.exe` 的那 8 KB 看到的是**還沒被覆蓋的內容**，
分析它等於分析一段永遠不會執行的程式碼。

`tools/apply_overlay.py` 把 overlay 疊上去，產生 `wl.merged.exe`。
位址體系與解包版完全相同（載入基址 `0x10000`），差別只在 `0x10000`–`0x11071` 這一段。

| 映像 | SHA-256 | 用途 |
|---|---|---|
| `wl.unpacked.exe` | `b5eb39f0…31a0` | 主程式（`0x11071` 之後） |
| `wl.merged.exe` | `cd5b07ea…8118` | overlay 區（`0x10000`–`0x11071`）**本專案合成，非原版檔案** |

IDA 自動辨識函式數：解包版 614 → 合成版 641。

## 2. 覆蓋前的那 8 KB 是一張安全網

被覆蓋的區域原本只有 **78 個非零 bytes**，其餘全是 0。那 78 bytes 是：

```
0x10000:  e9 b3 10   jmp 0x110B6   ; → start
0x10003:  e9 b0 10   jmp 0x110B6   ; → start
0x10006:  e9 ad 10   jmp 0x110B6   ; → start
…（26 個，全部指向 start）
```

**26 個 3-byte `jmp` 的 thunk 表**，未載入 overlay 時全部導回 `start`。
`wla.bin` 的開頭是同樣格式的 26 個 `jmp`，但指向 overlay 內部的真實實作：

```
0x10000:  e9 4b 00   jmp 0x1004E
0x10003:  e9 82 00   jmp 0x10088
0x10006:  e9 3b 01   jmp 0x10144
…
```

所以**主程式呼叫 overlay 一律經過這 26 個固定入口**，overlay 內部佈局可以自由改動。
`start` 最後那句 `call j_start`（`0x1127B`）算出來就是 `0x10000` ＝ slot 0。

## 3. Overlay API 表

`callers` 是合成映像裡對該 slot 的直接呼叫數，可當作「這支有多核心」的粗略排序。

| slot | thunk | 實作 | callers | 已解語意 |
|---:|---|---|---:|---|
| 0 | `0x10000` | `0x1004E` | 2 | 設定視訊模式並建列位址表（§4） |
| 1 | `0x10003` | `0x10088` | 1 | 圖磚 packed 4bpp → EGA 4 平面（`docs/re/24` §3.1） |
| 2 | `0x10006` | `0x10144` | 1 | `ALLPICS` 圖片的列間 delta 解碼（`docs/re/23` §2） |
| 3 | `0x10009` | `0x10156` | 1 | 未解 |
| 4 | `0x1000C` | `0x1029B` | 4 | **畫一格地圖**：背景圖形 AND 遮罩 OR 疊圖（`docs/re/24` §2.3） |
| 5 | `0x1000F` | `0x1060C` | 3 | 未解 |
| 6 | `0x10012` | — | 0 | IDA 未建函式，無直接呼叫端 |
| 7 | `0x10015` | — | 1 | 目標未解析 |
| 8 | `0x10018` | — | 0 | IDA 未建函式，無直接呼叫端 |
| 9 | `0x1001B` | `0x10762` | 10 | 清除矩形區域（§6） |
| 10 | `0x1001E` | `0x107D8` | 1 | 未解 |
| 11–14 | `0x10021`–`0x1002A` | `0x108A3`–`0x109BC` | 各 1 | 未解（四支大小相近，疑似同族） |
| 15 | `0x1002D` | `0x10A1A` | 1 | 未解 |
| 16 | `0x10030` | `0x10A7A` | 1 | 拆 `ALLPICS` 的參數區成兩張指標表（內容未解，`docs/re/23` §4） |
| 17 | `0x10033` | `0x10B11` | 1 | 未解 |
| 18 | `0x10036` | `0x10C5A` | 1 | 未解 |
| **19** | `0x10039` | `0x10CB6` | **71** | **繪製一個字元**（§5） |
| 20 | `0x1003C` | `0x10D4D` | 1 | 未解 |
| 21 | `0x1003F` | `0x10E02` | 5 | 未解 |
| 22 | `0x10042` | `0x10E45` | 2 | 未解 |
| 23 | `0x10045` | `0x10F12` | 2 | 把 `seg003:0x920` 的 `0x24` × `0x80` 區塊搬上螢幕位移 `0x141` |
| 24 | `0x10048` | `0x10F64` | 1 | 未解 |
| 25 | `0x1004B` | `0x10FA7` | 1 | 未解 |

⚠ 「callers＝1」多半代表**呼叫端是同一支包裝函式**，不代表這支不重要；
slot 19 的 71 個呼叫端則遍布 `sub_17E42`（18 次）、`sub_189B1`（8 次）、`sub_16F70`（7 次）等。

## 4. slot 0：視訊模式與列位址表

```asm
0x1004E  or   al, al
0x10050  jz   short loc_1005A
0x10052  mov  ax, 0Dh
0x10055  int  10h            ; EGA 320×200 16 色
0x1005A  mov  ax, 3
0x1005D  int  10h            ; 文字模式（AL=0 時）
```

`start` 用 `mov al, 1` 呼叫，所以遊戲跑在 **EGA mode 0Dh（320×200、16 色、4 平面）**。

接著就地產生一張表：

```asm
0x10064  mov  di, 8DF9h
0x10067  mov  ax, 0
0x1006A  mov  cx, 0C8h       ; 200 列
0x1006D  stosw
0x1006E  add  ax, 28h        ; 每列 40 bytes
0x10071  loop loc_1006D
```

`ds:8DF9h` 起是 **200 個 word 的列起始位移表**（`y × 40`）。
320 pixels ÷ 8 bits ＝ 40 bytes／列，與 EGA planar 佈局一致。
之後再整張複製到 `seg003:0xDFE0` 備一份。

## 5. slot 19：畫一個字元（`0x10CB6`）

overlay 裡呼叫最多的一支，也是把 `COLORF.FNT` 的用途釘死的一支。

```asm
0x10CE5  mov  bl, ds:8DD1h     ; 游標列
0x10CE9  shl  bx, 1 ×3         ; ×8
0x10CEF  mov  dl, ds:8DD0h     ; 游標行
0x10CF3  inc  byte ptr ds:8DD0h ; 印完自動右移一格
0x10CF7  shl  ax, 1 ×5         ; 字元碼 × 32
0x10D01  add  ax, 0B4E0h       ; ＋ 字型基址
0x10D04  mov  si, ax
0x10D06  shl  bx, 1            ; 列 × 16
0x10D08  add  dx, [bx-7207h]   ; ＋ 列位址表（16-bit wrap：0x10000-0x7207 = 0x8DF9）
0x10D0E  mov  ax, 0A000h
0x10D11  mov  es, ax           ; EGA 視訊記憶體
0x10D19  mov  dx, 3C4h         ; sequencer address
0x10D1E  out  dx, al
```

三件事因此確定：

1. **字型基址 `0xB4E0` 就是 `COLORF.FNT` 的載入位址**（`docs/re/03` §5 的表）。
   兩份證據獨立得到同一個數字。
2. **每個字元 32 bytes**（`× 32` 來自五次 `shl`）。
   `colorf.fnt` 5,504 ÷ 32 ＝ **172 個字元**，整除。
3. 字元格是 **8×8 pixel**：列位址用 `y × 16`，而列位址表是每像素列 2 bytes，
   `16 ÷ 2 ＝ 8` 像素列；32 bytes ÷ 8 列 ＝ 4 bytes／列 ＝ EGA 的 4 個平面各 1 byte。

游標位置在 `ds:8DD0h`（行）與 `ds:8DD1h`（列），印一個字元後行自動 +1。

### 一個尚未解釋的字元碼重映射

```asm
0x10CD4  cmp  al, 18h
0x10CD6  jb   short loc_10CE5
0x10CD8  cmp  al, 33h
0x10CDA  ja   short loc_10CE5
0x10CDC  cmp  byte ptr ds:722Fh, 0
0x10CE1  jnz  short loc_10CE5
0x10CE3  add  al, 1Ch          ; 字元碼 0x18–0x33 且旗標為 0 → ＋0x1C
```

字元碼 `0x18`–`0x33` 這 28 個在 `ds:722Fh` 為 0 時會被平移 `0x1C`。
**用途未解**，但這是中文化時必須先弄清楚的東西——字型表不是單純的線性索引。

## 6. slot 9：清除矩形（`0x10762`）

用 `ds:4674h`–`ds:4677h` 四個 byte 當矩形邊界，
把 EGA sequencer 的 map mask 設成 `0x0F`（四個平面一起寫）後填 0，
最後把矩形參數收進 `ds:4672h`／`ds:4673h`。

繪圖層的座標與矩形狀態集中在 `ds:4672h`–`ds:4678h` 這一小段。

## 7. 已確定的位址對照（累積）

| 位址 | 內容 | 來源 |
|---|---|---|
| `seg003:0x0100` | `TRANSTBL`（800 bytes） | `docs/re/03` §5 |
| `seg003:0x0420` | `IC0_9.WLF`（1,280） | `docs/re/03` §5 |
| `seg003:0x0920` | `TITLE.PIC`（18,432，XOR 解碼後） | `docs/re/03` §5–6 |
| `seg003:0xB4E0` | `COLORF.FNT`（5,504 ＝ 172 字 × 32） | `docs/re/03` §5＋本篇 §5 |
| `seg003:0xDA60` | `MASKS.WLF`（320） | `docs/re/03` §5 |
| `seg003:0xDFE0` | 列位址表副本（200 words） | 本篇 §4 |
| `seg002:0x7E0B` | `CURS`（2,016 讀入） | `docs/re/03` §5 |
| `seg002:0x8DD0`／`0x8DD1` | 文字游標 行／列 | 本篇 §5 |
| `seg002:0x8DF9` | 列位址表（200 words） | 本篇 §4 |
| `seg002:0x4672`–`0x4678` | 繪圖矩形參數 | 本篇 §6 |
| `seg002:0x722F` | 字元碼重映射的開關 | 本篇 §5 |
| `seg002:0x9174` | 目前檔案 handle | `docs/re/03` §2 |

## 8. 下一步

1. 逐一解 overlay 其餘的未解 slot（`0x1029B` 已解，見 `docs/re/24` §2.3）。
2. 追 slot 19 的 71 個呼叫端，找出字串輸出層（那裡會接上文字編碼）。
3. `ds:722Fh` 的字元碼重映射是什麼——中文化前必須先解。
4. `GAME1`／`GAME2` 的載入點仍未追（`docs/re/03` §9）。

## 9. 重跑方式

```sh
docker run --rm --network none --memory 1g --cpus 1 --pids-limit 256 \
  --user "$(id -u):$(id -g)" -v "$PWD:/workspace" -w /workspace \
  ida-pro-9.4-idapython:py312-v1 /opt/venv/bin/python3 tools/apply_overlay.py \
    workplace/analysis/unpacked/wl.unpacked.exe workplace/orig/wastland/wla.bin \
    workplace/analysis/unpacked/wl.merged.exe workplace/analysis/unpacked/overlay-report.json

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" tools/ida.sh build
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
    workplace/analysis/dumps/overlay-thunks.json \
    $(python3 -c "print(' '.join(f'0x{0x10000+i*3:X}' for i in range(26)))")
```

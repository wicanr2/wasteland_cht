# 03：開機序列、檔名表與資產載入

日期：2026-08-13 ｜ 輸入：`wl.unpacked.exe`
SHA-256 `b5eb39f094e0274165eab5e1584e78ff5b54c7228d8db273573d2bd951ea31a0`

分析對象是 `start`（`0x110B6`–`0x1127E`，456 bytes、181 條指令）與它呼叫的三個檔案 I/O 包裝。
位址一律是 IDA 線性位址；`ds:XXXX` 是 `seg002` 相對位移，換算 **線性 ＝ `0x1CE20` ＋ 位移**。

## 1. 資料段與段基址

`start` 開頭就把 DS 換成 `seg002`、把原本的 DS（PSP）存進 `ds:9172h`：

```asm
0x110B6  mov  bx, ds
0x110B8  mov  ax, seg seg002      ; DS ← 資料段
0x110BB  mov  ds, ax
0x110BD  mov  word_25F92, bx      ; ds:9172h ＝ PSP segment
0x110C1  mov  ax, seg seg003      ; ds:9165h ＝ 第二資料段（素材緩衝）
0x110C4  mov  seg_25F85, ax
```

`seg002`（`0x1CE20`）是主資料段，`seg003`（`0x2AE20`）是素材緩衝段。
兩個段基址一路被 `push ds` / `mov ds, seg seg003` 切換使用。

## 2. 檔案 I/O 的三個包裝函式

| 函式 | 服務 | 介面 |
|---|---|---|
| `sub_11384` | `AH=3Dh` open（`AL=2` 讀寫） | 入：`DS:DX` ＝ 檔名。出：`AX` ＝ handle，同時寫進 `ds:9174h` |
| `sub_113B2` | `AH=3Fh` read | 入：`BX` ＝ handle、`CX` ＝ 長度、`DS:DX` ＝ 緩衝 |
| `sub_113A9` | `AH=3Eh` close | 入：`BX` ＝ handle |

open 與 read **失敗時不回報錯誤，而是原地重試**：

```asm
0x11394  push bx / push cx / push dx
0x1139B  mov  word ptr ds:4680h, 9274h   ; 訊息指標 → ds:9274h
0x113A1  call sub_1786E                  ; 顯示訊息並等待
0x113A7  jmp  short sub_11384            ; 重新開檔
```

`ds:9274h` ＝ `0x26094`，指向 `A disk error has occurred. Please check the disks.`。
這條迴圈就是原版「請插入磁片」的機制：換片之後重試，成功才往下走。

## 3. `info`：硬碟安裝資訊，兩個 byte

`start` 做的第一件檔案操作不是載入素材，而是讀一個兩 bytes 的檔：

```asm
0x110E9  mov  ah, 19h
0x110EB  int  21h                  ; 取預設磁碟機 → ds:92F4h
0x110F0  mov  dx, 9204h            ; 檔名 "info"（0x26024）
0x110F5  mov  ah, 3Dh / int 21h
0x11100  mov  cx, 2                ; 只讀 2 bytes
0x11103  mov  dx, 92F1h            ; → 0x26111、0x26112
0x11106  mov  ah, 3Fh / int 21h
0x11113  cmp  byte_26111, 'A'      ; 第一個 byte ＝ 磁碟機代號
0x1111A  cmp  byte_26111, 'B'
0x11121  mov  byte_27235, 1        ; 'B' → 這個旗標
0x11129  mov  byte_27234, 1        ; 其他（硬碟）→ 這個旗標
0x1112E  mov  al, byte_26112       ; 第二個 byte → ds:9167h
```

散布版目錄裡的 `info` 檔正好是 2 bytes，內容 `C ` ——
第一個 byte 是安裝磁碟機代號、第二個 byte 目前用途未解（存進 `ds:9167h`）。
`'A'`／`'B'` 走磁片路徑，其他走硬碟路徑。

**這解釋了為什麼這份散布版能從硬碟執行**：`info` 是安裝程式留下的轉接資訊。

## 4. 檔名表（`0x25FAD`–`0x26028`）

13 個 NUL 結尾字串連續排列，長度不等，沒有指標表：

| 線性位址 | `ds:` 位移 | 檔名 | 開機時載入 |
|---|---|---|---|
| `0x25FAD` | `918Dh` | `ALLPICS1` | ✗ |
| `0x25FB6` | `9196h` | `ALLPICS2` | ✗ |
| `0x25FBF` | `919Fh` | `GAME1` | ✗ |
| `0x25FC5` | `91A5h` | `GAME2` | ✗ |
| `0x25FCB` | `91ABh` | `ALLHTDS1` | ✗ |
| `0x25FD4` | `91B4h` | `ALLHTDS2` | ✗ |
| `0x25FDD` | `91BDh` | `END.CPA` | ✗ |
| `0x25FE5` | `91C5h` | `WLA.BIN` | ✓ |
| `0x25FED` | `91CDh` | `COLORF.FNT` | ✓ |
| `0x25FF8` | `91D8h` | `CURS` | ✓ |
| `0x25FFD` | `91DDh` | `IC0_9.WLF` | ✓ |
| `0x26007` | `91E7h` | `MASKS.WLF` | ✓ |
| `0x26011` | `91F1h` | `TRANSTBL` | ✓ |
| `0x2601A` | `91FAh` | `TITLE.PIC` | ✓ |
| `0x26024` | `9204h` | `info` | ✓（見 §3） |

引用方式是把位移當立即數直接載入 DX（`mov dx, 91C5h`），
所以 IDA 的 xref 圖對這些字串是空的——它們的 0 xref 不代表沒人用（見 §7）。

開機**沒有**載入的六個檔（`GAME1`／`GAME2`／`ALLPICS*`／`ALLHTDS*`／`END.CPA`）
是遊戲執行中才讀的，載入點還沒追。

## 5. 開機載入表

每一項的形狀都一樣：`mov dx, <檔名位移>` → `sub_11384` open →
（必要時切 DS 到 `seg003`）→ `mov dx/cx` → `sub_113B2` read → `sub_113A9` close。

| 順序 | 檔名 | 目標 | 讀取長度 | 檔案實際大小 | 對照 |
|---|---|---|---:|---:|---|
| 1 | `WLA.BIN` | **`CS:0000`** | `0x2000` (8,192) | 4,209 | 讀到 EOF 為止 |
| 2 | `COLORF.FNT` | `seg003:0xB4E0` | `0x1580` (5,504) | 5,504 | 完全相符 |
| 3 | `TITLE.PIC` | `seg003:0x0920` | `0x4800` (18,432) | 18,432 | 完全相符 |
| 4 | `CURS` | `seg002:0x7E0B` | `0x07E0` (2,016) | 2,048 | 少讀 32 bytes |
| 5 | `IC0_9.WLF` | `seg003:0x0420` | `0x0500` (1,280) | 1,280 | 完全相符 |
| 6 | `MASKS.WLF` | `seg003:0xDA60` | `0x0500` (1,280) | 320 | 讀到 EOF 為止 |
| 7 | `TRANSTBL` | `seg003:0x0100` | `0x0320` (800) | 800 | 完全相符 |

四個長度與檔案大小完全相符，是「這張表讀對了」的算術佐證。

### `WLA.BIN` 是程式碼 overlay

第一項最特別——它讀進的是**程式碼段**：

```asm
0x11145  push ds
0x11146  mov  ax, cs
0x11148  mov  ds, ax          ; 緩衝區在 CS
0x1114A  mov  dx, 0           ; CS:0000
0x1114D  mov  cx, 2000h
0x11150  call sub_113B2
```

`CS` 基址是 `0x10000`，檔案 4,209 bytes，所以覆蓋 `0x10000`–`0x11071`，
而 `start` 本身從 `0x110B6` 開始——**剛好落在覆蓋範圍之外**。
`wl.exe` 開頭這 8 KB 是為 overlay 預留的空洞，`wla.bin` 是遊戲主程式的第一塊程式碼。

## 6. `TITLE.PIC` 的 XOR 自參考解碼（已確認）

讀進來之後就地解碼：

```asm
0x111B1  push ds
0x111B2  mov  ax, seg seg003
0x111B5  mov  ds, ax
0x111B7  mov  es, ax
0x111B9  mov  si, 920h        ; 影像起點
0x111BC  mov  di, si
0x111BE  add  di, 90h         ; di ＝ si + 0x90
0x111C2  mov  cx, 23B8h
0x111C5  lodsw                ; ax ← ds:[si]（已解碼的前段）
0x111C6  xor  ax, [di]        ; 與待解碼位置 XOR
0x111C8  stosw                ; 寫回 es:[di]
0x111C9  loop loc_111C5
```

`si` 永遠落後 `di` 0x90 bytes，而 `si` 讀到的是**已經解好的**內容，
所以這是以自身前 144 bytes 為滾動金鑰的串流解碼：

```
out[n] = in[n] XOR out[n - 0x90]
```

前 `0x90` bytes 不解碼，直接當種子。

算術驗證：`0x90` ＋ `0x23B8` × 2 ＝ 144 ＋ 18,288 ＝ **18,432**，
與 `title.pic` 檔案大小完全相同。整張圖一個 byte 不多不少。

## 7. ⚠ 工具坑：`get_operand_value` 會把 16-bit 立即數符號擴展

追這些檔名引用時，第一版的立即數掃描器回報「檔名表沒有任何引用」——
而實際上 `start` 裡到處都是 `mov dx, 91C5h`。

原因：`idc.get_operand_value()` 對 `0x91C5`（bit 15 ＝ 1）回傳的是符號擴展後的
`0xFFFFFFFFFFFF91C5`。拿它加段基址算出來的位址不存在，於是一筆都對不上。

**症狀是安靜的零命中，和「真的沒人引用」長得一模一樣。**
修法是 `& 0xFFFF`；`tools/ida/export_file_io.py` 已修正並在該處留了註解。
修正前後：立即數命中數 67 → 230，`seg002` 裡被引用的字串 0 → 50。

順帶一提，第一版把 ASCII 掃描的最短長度設成 4，`CURS` 與 `info` 這兩個
四字元檔名剛好卡在邊界——現在改成 2。

## 8. 絕對磁區讀寫（`int 25h`／`int 26h`）

全檔只有兩處中斷是磁碟存取，都不是 BIOS `int 13h`：

| 函式 | 中斷 | callers |
|---|---|---|
| `sub_116AC` | `int 25h` 絕對磁區讀 | `sub_115E5`、`sub_11730`（兩處）、`sub_118D2` |
| `sub_11854` | `int 26h` 絕對磁區寫 | `sub_11730` |

兩支的位址計算完全相同：

```asm
mov  al, ds:92E4h      ; cylinder
shl  ax, 1             ; × 2 面
mov  dl, ds:92E5h      ; head
add  ax, dx
mov  cx, 9
mul  cx                ; × 9 磁區／磁軌
mov  dl, ds:92E6h      ; sector
add  dx, ax            ; DX ＝ 邏輯磁區號
mov  al, ds:92E2h      ; 磁碟機號
mov  cl, ds:92EEh      ; 磁區數
mov  bx, ds:472Fh      ; 緩衝區 offset
int  25h
```

「2 面 × 9 磁區／磁軌」是 **360 KB 5.25 吋磁片**的幾何。
這條路徑存在且有呼叫端（不是死碼），但**是否會在硬碟安裝下執行還沒追**——
`info` 的磁碟機代號旗標（`ds:A414h`／`ds:A415h`）很可能就是分流點。
在追出來之前，不要斷言「硬碟版不走這條」。

## 9. 未解與下一步

| 項目 | 狀態 |
|---|---|
| `GAME1`／`GAME2`／`ALLPICS*`／`ALLHTDS*`／`END.CPA` 的載入點 | 未追 |
| `wla.bin` 的內容（overlay 程式碼） | 未分析。它載到 `CS:0000`，需要單獨建一份資料庫 |
| `info` 第二個 byte 的用途 | 未解（存進 `ds:9167h`） |
| `int 25h`／`26h` 路徑在硬碟安裝下是否執行 | 未追 |
| `sub_1786E`（錯誤訊息顯示） | 未分析 |
| `CURS` 為何只讀 2,016／2,048 | 未解 |

## 10. 重跑方式

```sh
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.unpacked.exe" \
  tools/ida.sh run tools/ida/export_file_io.py docs/re/generated/ida94/file-io.json

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.unpacked.exe" \
  tools/ida.sh run tools/ida/export_function.py \
    workplace/analysis/dumps/boot-and-fileio.json \
    0x110B6 0x11384 0x113A9 0x113B2 0x119A8 --callers

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.unpacked.exe" \
  tools/ida.sh run tools/ida/export_function.py \
    workplace/analysis/dumps/disk-io.json 0x116AC 0x11854 --callers
```

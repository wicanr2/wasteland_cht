# 06：`GAME1`／`GAME2` 的資源目錄，與文字輸出層

日期：2026-08-13 ｜ 輸入：`wl.unpacked.exe`（`b5eb39f0…31a0`）、`wl.merged.exe`（`cd5b07ea…8118`）

## 1. 資源定址的三張表

`GAME1`／`GAME2` 內部沒有自帶目錄——**目錄在執行檔的資料段裡**。
載入器 `sub_183B1`（110 bytes）與 `sub_1841F`（201 bytes）用的是同一套定址：

```asm
0x183BD  mov  bl, ds:46E0h        ; 資源編號
0x183C1  mov  al, [bx-4137h]      ; ← 目錄表 ds:BEC9h（16-bit wrap）
0x183C5  and  al, 0C0h            ; 高 2 bits ＝ 在哪一張磁碟
0x183C7  test ds:9168h, al        ; 目前插著的磁碟
0x183CD  xor  byte ptr ds:9168h, 0C0h   ; 不符就切換
0x183D2  mov  dl, 1
0x183D5  call sub_11445           ; 開資源 1（GAME1，切換後為 GAME2）
0x183DC  mov  bl, [bx-4137h]
0x183E0  and  bl, 3Fh             ; 低 6 bits ＝ 位移表索引
0x183E3  shl  bx, 1 ×2            ; ×4 ＝ 每筆 4 bytes
0x183E7  cmp  byte ptr ds:9168h, 80h
0x183EE  mov  dx, [bx-4336h]      ; 表 B（ds:BCCAh）低 word
0x183F2  mov  cx, [bx-4334h]      ;         高 word
0x183F9  mov  dx, [bx-4386h]      ; 表 A（ds:BC7Ah）低 word
0x183FD  mov  cx, [bx-4384h]      ;         高 word
0x18401  call sub_11534           ; seek 到該位移
```

| 表 | `ds:` 位移 | 線性 | 內容 |
|---|---|---|---|
| 資源目錄 | `BEC9h` | `0x28CE9` | 每個資源 1 byte：高 2 bits ＝ 磁碟、低 6 bits ＝ 位移表索引 |
| 位移表 A | `BC7Ah` | `0x28A9A` | 每筆 4 bytes ＝ 32-bit 檔內位移 |
| 位移表 B | `BCCAh` | `0x28AEA` | 同上，另一張磁碟用 |

### 位移表的實際數值

```
表 A：0x00000000  0x00002ACE  0x00004139  0x000066B9  0x00008E2A
      0x0000AE26  0x0000BFD8  0x0000E774  0x000103DA  0x00011EEF
      0x00014503  0x00016E0E  0x000188F6  0x00019D5F  …
表 B：0x00000000  0x000010E0  0x0000257E  0x000055C3  0x00007440
      0x0000895F  0x00009F94  0x0000B6C9  0x0000CA21  0x0000E296
      0x00010555  0x00011CE7  0x0001486E  0x00017042  …
```

嚴格遞增、高位 word 在跨 64 KB 時正確進位、最大值落在
`game1`（159,429 bytes）與 `game2`（172,235 bytes）之內。
**這三項一起構成「這是 32-bit 檔內位移」的算術證據。**

### 資源目錄的實際數值

```
80 81 82 83 84 85 86 40 87 88 89 41 42 43 0E 44 45 46 47 48 49 4A 4B 4C …
```

高 2 bits 的三種值：`0x80`、`0x40`、`0x00`。程式只比較 `ds:9168h` 與 `0x80`，
兩個載入器都用 `xor 0C0h` 在兩態之間翻轉，所以 `0x80`／`0x40` 分別對應兩張磁碟
（也就是 `GAME2`／`GAME1`）是**強證據**；`0x00`（例如第 15 項的 `0E`）目前**未解**，
不要當成第三張磁碟。

## 2. 資源開頭是 4 bytes magic

seek 之後載入器先讀 4 bytes，再依讀到的內容決定後續：

```asm
0x18404  mov  word ptr ds:92EAh, 4
0x1840A  call sub_119DB           ; 先讀 4 bytes
0x1840F  mov  bx, ds:46B0h
0x18413  mov  cx, [bx]            ; 從 ds:46B0h 指向處取一個 word
0x18415  mov  dx, 0
0x18418  call sub_11A10           ; 再讀 cx bytes
0x1841B  call sub_118C3           ; close
```

那 4 bytes 是 **magic**：`game1` 的每個資源都是 `msq0`，`game2` 都是 `msq1`
（42 個資源無一例外，見 `docs/re/07`）。

`cx` 取自 `ds:46B0h` 指向的位置，而 `ds:46B0h` 在 `sub_1841F` 開頭被設成 `0x1800` 或 `0x600`
——那是緩衝區位址，不是剛讀進來的 header。第二次讀取的長度從哪裡來**仍未解**。

`sub_1841F` 另外會依 `ds:4655h` 與另一張表 `[bx-40E4h]`（`ds:BF1Ch`）
決定緩衝區大小是 `0x1800` 還是 `0x600`——不同類資產有不同上限。

## 3. 文字輸出層

| 函式 | callers | 作用 |
|---|---:|---|
| `sub_1786E` | 28 | 印出 `ds:4680h` 指向的字串，逐 byte 直到 `0`，印完把指標推過結尾 |
| `sub_1789C` | — | 印一個字元（`sub_1786E` 的內層） |
| `sub_10039`（overlay slot 19） | 71 | 真正把字模畫進 EGA 視訊記憶體（`docs/re/04` §5） |

`ds:4680h` 是全域字串指標。開檔失敗時的
`mov word ptr ds:4680h, 9274h` ＋ `call sub_1786E`（`docs/re/03` §2）
與這裡是同一套機制，兩處互相印證。

**`wl.exe` 內的介面文字是明文 ASCII**（`Yes`、`CREATE DELETE PLAY`、
`YOU ARE NOT SMART ENOUGH!`、`Money = $` 等，見 `docs/re/generated/ida94/file-io.json`），
沒有經過編碼。中文化時這一批可以直接處理，但受限於 8×8 字型與固定欄寬。

至於劇情文字，目前**沒有證據**顯示它在 `wl.exe` 裡——合理的位置是 `GAME1`／`GAME2`，
編碼方式未知。先前把 `0x29FAE` 的 ` etraoishlnd` 之類字串當成頻率表是**沒有根據的猜測**，
追下去發現 `0x2786D` 的 `bcdefghijklmdenopq` 其實是被 `sub_166D3` 逐字印出的**畫面內容**，
不是解碼表。這類字串在證實之前不要當線索用。

## 4. 未解與下一步

| 項目 | 狀態 |
|---|---|
| 資源目錄高 2 bits 的 `0x00` | 未解 |
| 第二次讀取的長度來源 | 未解 |
| 資源總數 | 已定：目錄 50 項、表 A 20 筆、表 B 22 筆（`docs/re/07` §2） |
| 每個資源編號對應什麼資產（地圖／文字／圖） | 未解——下一步的主線 |
| 劇情文字的存放與編碼 | 未解 |
| `sub_11534`（seek）／`sub_119DB`／`sub_11A10` 的完整介面 | 只知大致職責 |

## 5. 重跑方式

```sh
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
    workplace/analysis/dumps/text-and-loaders.json 0x1786E 0x183B1 0x1841F
```

三張表可直接從解包映像讀（線性 → 檔案位移 ＝ 線性 − `0x10000` ＋ 176）：
目錄 `0x28CE9`、表 A `0x28A9A`、表 B `0x28AEA`。

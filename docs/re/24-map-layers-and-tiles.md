# 24：地圖的三層結構與 `ALLHTDS` 圖磚

日期：2026-08-14 ｜ 對應盤點 **A5**（地圖）、**A8**（圖磚像素格式）、**A10**（`ALLHTDS`）、
**A12** 的 `IC0_9.WLF` 與 `MASKS.WLF`

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2`／`allhtds1`／`allhtds2` 的 SHA-256 見 `docs/re/01`。

---

## 1. 地圖是正方形，邊長寫在記錄區標頭

邊長 D 由記錄區標頭的 `+0x2C` 一個 byte 決定，而且**同一個值同時當寬與高**
（`sub_18350`）：

```
0x1836A  b32c        mov  bl, 2Ch
0x1836C  8b3eb046    mov  di, ds:46B0h    ; 記錄區起點 P
0x18370  8a01        mov  al, [bx+di]
0x18372  a2a446      mov  ds:46A4h, al    ; 寬
0x18375  a2a546      mov  ds:46A5h, al    ; 高 ← 同一個 al
```

實測 42 個區塊只有兩種值：**D ＝ 32（38 塊）／D ＝ 64（4 塊）**。

## 2. 三層，全部以 D × D 為單位

| 層 | 位置 | 每格 | 大小 | 取值函式 |
|---|---|---|---|---|
| 1 | 區塊位移 `0` | **4 bits** | D²÷2 | `sub_17C20` → `ds:46AAh` |
| 2 | 區塊位移 D²÷2 | 1 byte | D² | `sub_17C72` → `ds:46ACh` |
| 3 | **Huffman 尾段**，載到 `ds:3448h` | 1 byte | D² | `sub_17FC8` → `ds:46CAh` |

第 1 層說「這一格有什麼」、第 2 層說「是第幾筆」、第 3 層說「畫成什麼樣子」。

第 1、2 層加起來就是記錄區起點 P：

```
D ＝ 32：  512 ＋ 1,024 ＝ 1,536 ＝ 0x600
D ＝ 64：2,048 ＋ 4,096 ＝ 6,144 ＝ 0x1800
```

第 3 層的長度也剛好是 D²：**1,024（D＝32）／4,096（D＝64）**，
與 `sub_1841F` 解壓尾段時的 `mov cx, 1000h` 上限一致。

驗證（`tools/summarize_map_layers.py`）：

| 檢查 | 結果 |
|---|---|
| 前兩層填滿地圖區 | **42／42** |
| 尾段長度 ＝ D² | **42／42** |
| 第 1 層垂直自相關的最高列寬 ＝ D | **41／42** |

### 2.1 定址（`sub_17C20`）

```
0x17C37  mov  al, ds:46A5h        ; D
0x17C3A  shr  al, 1  / jb …       ; 先丟掉一位 → 乘數 ＝ D ÷ 2
0x17C3E  shr  al, 1  / jb …
0x17C42  shl  [4661], 1 / rcl [4662], 1
                                  ; 每遇到一個 0 位就把列號左移一位
0x17C5A  shr  al, 1               ; 行號 ÷ 2
0x17C62  mov  al, [bx+di]
0x17C64  jnb  …                   ; 偶數行 → 取高 4 位
0x17C66  and  al, 0Fh             ; 奇數行 → 取低 4 位
```

「左移到最低的 1 位為止」就是**乘以 D 的最低設定位**——D 是 2 的冪，所以
這等於乘以 D（第 2、3 層）或 D÷2（第 1 層的 nibble 列寬）。同一套寫法在
`sub_17C72`（第 2 層）與 `sub_17FC8`（第 3 層）各出現一次，只差先丟掉幾位。

### 2.2 第 1 層是「這一格有什麼東西」，第 2 層是「第幾筆」

`sub_14664` 的地圖迴圈裡：

```
0x1472B  call sub_17C20           ; al ← 第 1 層的 nibble
0x1472E  mov  ds:46B3h, al
0x14731  cmp  al, 3   / jz 14776
0x14735  cmp  al, 0Fh / jz 14776
…
0x14783  mov  bl, ds:0A60Dh
0x14787  mov  di, ds:46ACh        ; 第 2 層
0x1478B  mov  al, [bx+di]         ; al ← 第 2 層的 byte
0x1478D  mov  bl, ds:46B3h        ; bl ← 剛才那個 nibble
0x14791  call sub_1379E           ; → sub_17CB1（取第 al 筆、型別 bl 的記錄）
```

`sub_17CB1` 的第二個參數就是 **section 型別**（`docs/re/16` §3），所以：

- **第 1 層的 nibble ＝ 這一格屬於哪一種 section**（0 ＝ 沒東西）
- **第 2 層的 byte ＝ 該 section 裡的第幾筆記錄**

實測第 1 層只用到 `0`–`4`、`6`、`8`–`12` 這幾個值，落在 section 型別的範圍內
（`docs/re/16` 的 24 種型別，其中 3／5／15／16／17 是指標陣列型）。

推論等級：**強證據**（定址與呼叫鏈讀完，但只讀了 nibble ＝ 3／15 這一條分支）。

### 2.3 第 3 層是畫面上的圖形編號

四個讀取端（`0x147FD`、`0x16468`、`0x16742`、`0x18092`／`0x180B0`）形狀完全一樣：
把第 3 層的 byte 放進 `ah`、把一個小常數放進 `al`，寫好螢幕座標之後呼叫
overlay slot 4（`sub_1029B`）。座標由 `sub_17FEE` 從地圖座標換算：

```
ds:4685h ＝ (地圖行 − ds:464Eh) × 2     ; 螢幕的 byte 欄，一格 2 bytes ＝ 16 像素
ds:4686h ＝ (地圖列 − ds:464Fh) × 16    ; 螢幕的像素列（sub_12738 是四次 shl）
```

`sub_1029B` 用 `ds:4686h` 去查 slot 0 建的列位址表（`ds:8DF9h`，每列 40 bytes），
再加上 `ds:4685h` 得到寫入位址。

`sub_1029B` 開頭把這兩個值換算成三個指標：

```
0x102BD  xor  dx, dx
0x102BF  mov  dh, ah          ; dx ＝ ah × 256
0x102C1  xor  ah, ah          ; ax ＝ al
0x102C3  mov  cl, 5 / shl ax, cl
0x102C7  shr  dx, 1           ; dx ＝ ah × 128
0x102C9  mov  bx, ax          ; bx ＝ al × 32
0x102CB  shl  ax, 1 / shl ax, 1
0x102CF  add  ax, 420h        ; si ＝ 0x420 ＋ al × 128   ← 疊圖
0x102D4  add  bx, 0DA60h      ; di ＝ 0xDA60 ＋ al × 32   ← 遮罩
0x102DC  add  bp, 420h        ; bp ＝ 0x420 ＋ ah × 128   ← 背景
```

合成方式（`0x1035D` 起，每列兩個 word、每寫一個 word 跳 `0x26` 到下一列）：

```
螢幕 ← (背景[bp] AND 遮罩[bx]) OR 疊圖[si]
```

三個基址各自對上一個已知素材（`docs/re/03` §5）：

| 基址 | 素材 | 大小 | 一筆 | 筆數 |
|---|---|---:|---:|---:|
| `seg003:0x0420` | `IC0_9.WLF` | 1,280 | 128（4 平面 16 × 16） | **10** |
| `seg003:0x0920` | 圖磚組（`sub_10088` 轉出來的） | ≤20,864 | 128 | 66–163 |
| `seg003:0xDA60` | `MASKS.WLF` | 320 | 32（單平面 16 × 16） | **10** |

`0x420 ＋ 10 × 128 ＝ 0x920`——**`IC0_9.WLF` 與圖磚組在記憶體裡是連續的一張表**，
所以第 3 層的 byte 是這張表的索引：

```
0–9    IC0_9.WLF 的十個圖形
≥10    ALLHTDS 的圖磚，圖磚編號 ＝ 值 − 10
```

`al` 那個小常數則是疊在上面的東西（隊伍、遭遇標記等），範圍 0–9，
與 `IC0_9.WLF` 和 `MASKS.WLF` 各 10 筆一致——**兩個素材的筆數同時吻合**。

驗證：

| 檢查 | 結果 |
|---|---|
| `max(第 3 層) − 10 < 該地圖圖磚組的張數` | **42／42** |
| 值落在 0–9（`IC0_9.WLF`）的格子 | 全 42 張地圖合計 457 格（其中 448 格是 0） |
| 用第 3 層畫出縮圖 | 城鎮有牆與房間、沙漠有道路與山脈（`tools/render_map.py`） |

推論等級：**已確認**（算術、三個基址與素材大小、值域範圍、縮圖四項互相印證）。

## 3. 圖磚在 `ALLHTDS`，一張 128 bytes

`sub_186B6` 載入一組圖磚：解壓到 `seg003:0x2F60`（上限 `0x5180` ＝ 20,864），
接著跑 delta：

```
0x18720  mov  dx, 0A3h        ; 163 次（外層）
0x18723  mov  bx, 8
0x18726  mov  si, 2F60h
0x18729  mov  di, si
0x1872B  add  di, bx          ; di ＝ si ＋ 8
0x1872D  mov  cx, 3Ch         ; 60 個 word（內層）
0x18730  lodsw
0x18731  xor  [di], ax
0x18733  inc  di / inc di
0x18735  loop loc_18730
0x18737  add  di, bx          ; 兩邊各跳過 8 bytes
0x18739  add  si, bx
0x1873B  dec  dx / jnz loc_1872D
```

一輪處理 8（種子）＋ 120（delta）＝ **128 bytes**，共 163 輪 ＝ 20,864。
**回看距離 8 就是列寬**（`docs/re/23` §7 的規則）：8 bytes × 2 像素 ＝ **16 像素寬**，
128 ÷ 8 ＝ **16 列** → **16 × 16、packed 4bpp**，與圖片同一套格式。

驗證：9 個 `ALLHTDS` 子區塊的解壓長度**全部整除 128**——
66、141、163、107、127、118、90、104、135 張。

### 3.1 載入時轉成 EGA 4 平面

overlay 的 `sub_10088` 把整組圖磚從 packed 4bpp 轉成 4 平面：

```
0x1009B  mov  bp, 0A3h        ; 163 張
0x1009E  mov  cx, 20h         ; 每張 32 次
0x100A1  lodsw … shl/rcl ×32  ; 4 bytes（8 像素）拆進 bl／bh／dl／dh
0x10123  mov  [di], bl
0x10125  mov  [di+20h], bh
0x10128  mov  [di+40h], dl
0x1012B  mov  [di+60h], dh
0x10135  add  di, 60h         ; 下一張（32 × 4 ＝ 128 bytes）
```

每個平面 32 bytes ＝ 16 列 × 2 bytes，再次確認 **16 × 16**。
目的地是 `seg003:0x920`，也就是圖片緩衝區（`TITLE.PIC` 用的那塊）。
來源在 `0x2F60`、目的在 `0x920`，**寫入點永遠落在讀取點後面**，所以就地轉換是安全的。

### 3.2 哪一組圖磚由地圖決定

記錄區標頭 `+0x30` 是圖磚組編號（`sub_18350`）：

```
0x1838C  mov  bl, 30h
0x18392  mov  al, [bx+di]
0x18396  cmp  al, ds:0BF13h     ; 與目前載入的比較，一樣就不重載
0x1839F  mov  ds:4655h, al
0x183A2  call sub_186B6
```

`sub_186B6` 用 `cmp byte ptr ds:4655h, 4` 分檔：**0–3 在 `ALLHTDS1`、4–8 在 `ALLHTDS2`**
（`ds:9168h` 寫 `0x80`／`0x40` 選檔，索引 ≥4 時再減掉表的第 4 筆位移）。

實測 42 個區塊用到的編號正好是 **0–8 九個**，與兩個檔案裡 4 ＋ 5 ＝ 9 個子區塊
一一對應。

## 4. 可重跑的完整指令

```bash
python3 tools/render_map.py \
  workplace/analysis/unpacked/wl.merged.exe \
  workplace/orig/wastland/game1 workplace/orig/wastland/game2 \
  workplace/orig/wastland/allhtds1 workplace/orig/wastland/allhtds2 2

python3 tools/summarize_map_layers.py \
  workplace/analysis/unpacked/wl.merged.exe \
  workplace/orig/wastland/game1 workplace/orig/wastland/game2 \
  docs/re/generated/ida94/map-layers.json

python3 tools/decode_pic.py tile workplace/orig/wastland/allhtds1 0 0

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/maps.json \
  0x17C20 0x17C72 0x17FC8 0x18350 0x1841F 0x186B6 0x10088 0x1029B 0x1379E --callers
```

## 5. 還沒解的

- 第 1 層 nibble 為 3／15 以外的值走哪條路（`sub_14664` 只讀了那一條分支）。
- `IC0_9.WLF` 那十個圖形各自是什麼。**第 7 張已經解出來：隊伍圖示**
  （`sub_16716` 的 `mov al, 7` ＋ overlay slot 4，實機對拍確認，`docs/re/47` §6.4）。
  另外九張還沒一一對上。
  順帶：`MASKS.WLF` ＝ 320 bytes ＝ **10 張 × 32**，一張是 16 列 × 2 bytes、
  **一個位元一個像素**（不是 4 平面）。
- 疊圖編號 `al` 的來源：`sub_14664` 走 `ds:0AA17h` 那張表，其餘呼叫端是常數
  5／7／9，對應關係還沒逐一對上。
- 地圖視窗的螢幕幾何（盤點 B6）：`sub_1029B` 每列寫兩個 word、跳 `0x26`，
  列位址查 `ds:8DF9h` 的表；`sub_10F12` 另外把 `seg003:0x920` 的
  `0x24` × `0x80` 區塊整塊搬上螢幕。兩條路徑還沒對上。
- `CURS`（`seg002:0x7E0B`，2,048 bytes）仍未解。

## 6. 這一輪學到的（寫成規則）

- **「乘以某個變數的最低設定位」是 2 的冪維度的慣用寫法。** 看到
  `shr al,1 / jb 出口 / shl 值,1 / jmp` 這種迴圈，不要當成位元運算，
  那是**乘以邊長**；先丟掉幾位就是先除以 2 的幾次方，也就是每格幾個 bit。
- **量測要挑對層。** 先前把整個地圖區當成單一 nibble 平面做自相關，
  峰值出現在 64 nibble，於是得到「地圖 64 × 48」。實際上峰值來自
  **占三分之二的 byte 層**（D ＝ 32 的 byte 列在 nibble 視角下正好是 64）。
  資料是對的、統計是對的，**切層切錯，結論就錯**。
- **一個 byte 同時寫進兩個變數，通常代表正方形。** `ds:46A4h`／`ds:46A5h`
  來自同一個 `al`，這比任何尺寸推測都直接。
- **索引超出範圍一點點，通常是「前面還接了另一張表」。** 第 3 層的值最多
  超出圖磚張數 10，而 `IC0_9.WLF` 正好是 10 筆、正好排在圖磚組前面
  （`0x420 ＋ 10 × 128 ＝ 0x920`）。**先去看那個緩衝區前面放了什麼**，
  比猜「大概有追加圖形」快得多。

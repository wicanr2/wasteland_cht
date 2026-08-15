# 23：圖片格式 —— packed 4bpp ＋ 列間 XOR delta

日期：2026-08-14 ｜ 對應盤點 **A9**（`ALLPICS`）、補完 **A14** 與 `docs/re/03` §6

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`allpics1/2`、`title.pic` 的 SHA-256 見 `docs/re/01`。

---

## 1. 結論

原版的圖片全部是**同一套格式**，只有尺寸不同：

```
packed 4bpp        一個 byte 兩個像素，高 4 位在左
列間 XOR delta     out[n + stride] ＝ in[n + stride] XOR out[n]
                   **XOR 的回看距離就是一列的 byte 數**
```

| 來源 | 大小 | stride | 尺寸 | 解碼碼在哪 |
|---|---:|---:|---|---|
| `ALLPICS1/2` 的圖片子區塊 | 4,032 | 48 | **96 × 84** | overlay slot 2（`sub_10144`） |
| `TITLE.PIC` | 18,432 | 144 | **288 × 128** | `start` 內嵌（`docs/re/03` §6） |
| `END.CPA` | 18,432 | 144 | 288 × 128 | `sub_1B7FE` |

`stride × 高 ＝ 檔案大小` 三個來源都整除，而且**每張圖都用滿 16 種顏色**——
與 EGA mode 0Dh 的 16 色調色盤一致。

## 2. 解碼：`sub_10144`（overlay slot 2）

整支只有 18 bytes：

```
0x10144  8b362d47  mov  si, ds:472Dh     ; 圖片緩衝區
0x10148  bb2e00    mov  bx, 2Eh          ; 46
0x1014B  b9c807    mov  cx, 7C8h         ; 1,992 次
0x1014E  ad        lodsw                 ; ax ← [si]，si += 2
0x1014F  3100      xor  [bx+si], ax      ; [si + 46] ^= ax
0x10151  e2fb      loop loc_1014E
```

`lodsw` 之後 `si` 已經加 2，所以寫入點是**讀取點 ＋ 48**。
1,992 × 2 ＝ 3,984，加上跳過的前 48 bytes 正好是 **4,032**——緩衝區剛好用完。

`si ≥ 48` 之後讀到的是**已經解過的內容**，所以這是滾動的自參考解碼，
順序不能顛倒。與 `TITLE.PIC` 的 `0x111C5` 那段是同一個寫法，只有距離不同
（`0x90` ＝ 144）。

## 3. 為什麼會走錯一次

第一次把它當成 EGA 4 平面（`COLORF.FNT` 是 4 平面，`docs/re/14` §3），
畫出來是雜訊。用兩個量測把方向掰回來：

| 量測 | 未解碼 | 解碼後 |
|---|---:|---:|
| `0x00` 的個數 | 3,077 ／ 4,032 | 661 |
| 位元組自相關的最高峰 | stride **48**（0.696） | stride **48**（0.768） |

- 自相關在 48 有明顯尖峰 → **一列就是 48 bytes**，與 XOR 距離一致。
- 解碼後自相關**上升**（0.696 → 0.768）→ 方向對了（delta 解開後列與列更像）。
- `4,032 ＝ 96 × 84 ÷ 2` → 4bpp 而不是 4 平面。

**「同一個檔案裡有 4 平面字型」不代表圖片也是 4 平面。**
`COLORF.FNT` 是 4 平面（字元要用 EGA 的 map mask 一次寫一個平面），
圖片是 packed 4bpp（整塊搬進畫面）——兩種在同一個遊戲裡並存。

## 4. `ALLPICS` 的容器結構

`split_all` 的結果是嚴格交替：

```
allpics1: 66 個子區塊 ＝ 33 × (4,032 圖片 ＋ 變動長度參數區)
allpics2: 98 個子區塊 ＝ 49 × (4,032 圖片 ＋ 變動長度參數區)
```

共 **82 張圖片**。參數區長度 430–2,490 bytes 不等，內容是**局部動畫**——
`sub_10A7A`（overlay slot 16）拆成兩張表，`sub_10B11` 逐格疊上去，見 §5。

載入在 `sub_184E8`：先解壓 `0xFC0` bytes 到 `ds:472Dh`（圖片），
再解壓參數區到 `ds:2700h`（或圖片編號 `0x34` 時到 `ds:0`），
然後呼叫 slot 2 解 delta、slot 16 處理參數。
`ds:BF12h` 是目前的圖片編號、`ds:BEFBh` 是上一張——**相同就直接返回**，
所以重複要求同一張圖不會重解一次。

**圖片編號是敵人資料的 `+0x07`**（`docs/re/37` §3.2）：遭遇時載入那種敵人的
肖像圖，而同一個編號查 `ds:A920h` 還會決定文字裡用 him／her／it。

## 5. 參數區 ＝ 局部動畫

### 5.1 兩張表（`sub_10A7A`）

overlay slot 16 把交錯在圖片之間的參數區拆成**兩張變動長度的記錄表**：

| 表 | 長度欄 | 分隔 | 一筆長什麼樣 | 指標存到 |
|---|---|---|---|---|
| A | 開頭一個 word | `0xFF` | **`(延遲, 格編號)` 交錯**，第一個 byte 是初始延遲 | `ds:8FC0h` |
| B | 接在 A 之後一個 word | `0xFFFF` | word 標頭 ＋ `(word >> 12) + 1` bytes 酬載 | `ds:8FD4h` |

第三趟（`0x10AF1`）把表 A 每一筆的**第一個 byte 抽出來當初始倒數**放
`ds:8FBBh`、指標往後挪一格放 `ds:8FCAh`——所以那個 byte 是延遲不是格編號，
記錄從第二個 byte 起才是「格、延遲、格、延遲…」。

**表 A ＝ 每個通道的播放腳本，表 B ＝ 每一格的像素。**

#### 這些位移落在哪：**正好是會動的那一塊**

`tools/summarize_picparams.py` 把 `allpics1` 全部拆出來，每張圖的表 B
座標範圍都很窄（例如第 3 張是 `x 56–79、y 8–18`）——建築物右上角那個
碟形天線的位置，與 `docs/re/54` 拿實機截圖對出來的差異區完全同一塊。

推論等級：兩張表的**結構** **已確認**（`sub_10A7A` 逐條讀完，
`allpics1`／`allpics2` 全部拆得開、長度自洽，且重建結果通過實機驗收，見 §5.3）。

### 5.2 疊法：overlay slot 17（`sub_10B11`）

消費端是 **overlay slot 17**（`0x10B11`，`docs/re/04` §3 記成「動畫更新」）：

```
al ← BIOS 0040:006C（計時器低位元組）
al ＝ ds:9164h → 直接 retn            ; 這一格還沒到，不動
每個通道 bx ＝ 0 … ds:8FB9h − 1：
    [ds:8FBBh + bx] ≠ 0 → 減一，跳過   ; 倒數計時
    否則：
        si ← [ds:8FCAh + bx×2]；lodsb  ; 表 A 的下一個 byte ＝ **格編號**
        si ← [ds:8FD4h + 格編號×2]     ; 表 B 的那一格
        逐 word 疊上去（見下）直到 0xFFFF
        再讀表 A 的下一個 byte：
            0xFF → si ← [ds:8FC0h + bx×2]  ; **回到開頭，循環**
            其餘 → 寫進 [ds:8FBBh + bx]     ; **下一次要等幾格**
```

所以**表 A 是每個通道的腳本**（格編號與延遲交錯、`0xFF` 循環），
**表 B 是每一格的像素**。

一個元素 word 的解法（`0x10BA2`–`0x10BD0`）：

```
相位 ← al & 3；dl ← 3 − 相位
ax  >>= 2
長度 ← (ah >> 2) → ds:8FB5h            ; 實際 bytes ＝ 長度 ＋ 1
ah  &= 3
ax ÷ 12 → 列（商）、欄（餘）            ; **一列 12 bytes ＝ 96 像素 ÷ 8**
di ← 列位址表[列] ＋ 欄 ＋ 1
```

酬載的每個 byte 被拆成四個平面位元（八次 `shl`／`rcl`），
然後**逐平面 XOR 進螢幕**（`xor es:[di], bl` 等，中間切 EGA 的
sequencer map mask 與 graphics 暫存器）。

> **是 XOR 進螢幕，不是疊回圖片緩衝區**（`xor es:[di]`，ES ＝ 0xA000）。
> 原版沒有回頭改那份 96 × 84 的解碼緩衝區。

**列位址表就是 overlay slot 0 那張。** slot 0 在 `0x10064` 從 `ds:8DF9h`
起建 200 個 word 的 `y × 40`（`docs/re/04` §4）；`sub_10B11` 索引的
`[bx−71F7h]` ＝ `ds:8E09h` ＝ **同一張表往後 8 個 word**。
所以它查到的是**螢幕列 ＝ 圖片列 ＋ 8**，`+ 欄 + 1` 再給 x ＋ 8——
正好是設施圖擺在 `(8, 8)` 的那個位置（`docs/re/54` §2），**位移已經烘在表裡**。

**相位是左邊缺幾對像素。** `dl = 3 − 相位` 讓第一組只吃 `4 − 相位` 個 byte，
而累加器由 0 往左移，所以那幾個像素落在**低位**（螢幕 byte 的右半）——
等於這一段從 `欄 × 8 + 2 × 相位` 開始畫。收尾那組不足 4 個時
`0x10C00` 補左移，把它推到高位，所以尾巴切在酬載用完的地方。

### 5.3 實機驗收：逐像素 0 差異

拿上面的解法重建第 3 張圖的動畫，照表 A 的順序累積 XOR，
與 Ranger Center 的實機截圖比對：

```
底圖（一格都沒疊）：差 126 像素
  疊到第 3 步（格 3）：差 0 像素        ← 截圖抓到的就是這一格
  疊到第 8 步（格 8）：差 126 像素      ← 一輪播完回到底圖
```

**一輪播完 XOR 互相抵消**，這一點在 `allpics1`／`allpics2` 的
**82 張圖上全部成立**。這是全檔級的自洽佐證：相位、位置、長度、播放順序
任何一項讀錯，這個恆等式都不可能對 82 張都對。
原版正是靠它才能無限循環播放而不必存背景。

推論等級：**已確認**——`sub_10B11` 的控制流逐條讀完，
重建結果與實機逐像素吻合，且恆等式在全部 82 張圖上成立。

## 6. 還沒解的

- `ALLHTDS1/2`：4／5 個 8,448–20,864 bytes 的大塊，**大小各不相同**，
  所以不是固定尺寸的圖；`sub_186B6` 解壓到 `seg003:0x2F60`。用途未解。
- ~~`END.CPA` 的解碼~~ → **已解**，見 §9。第二塊（`0x3A98` bytes，
  多半是結局敘述，`ds:D18Eh` 的字串表）還沒解。

## 7. 調色盤：原版從來沒設過

**原版從來沒設過調色盤**，用的就是 mode 0Dh 的預設 16 色。兩層證據：

- **實機**：`TITLE.PIC` 的解碼結果與原版畫面**逐像素 100% 吻合**，
  比對用的就是預設的 EGA 16 色——若調色盤被改過不可能全中（`docs/re/47` §4）。
- **全檔掃描**：碰 EGA 埠的指令幾乎都是 sequencer（`0x3C4`／`0x3C5`）、
  graphics（`0x3CE`／`0x3CF`）與等垂直同步（`0x3DA`）。
  碰 `0x3C0`（Attribute Controller，調色盤暫存器在那裡）的**只有一處**，
  而且它設的是**邊框色**，不是那 16 格（見下）。
  BIOS 那條路也堵了：全檔只有兩處 `int 10h`，都在 overlay slot 0
  （`0x10052` 設 mode 0Dh、`0x1005A` 回文字模式 3），**沒有 `AH ＝ 10h`**。
  這一條是 bytes 層掃出來的（在映像裡找 `cd 10`），不是只信 IDA 的反組譯——
  兩種數法都是 2 處，位址也一致。

### 7.1 唯一碰 `0x3C0` 的那支，入口被 patch 成 `retn`

```
0x10E9A  c3           retn            ← 三個呼叫端全部落在這裡
0x10E9B  8a c8        mov  cl, al     ← 真正的程式碼從這裡開始，沒有呼叫端
0x10E9D  fa           cli
0x10E9E  ba da 03     mov  dx, 3DAh   ; 讀 status register 重置 flip-flop
0x10EA1  ec           in   al, dx
0x10EA2  b2 c0        mov  dl, 0C0h   ; → 0x3C0
0x10EA4  b0 11        mov  al, 11h    ; index 0x11 ＝ **overscan（邊框）色**
0x10EA6  ee           out  dx, al
0x10EA7  8a c1        mov  al, cl
0x10EA9  ee           out  dx, al
0x10EAA  b0 20        mov  al, 20h    ; bit5 ＝ PAS，重新開啟畫面輸出
0x10EAC  ee           out  dx, al
0x10EAD  fb           sti
0x10EAE  c3           retn
```

`sub_10E45` 裡三個 `call`（`0x10E56` 傳 al ＝ 0、`0x10E81` 傳 4、`0x10E8C` 傳 3）
算出來的目標都是 `0x10E9A`，而那裡是一條 `retn`——**出貨版把這支停用了**，
三次呼叫一次都沒有效果。IDA 把 `0x10E9A` 命名為 `nullsub_1`。

即使它沒被停用，改的也只是 index `0x11`（邊框），**那 16 格
（index `0x00`–`0x0F`）全檔沒有任何人寫**。

⚠ `mov dl, 0C0h` 的另一處在 `0x2ABF9`，那落在資料段（`ds:DDD9h`）、
前後都是雜資料、後面沒有 `out`——是巧合，不是第二個寫入點。

這一項原本掛在盤點 A14，缺的就是這個全檔掃描。

## 8. 可重跑的完整指令

```bash
python3 tools/decode_pic.py allpics workplace/orig/wastland/allpics1 0
python3 tools/decode_pic.py title   workplace/orig/wastland/title.pic

# 動畫的實機驗收（差 0 才 exit 0）
python3 tools/verify_pic_anim.py workplace/orig/wastland/allpics1 3 \
  --screen workplace/dosbox/shots/db05.ppm --at 8,8

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/pictures.json \
  0x10144 0x10A7A 0x184E8 0x186B6 0x1B7FE --callers
```

## 9. 這一輪學到的（寫成規則）

- **XOR 自參考解碼的回看距離，就是那筆資料的列寬。** 這是免費的幾何資訊：
  `TITLE.PIC` 的 `0x90` 給了 288 px、`ALLPICS` 的 48 給了 96 px，
  兩個都不用猜。看到 delta 解碼先把距離記下來當 stride。
- **像素格式要量，不要類推。** 同一個遊戲裡 4 平面與 packed 4bpp 並存，
  用途不同（字型 vs 整塊圖）。判斷方法是算術：
  `4,032 ＝ 96 × 84 ÷ 2` 對得上 4bpp，對不上 4 平面。
- **「解碼後更有結構」是可以量的。** 自相關上升、尖峰位置不變，
  就是方向對了；靠肉眼看文字圖容易被雜訊誤導。
- **重建不出畫面時，先確認結構讀對了，不要調參數。** 三種疊法都在
  80–130 之間游移是**參數在補結構的錯**——真正缺的是「哪一格」與
  「相位是什麼」兩件結構性的事，補上之後直接掉到 0，中間沒有過渡。
  **殘差在一個區間裡浮動而不逼近 0，就是還有一項讀錯了。**
- **偏移量會被烘在查表裡。** 列位址表擺在 `ds:8DF9h`，消費端卻從
  `ds:8E09h` 索引——差的 8 個 word 就是圖片在螢幕上的 y 位移。
  看到「基址對不上」先算差幾筆，那個差值通常本身就是答案。
- **找得到一個全檔都該成立的恆等式，就用它當驗收。** 「一輪播完回到底圖」
  比單張截圖強得多：單張只證明某一組參數在某一格對得上，
  恆等式要求整組解析在 82 張圖上都自洽。

## 9. `END.CPA`：整份就是 Huffman

```
sub_1B7FE:
  sub_11445(dl ＝ 6)                     ; 開資源索引 6
  sub_11AE8(1)                           ; 讀 8 bytes 段標頭 ＋ 驗 'ms' magic
  sub_11B83(cx ＝ 4800h, dx ＝ 920h)     ; 解出 0x4800 bytes 到 seg003:0x920
  sub_11AE8(0)                           ; 第二段的標頭（不驗）
  sub_11B83(cx ＝ 3A98h, dx ＝ 5120h)    ; 解出 0x3A98 bytes 到 seg003:0x5120
```

`sub_11B83` **就是 Huffman 解壓器**（與 `docs/re/11` 同一套）：

```
di ← ds:9509h              ; 樹根
loc_11BC3:
  [di] ＝ 0 → 葉節點        ; 節點 6 bytes：[di] 左、[di+2] 右、[di+4] 值
  位元用完 → lodsb 取下一個 byte、ah ← 80h
  test ah, al → 右／左
  shr al, 1
葉節點:
  al ← [di+4]；stosb
```

所以檔案的第 0 個 byte 起就是標準串流：**前 4 bytes 是解開後的長度 `0x4800`**
＝ 288 × 128 的 packed 4bpp，與 `sub_1B7FE` 的 `cx`、`TITLE.PIC` 的尺寸三邊對上。
解完再走列間 XOR delta（stride 144），與 `TITLE.PIC` 同一條路。

⚠ **不要先跳過那 4 bytes。** 把它當「檔頭」跳掉、再拿 `+0x04` 的 `msq` 當
MSQ 容器去解密，六種 checksum／body 起點的組合全部失敗——那是在猜參數。
`Decompress` 自己會讀長度欄，整份丟進去就對了。

解錯時**值域一樣是 0–15**，值域檢查完全擋不住；是**顏色分布**抓出來的——
解錯時最多的一種顏色只佔 6%（雜訊的均勻分布），正確的圖是 22%。
`TestEndPicture` 的門檻因此寫成分布而不是值域。

remake 這一側：`Rom.End()`（`internal/assets/pic.go`）＋ `wl-shot -mode end`。

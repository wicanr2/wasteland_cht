# 13：亂數產生器與擲骰層

日期：2026-08-14 ｜ 對應盤點 **D7**（`docs/re/00-remake-knowledge-gaps.md`）

輸入：`wl.merged.exe`（解包映像 ＋ `wla.bin` overlay，本專案合成），
SHA-256 `cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

---

## 1. 怎麼找到的

上一輪把 `mul`／`imul` 全掃過，六處都不是 RNG，於是排除了乘法型（LCG）。
這一輪從**形狀**下手而不是從指令下手：RNG 一定是「讀一個全域、算一算、寫回同一個
全域」，而且被很多地方呼叫。

第一次嘗試用 xref 圖問「誰在寫這個變數」，得到 22 個位址——**這是假零**。
這份資料庫裡 IDA 只替 22 個位址建了資料 xref，`mov ax, ds:46B7h` 這類直接定址
根本沒進 xref 圖。改成自己解碼運算元（`tools/ida/export_memops.py`），
同樣的問題得到 **827 個全域變數、4,932 筆直接定址存取**。

之後把整個 CODE 區倒成逐指令 JSON（`tools/ida/export_listing.py`，20,177 條指令），
所有形狀比對都變成離線的集合運算。篩「同一函式讀寫同一位址 ＋ 含移位／加法」
就在第四名撞到 `sub_18E41`，而它呼叫的 `sub_18E6B` 就是產生器本體。

## 2. 產生器本體：`sub_18E6B`

線性位址 `0x18E6B`，`seg000+0x8E6B`，37 bytes，3 個呼叫端。
原始 bytes（直接從 `wl.merged.exe` 檔案讀出，未經 IDA）：

```
fe065c46 02065c46 12065d46 a25d46 12065e46 a25e46
12065f46 a25f46 12066046 a26046 c3
```

未過濾的反組譯：

```
0x18E6B  fe065c46    inc     byte ptr ds:465Ch
0x18E6F  02065c46    add     al, ds:465Ch
0x18E73  12065d46    adc     al, ds:465Dh
0x18E77  a25d46      mov     ds:465Dh, al
0x18E7A  12065e46    adc     al, ds:465Eh
0x18E7E  a25e46      mov     ds:465Eh, al
0x18E81  12065f46    adc     al, ds:465Fh
0x18E85  a25f46      mov     ds:465Fh, al
0x18E88  12066046    adc     al, ds:4660h
0x18E8C  a26046      mov     ds:4660h, al
0x18E8F  c3          retn
```

寫成式子（`s0`…`s4` ＝ `ds:465Ch`…`ds:4660h`，`AL` 是傳進來也傳出去的值）：

```
s0 ← s0 + 1
AL ← AL + s0                    ; 進位留著
AL ← AL + s1 + carry ; s1 ← AL
AL ← AL + s2 + carry ; s2 ← AL
AL ← AL + s3 + carry ; s3 ← AL
AL ← AL + s4 + carry ; s4 ← AL   ; 回傳 AL
```

`s0` 是純計數器，其餘四個位元組是它的**四重前綴和**，中間靠進位互相耦合。
沒有乘法、沒有除法、沒有查表——這解釋了上一輪為什麼掃 `mul` 一無所獲。

**`AL` 是介面的一部分。** 進去的 `AL` 會參與運算，出來的 `AL` 是新的 `s4`，
而兩個包裝函式都把上一次的結果餵回下一次。重製時不能只搬狀態、不搬 `AL`。

推論等級：**已確認**（程式碼讀完、逐位元組對過、行為以模型復現）。

### 2.1 狀態變數：`ds:465Ch`–`ds:4660h`

| 項目 | 值 |
|---|---|
| 線性位址 | `0x2147C`–`0x21480` |
| `segment:offset` | `seg002+0x465C`–`seg002+0x4660` |
| 大小 | 5 bytes |
| 映像裡的初值 | `00 00 00 00 00` |

**全檔只有 `sub_18E6B` 會寫這五個位元組**，沒有任何地方設種子。兩道全檔掃描：

| 掃描 | 對象 | 結果 |
|---|---|---|
| 直接定址 | 20,177 條指令的每一個 `o_mem` 運算元 | 寫入點只有上面那五條 `mov` |
| 取址 | 任何以 `0x4650`–`0x4670` 當立即數／位移的指令 | **零** |

第二道是必要的：xref 圖與直接定址掃描都看不到「把位址當純數字算出來再間接寫」
的程式碼。沒有任何地方取這五個位元組的位址，就排除了那條路。
剩下的理論漏洞只有「某張表的基底加索引剛好落進這個範圍」，
目前已知的基底（`0x7131` 角色記錄、`0x6B31`）都不在附近。

所以每次開機的序列完全相同，**亂數的變化來源只有「產生器被呼叫了幾次」**。

那個次數來自玩家：`sub_18EFE`（鍵盤輪詢，`0x18F09`）每輪詢一次就叫一次
`sub_18E6B`，而 `sub_18E90`（等待按鍵，28 個呼叫端）會一直輪詢到有鍵為止。
**玩家在選單前猶豫多久，就是這個遊戲的熵。**

推論等級：**已確認**（初值直接從映像讀出，寫入點是全檔掃描的結果）。

## 3. 擲骰層

四支包裝函式，全部建立在 `sub_18E6B` 上。

| 函式 | 位址 | 大小 | 呼叫端 | 語意 |
|---|---|---:|---:|---|
| `sub_18E5F` | `0x18E5F` | 12 | 5 | d6，回傳 1..6 |
| `sub_18E41` | `0x18E41` | 30 | 24 | dN，回傳 1..N |
| `sub_19D86` | `0x19D86` | 44 | 3 | 累加 N 顆 d6，16-bit 結果 |
| `sub_19C84` | `0x19C84` | 40 | 11 | 2d6，逢同點續擲並累加 |

### 3.1 `sub_18E5F`：d6

```
0x18E5F  e80900   call    sub_18E6B
0x18E62  2407     and     al, 7
0x18E64  3c06     cmp     al, 6
0x18E66  73f7     jnb     short sub_18E5F     ; >= 6 就整支重來
0x18E68  0401     add     al, 1
0x18E6A  c3       retn
```

取低三位元得 0..7，**6 和 7 直接重抽**，再加一 → 1..6。
是拒絕取樣，不是取模，所以六面等機率（模擬 60 萬次，各面偏差 < 0.7%）。

### 3.2 `sub_18E41`：dN

進入時 `AL` ＝ 面數 N。

```
0x18E41  a259c0     mov     ds:0C059h, al       ; 存下 N
0x18E44  3c01       cmp     al, 1
0x18E46  7616       jbe     short locret_18E5E  ; N<=1 原樣回傳
0x18E48  b400       mov     ah, 0
0x18E4A  f9         stc                          ; ┐ 造遮罩
0x18E4B  d0d4       rcl     ah, 1                ; │ ah = (ah<<1)|1
0x18E4D  d0e8       shr     al, 1                ; │
0x18E4F  75f9       jnz     short loc_18E4A      ; ┘ 跑 N 的位元數次
0x18E51  e81700     call    sub_18E6B
0x18E54  22c4       and     al, ah               ; 遮到 0..2^k-1
0x18E56  3a0659c0   cmp     al, ds:0C059h
0x18E5A  73f5       jnb     short loc_18E51      ; >= N 就重抽
0x18E5C  fec0       inc     al
0x18E5E  c3         retn
```

遮罩是「不小於 N 的最小的 2^k−1」（N=6 → `0b111`），再拒絕取樣到 0..N−1，
最後加一 → **1..N 閉區間，等機率**。`ds:0C059h`（`0x28E79`）只是暫存 N，
除了這支沒有別人碰。

邊界：**N=0 回傳 0，N=1 回傳 1**（`jbe` 那條直接跳出，`AL` 原樣送回）。

### 3.3 `sub_19D86`：累加 N 顆 d6

進入時 `AL` ＝ 基底、`DL` ＝ 顆數；回傳 `AL` ＝ 低位、`DL` ＝ 高位。

```
0x19D86  mov     ds:0CD3Fh, al      ; 累加器低位 ← 基底
0x19D89  mov     bl, 0              ; 累加器高位
0x19D8B  cmp     dl, 0
0x19D8E  jz      short loc_19DAA
0x19D90  call    sub_18E5F          ; d6
0x19D93  add     al, ds:0CD3Fh
0x19D97  mov     ds:0CD3Fh, al
0x19D9A  jnb     short loc_19DA6
0x19D9C  inc     bl                 ; 低位溢位 → 高位加一
0x19D9E  jnz     short loc_19DA6
0x19DA0  mov     bl, 0FFh           ; 高位也溢位 → 飽和成 0FFFFh
0x19DA2  mov     ds:0CD3Fh, bl
0x19DA6  dec     dl
0x19DA8  jnz     short loc_19D90
0x19DAA  mov     al, bl
0x19DAC  mov     dl, al
0x19DAE  mov     al, ds:0CD3Fh
0x19DB1  retn
```

也就是 `base + XdY` 的 X 部分。模型跑 `sum_d6(0, 6)` 二十萬次得平均 20.99、
值域 6–36，與 6d6 的理論值 21 相符。

三個呼叫端：`0x141D0`（`sub_14193`）、`0x14442`、`0x1585E`（`sub_157D6`）。

### 3.4 `sub_19C84`：2d6，同點續擲

```
0x19C84  mov     al, 0
0x19C86  mov     ds:0CD4Ah, al      ; 累加器歸零
0x19C89  call    sub_18E5F          ; 第一顆
0x19C8C  mov     ds:0CD49h, al
0x19C8F  call    sub_18E5F          ; 第二顆
0x19C92  push    ax
0x19C93  clc
0x19C94  adc     al, ds:0CD49h      ; 累加器 += 兩顆之和
0x19C98  adc     al, ds:0CD4Ah
0x19C9C  mov     ds:0CD4Ah, al
0x19C9F  pop     ax
0x19CA0  cmp     al, ds:0CD49h
0x19CA4  jz      short loc_19C89    ; 兩顆同點 → 再擲一對，繼續累加
0x19CA6  mov     al, ds:0CD4Ah
0x19CA9  mov     dl, 0
0x19CAB  retn
```

擲一對 d6 累加，同點就再來一對。是會「爆開」的骰：最小值 3（1+2），
期望值 2d6 的 7 再加上 1/6 機率的續擲，理論 8.4，模型跑 30 萬次得 8.37。
11 個呼叫端，是這一族裡分佈最特別的一支。

## 4. 參考模型與驗證

`tools/rng.py` 是五支函式的 Python 模型，附自我測試：

```bash
python3 tools/rng.py        # 全部通過
```

驗證項目與結果：

| 檢查 | 結果 |
|---|---|
| 全零狀態起步的前 7 個輸出 ＝ `C(n+4,5) mod 256` | 通過（`1 6 21 56 126 252 206`） |
| 第 8 項起與二項式分歧（進位開始回饋） | 通過（實際 `25`，二項式 `24`） |
| `roll(0)`＝0、`roll(1)`＝1 | 通過 |
| `roll(3)` 值域 ＝ {1,2,3}；`roll(100)` 值域 ＝ 1..100 | 通過 |
| d6 六面分佈（60 萬次） | 各面偏差 < 3,000／100,000 |
| `sum_d6(0,6)` 平均 ＝ 21.0 ± 0.1 | 通過（20.99） |
| `pair_d6()` 平均 ＝ 8.4 ± 0.1 | 通過（8.37） |
| 300 萬次呼叫內狀態不重複 | 通過 |

前七項等於二項式係數不是巧合：五重進位鏈在還沒有任何位元組溢位之前，
就是計數器的五重前綴和，而前綴和的封閉形式正是 `C(n+4,5)`。
這同時是對「這五個位元組真的從零開始」的獨立佐證——只要初值不是全零，
第一個輸出就不會是 1。

## 5. 對 remake 的意義

- 擲骰語意可以照抄：`roll(N)` ＝ 1..N 等機率、`d6()` ＝ 1..6 等機率。
- **不要用現代 RNG 換掉它**。原版的分佈是拒絕取樣得來的乾淨均勻分佈，
  但序列本身是決定性的；若要做「重播同一場戰鬥」這類驗證，
  必須連狀態、`AL` 回饋與呼叫次數一起復現。
- 對拍驗證的作法：把 `Rng` 的狀態設成與原版同一時刻相同，
  比對之後 N 次呼叫的輸出序列。
- 熵來自鍵盤輪詢次數這件事，remake 要自己決定怎麼處理——
  照抄會讓「開機後不碰鍵盤直接進戰鬥」每次結果都一樣。這是設計決策，
  等規格階段再定，不在這裡下結論。

## 6. 可重跑的完整指令

```bash
# 1. 逐指令倒出整個 CODE 區（20,177 條）
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_listing.py workplace/analysis/dumps/listing.json

# 2. 直接定址的記憶體存取（827 個全域、4,932 筆存取）
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_memops.py workplace/analysis/dumps/memops.json

# 3. 四支函式的完整倒出
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/rng-family.json \
  0x18E6B 0x18E5F 0x18E41 0x19D86 0x19C84 --callers

# 4. 參考模型自我測試
python3 tools/rng.py
```

`ds:` 位移換線性位址：`seg002` 起點 `0x1CE20` ＋ 位移
（校正依據見 `tools/ida/export_memops.py` 的 `calibration_samples`）。

## 7. 這一輪學到的（寫成規則）

- **xref 圖不能拿來問「誰在寫這個全域」，除非先確認 xref 真的建起來了。**
  這份資料庫 4,932 筆直接定址存取裡只有 22 筆進了 xref 圖。
  下「只有一個地方寫」這種結論之前，先用另一條路徑數一次總量做正對照。
- **排除一種指令不等於排除一類機制。** 上一輪排掉 `mul` 是對的，
  但「RNG 應該是移位／加法型」這個推測也不完全對——它是**純加法**的，
  一個 `shl` 都沒有。用形狀（讀寫同一全域）找比用指令找可靠。

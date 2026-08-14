# 15：角色記錄的定址與欄位

日期：2026-08-14 ｜ 對應盤點 **C3**（角色記錄結構）、部分 **C4**

輸入：`wl.merged.exe`（解包映像 ＋ `wla.bin` overlay，本專案合成），
SHA-256 `cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

---

## 1. 定址：兩層

角色記錄不是用固定位址存取的，是兩層間接。兩支函式各負責一層。

### 1.1 `sub_17208`：算出隊伍槽表的位址（28 個呼叫端）

```
0x17208  8ad0        mov  dl, al           ; al ＝ 哪一組
0x1720A  8bf2        mov  si, dx
0x1720C  32ed        xor  ch, ch
0x1720E  8b8c39b2    mov  cl, [si-4DC7h]   ; 查表（ds:B239h）
0x17212  b83171      mov  ax, 7131h
0x17215  03c1        add  ax, cx
0x17217  a3b746      mov  ds:46B7h, ax
0x1721A  c3          retn
```

`ds:B239h`（線性 `0x28059`）的表值是 `00 0E 1C 2A`——**間隔 14**。
所以 `ds:46B7h` 指向 `0x7131` 起、每組 14 bytes 的**隊伍槽表**，
表裡每個 byte 是一個角色編號。

### 1.2 `sub_19614`：算出角色記錄的位址（43 個呼叫端，全檔最多）

```
0x19614  push ax
0x19615  mov  al, ds:4654h        ; 目前是哪一組
0x19618  call sub_17208           ; → ds:46B7h ＝ 隊伍槽表
0x1961B  pop  ax
0x1961C  mov  bl, al              ; 第幾個槽
0x1961E  mov  di, ds:46B7h
0x19622  mov  al, [bx+di]         ; 槽裡的角色編號 n
0x19624  8ae0  mov  ah, al
0x19626  32c0  xor  al, al        ; ax ＝ n × 256
0x19628  053171  add  ax, 7131h
0x1962B  a3b546  mov  ds:46B5h, ax  ; ← 角色記錄基址
0x1962E  b318    mov  bl, 18h
0x19630  mov  di, ds:46B5h
0x19634  mov  al, [bx+di]
0x19636  a20b47  mov  ds:470Bh, al  ; 順手取出 +0x18
0x19639  c3      retn
```

**角色記錄 ＝ `0x7131 ＋ 角色編號 × 256`，每筆 256 bytes。**
`0x7131` 正是存檔載入的目標位址（`docs/re/09` §4），所以編號 0 那 256 bytes
就是隊伍槽表所在的區塊，真正的角色記錄從編號 1（`0x7231`）開始。

推論等級：**已確認**（兩支函式讀完，表值從映像讀出）。

## 2. 欄位表

`ds:46B5h` 當基底、以 `mov bl, <常數>` ＋ `[bx+di]` 存取的位移，
由 `tools/summarize_record_fields.py` 掃出（工具刻意保守，
**22 個位移解出、141 處存取；另有 192 處因為位移是算出來的而追不到**——
所以下表是下界，不是全部）。

| 位移 | 大小 | 語意 | 依據 | 等級 |
|---|---|---|---|---|
| `+0x00` | 字串 | **名字** | `sub_171B9` 把 `ds:46B5h/46B6h` 直接搬進字串指標 `ds:4680h/4681h`，也就是從記錄開頭當 NUL 結尾字串印 | 已確認 |
| `+0x15`–`+0x17` | 24-bit | **可花用的 24-bit 計數器**（先比大小、不夠就失敗；語意像金錢但未由文字直接證實，見 `docs/re/19` §6） | `sub_17B3E` 用 `cmp`＋兩個 `sbb` 比大小、`sub_17B15` 用 `sub`＋兩個 `sbb` 扣款，對象是 `ds:466Bh`–`466Dh` | 已確認 |
| `+0x18` | byte | **性別**（0 ＝ Male、1 ＝ Female）。`sub_19362` 印 `'SEX: '` 之後做 `add al, 0A2h`，正好指到字串 162／163（`docs/re/17` §4.3）。`sub_19614` 每次算完記錄就把它取出放 `ds:470Bh` | 已確認 |
| `+0x19` | byte | **國籍**（0–4）。`sub_19362` 印 `'NATIONALITY: '` 之後做 `add al, 91h`，指到字串 145–149 ＝ `U.S.`／`Russian`／`Mexican`／`Indian`／`Chinese` | 已確認 |
| `+0x1A` | byte | **AC（護甲等級）**——名片行表頭 `'   NAME        AC AMM MAX CON WEAPON '` 的第二欄（`docs/re/17` §4.2） | 強證據 |
| `+0x1B`–`+0x1C` | 16-bit | **最大體力（MAXCON）** | `sub_19A41` 拿它跟 `+0x1D/+0x1E` 逐 byte 比相等（＝滿血判定）；名片行先印它 | 強證據 |
| `+0x1D`–`+0x1E` | 16-bit 有號 | **目前體力（CON）** | `sub_172AE`（10 個呼叫端）把兩個 byte `or` 起來測零＝死亡判定；`+0x1E` 為負時名片行改印狀態字 | 強證據 |
| `+0x1F` | byte | **裝備索引**：`shl 1 ＋ 0xBC` 之後指進**物品**陣列。名片行的 `AMM`（彈藥）欄印的就是它指到的那個 byte | 強證據 |
| `+0x21`–`+0x23` | 24-bit | 另一個 24-bit 計數器，只有「加」的路徑（`sub_19BC0`）。語意未解 | 強證據（結構） |
| `+0x26`–`+0x27` | 16-bit | 受傷前的 CON 備份（`sub_157D6` 扣血時寫入） | 強證據 |
| `+0x28` | byte | 非 0 時名片行開反白 | 假說 |
| `+0x80`–`+0xBB` | 30 × 2 | **技能**：偶數位移 ＝ 技能編號、奇數位移 ＝ 等級 | 已確認 |
| `+0xBD`–`+0xF8` | 30 × 2 | **物品**：奇數位移 ＝ 物品編號、偶數位移 ＝ 附屬 byte | 已確認 |

### 2.1 兩個陣列是怎麼分辨出來的

兩者的掃描形狀一模一樣（都是「起點每次 +2 掃到終點」），光看位置分不出來。
分辨的依據是**誰印哪個表頭**：

| 函式 | 印的表頭（`docs/re/17` 的字串編號） | 走的陣列 |
|---|---|---|
| `sub_1997B` | `0x99` ＝ 153 ＝ `'   LVL   SKILL'` | `+0x80` 起 |
| `sub_19454` → `sub_1963A` | `0x97` ＝ 151 ＝ `'    ITEM'` | `+0xBD` 起 |

`sub_1997B` 還帶一個值域檢查，把它釘死：

```
0x1998A  b380      mov  bl, 80h
0x19990  8b3eb546  mov  di, ds:46B5h
0x19994  8a01      mov  al, [bx+di]
0x19996  84c0      test al, al
0x19998  740d      jz   short loc_199A7    ; 0 ＝ 空槽
0x1999A  3c24      cmp  al, 24h            ; ← 技能編號上限 36
0x1999C  f5        cmc
0x1999D  7308      jb   short loc_199A7
```

**`0x24` ＝ 36，而執行檔裡正好有 35 個技能名**（編號 1–35，`docs/re/17` §4）。
物品名是編號 36–130，值域完全不同。

`sub_198CD`（搜尋技能）掃 `0x80`–`0xBA`（30 槽），
但 `sub_1997B`（顯示）掃到 `0xBC`——**多一槽**，而 `0xBC` 已經是物品陣列前一個 byte。
兩個迴圈的邊界在原版就不一致，照抄時要注意。

角色記錄 `+0x1F` 的裝備索引 `n` 指到 `0xBC + 2n`，落在**物品**陣列上——
裝備的是武器，武器是物品，這與上面的分辨互相印證。

## 3. 傷勢等級

`sub_19A1D`（4 個呼叫端）把 CON 換成 0–5 的傷勢等級：

```
0x19A1D  mov  dl, 5
0x19A1F  mov  bl, 1Eh
0x19A21  mov  di, ds:46B5h
0x19A25  mov  al, [bx+di]
0x19A27  dec  bl
0x19A29  or   al, [bx+di]
0x19A2B  jz   short locret_19A40    ; CON ＝ 0 → 回傳 5
0x19A2D  mov  al, [bx+di]           ; CON 低位
0x19A2F  dec  dl
0x19A31  mov  si, dx
0x19A33  cmp  al, [si-3333h]        ; 查門檻表
0x19A37  cmc
0x19A38  jnb  short locret_19A40
0x19A3A  jz   short locret_19A40
0x19A3C  dec  dl
0x19A3E  jnz  short loc_19A31
```

門檻表在 `ds:CCCEh`（線性 `0x29AEE`），四個 byte：

| bytes | 當有號數 |
|---|---|
| `F5 EC E2 D8` | **−11、−20、−30、−40** |

CON 會**掉到負的**，負得越多等級越差。等級再去查 `ds:B233h`（線性 `0x28053`）
的訊息碼表 `85 9A 9B 9C 9D 84`（6 個 byte，索引 0–5），
名片行就印對應的狀態字（`sub_1708B` 的 `0x17142`）。

訊息碼對應的字串（`docs/re/17` §4.4）：

| 等級 | 訊息碼 | 字串 |
|---:|---|---|
| 0 | `0x85` | `UNC`（unconscious） |
| 1 | `0x9A` | `SER`（serious） |
| 2 | `0x9B` | `CRT`（critical） |
| 3 | `0x9C` | `MRT`（mortal） |
| 4 | `0x9D` | `COM`（coma） |
| 5 | `0x84` | `0x7F` 那個骷髏字模（死亡） |

推論等級：**強證據**（程式碼讀完、兩張表從映像讀出、狀態字已對到字串表）。

## 4. 名片行（`sub_1708B`）

隊伍名單一行的排版，把上面幾個欄位串起來，也順便釘住了欄位的顯示順序：

表頭字串（`docs/re/17` 索引 136）是 `'   NAME        AC AMM MAX CON WEAPON '`，
欄位順序逐一對上：

| 欄（`ds:4672h`） | 表頭 | 內容 |
|---|---|---|
| — | NAME | 名字（`sub_171B9`） |
| `0x11` | AC | `+0x1A` |
| `0x15` | AMM | `+0x1F` 指到的物品項目（`and al, 3Fh` 之後） |
| `0x18` | MAX | `+0x1B/+0x1C`（MAXCON） |
| `0x1C` | CON | `+0x1D/+0x1E`；為負或為零時改印狀態字 |
| `0x20` | WEAPON | `sub_196C9` 取回的裝備 |

`ds:4672h` 是文字欄位計數器（`docs/re/14` §4），所以這些是**字元欄位座標**，
不是像素座標。中文化重排版面要從這裡開始。

## 5. 可重跑的完整指令

```bash
# 欄位掃描（純 Python，吃 export_listing.py 的輸出）
python3 tools/summarize_record_fields.py workplace/analysis/dumps/listing.json \
  > docs/re/generated/ida94/record-fields.json

# 相關函式完整倒出
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/character.json \
  0x17208 0x19614 0x1708B 0x172AE 0x19A1D 0x19A41 0x17B15 0x17B3E 0x198CD 0x1968C --callers
```

## 6. 這一輪學到的（寫成規則）

- **記錄結構的欄位不會出現在固定位址掃描裡。** 它們是
  `mov bl, <常數>` ＋ `[bx+di]`，`export_memops.py` 一筆都看不到。
  要另外做一次極小範圍的資料流（`summarize_record_fields.py`）。
- **保守的掃描要把「追不到的」數出來。** 這次 `ds:46B5h` 有 128 處解出、
  127 處追不到，兩個數字一起看才知道欄位表是下界而不是全貌。
  只印解出的那 23 個位移，會讓人以為記錄就這麼大。

# 109：角色記錄的 `+0x4D`–`+0x7F` 沒有任何存取點，階級字串沒有長度上限

日期：2026-08-17 ｜ 接 [`15`](15-character-record.md)（記錄佈局）、
[`31`](31-experience-and-skills.md)（階級表）、[`96`](96-ending.md) §5（`+0x4B`／`+0x4C` 兩個旗標）、
[`docs/spec/05`](../spec/05-character-and-save.md) §4

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
出廠存檔 `game1`／`game2`（checksum 逐份驗過）。

問題：角色記錄 256 bytes 裡，`+0x4D`–`+0x7F` 那 51 bytes 是什麼？

答案：**沒有任何程式碼讀它或寫它**，出廠也全是零。它是階級字串與技能陣列
中間的空地。順帶解掉一個一直寫錯的欄位大小：**階級字串沒有長度上限**，
不是 14 bytes。

---

## 1. 方法：欄位位移放 `bl`、記錄基底放 `di`

原版存取記錄的慣用寫法是：

```
0x1B4FB  mov  bl, 4Bh          ; 欄位位移
0x1B4FD  mov  di, ds:46B5h     ; ← 記錄基底
0x1B501  mov  al, [bx+di]      ; 讀 +0x4B
```

`ds:46B5h` 是**目前這個人的記錄位址**，由 `sub_19614` 算出來：

```
0x19614  al ← ds:4654h（目前隊員序號）→ sub_17208
0x1961C  bl ← al；di ← ds:46B7h        ; 隊伍槽表
0x19622  al ← [bx+di]                  ; 槽 → 角色編號
0x19624  ah ← al；al ← 0               ; ax ＝ 編號 × 256
0x19628  ax += 7131h                   ; ＋ 記錄區起點
0x1962B  ds:46B5h ← ax
```

⚠ **基底一定要跟著看。** 同一個 `[bx+di]` 慣用法在角色記錄、敵方記錄、
腳本記錄上長得一模一樣。只看位移的話，敵方記錄（`ds:46C8h`，94 bytes ＝ `0x5E`，
[`37`](37-enemy-records-and-hp.md)）的 `+0x5D` 會被算成角色記錄的欄位——
而 `0x5D` 正好落在這一輪要回答的區間裡。

盤點工具是既有的 `tools/summarize_record_fields.py`（`docs/re/15` 那一輪寫的，
純 stdlib，讀離線指令清單）。它刻意保守：只認「常數載入 ＋ `inc`／`dec`」這條鏈，
往前跳的目標清空追蹤狀態，追不到的算進 `unresolved`。

## 2. 結果：`ds:46B5h` 用到的位移

工具報出 22 個位移、141 處存取、**192 處追不到**：

```
0x00 0x0D 0x0E 0x0F 0x11 0x13 0x14 0x15 0x16 0x17 0x18 0x19 0x1A 0x1B 0x1C
0x1D 0x1E 0x1F 0x80 0xBB 0xBD 0xF8
```

追得到的部分裡 `0x4D`–`0x7F` 一筆都沒有。但 **192 處追不到，所以工具的輸出
本身不能當成「沒有」**——下面三層才是。

### 2.1 追不到的那 192 處

位移是算出來的或由輔助函式回傳。逐支追來源，全部落在兩個陣列或既有欄位：

| 輔助函式 | 回傳的位移 |
|---|---|
| `sub_19AC8` | `shl al,1；add al,0BBh` ＝ 索引 × 2 ＋ `0xBB` → **物品陣列**（`0xBD`–`0xF9`）|
| `sub_1968A` | 從 `0xBD` 起 `inc bl` 兩次一格掃到 `0xF9`，找不到回 `0xFF`（哨兵，不是位移）|
| `sub_196C9` | `0x1F`（裝備索引）|
| `sub_198CD`／`sub_1997B`／`sub_198F0` | `0x80`／`0xBB`／`0xBC`／`0xBE` → **技能陣列** |
| `sub_19614` | `0x18` |

> ⚠ **`call` 不能當成 `bl` 的障礙，兩個方向都會錯。**
> 有些輔助函式**用 `bl` 回傳位移**（上表），這時 `call` 前面那個 `mov bl, <常數>`
> 是給被呼叫者的參數——不擋就會憑空生出欄位：`0x12F94` 的 `mov bl,4Ah` 與
> `0x158BA` 的 `mov bl,51h` 其實都是 `sub_1393E` 的參數（那支查的是 `ds:46B0h`），
> 誤報出來會變成角色記錄的 `+0x4A`／`+0x51`，而兩個位置都落在已知欄位附近，
> **看起來完全合理**。
> 反過來，有些**保留 `bl`**：建立角色擲屬性那一圈是
> `mov bl,0Eh` → `call sub_1CAD1` → 寫 `[bx+di]` → `inc bl` → `cmp bl,15h`，
> 擋掉 `call` 會讓 `+0x0E`（七個屬性的起點）整個消失。
> 所以工具照報，**跨 `call` 的位移一律人工判讀**。

### 2.2 `inc bl` 走得到哪裡

會往上走的迴圈都有界：屬性 `cmp bl,15h`、技能 `cmp bl,0BCh`／`0BEh`、
物品 `cmp bl,0F9h`；其餘只 `inc` 一兩次（16-bit 欄位的高低位元組）。
沒有一個從 `0x4D` 以下走進 `0x4D`–`0x7F`。

**唯一沒有界的往上寫是階級字串**（§4）——它的計數器不在 `bl` 而在 `ds:D430h`，
所以上面的掃描看不到它。以出貨的字串表來說它停在 `+0x46`。

### 2.3 全檔立即數掃描

整份映像裡 `mov bl, imm` 的立即數落在 `0x4D`–`0x7F` 的**只有 5 處**，
逐處確認都不是角色記錄：

| 位址 | 立即數 | 實際上是什麼 |
|---|---|---|
| `0x119EF`／`0x11B1E` | `0x71` | 與 `mov bh, …` 合成一個 16-bit 值，不是位移 |
| `0x138B0` | `0x5D` | 清 `ds:46C8h`（**敵方**記錄）的 `0x00`–`0x5D` |
| `0x14BC9` | `0x5D` | 對調兩筆敵方記錄的 `0x00`–`0x5D` |
| `0x158BA` | `0x51` | `sub_1393E` 的參數，那支查的是 `ds:46B0h` |

推論等級：**已確認**（`0x4D`–`0x7F` 的「沒有」有三層佐證：全檔立即數掃描、
`inc bl` 迴圈的界、輔助函式逐支追來源。⚠ 例外寫在 §2.2：階級字串那一圈
**結構上走得進去**，只是出貨的字串表不夠長）。

## 3. 出廠資料：`0x39`–`0x7F` 全零

`game1`／`game2` 的 7 筆角色記錄（checksum 都驗過）：

| 位移 | 內容 |
|---|---|
| `0x32`–`0x38` | `PRIVATE`（四個出廠角色都是）|
| `0x39`–`0x7F` | **全零**，四個出廠角色與三個空槽都一樣 |

## 4. 階級字串沒有長度上限

寫階級的是 `0x1BB6C`，逐字元抄到 NUL 為止，**沒有任何長度檢查**：

```
0x1BB6C  cx ← 0D622h；ds:4692h ← cx      ; 階級字串表（64 條）
0x1BB73  sub_178B9                        ; 開始解一條字串
0x1BB76  bl ← 32h；ds:D430h ← bl          ; 目的位移從 +0x32 起
0x1BB7C  sub_17B8F → al                   ; 下一個字元
0x1BB7F  bl ← ds:D430h；di ← ds:46B5h
0x1BB87  [bx+di] ← al                     ; 寫進記錄
0x1BB89  al == 0 → retn                   ; 寫到 NUL 為止
0x1BB8D  inc byte ptr ds:D430h；jnz 回圈   ; 位移 +1（只有繞回 0 才會停）
```

所以**階級欄的長度由字串決定**。實際最長的階級是 `Lieutenant Commander`
（20 字元），寫完 NUL 落在 `+0x46`；下一個已知欄位是 `+0x4B`，
所以這一欄的安全範圍是 **`+0x32`–`+0x4A`（25 bytes）**。

> ⚠ 這一條推翻了「階級欄 14 bytes」。14 是從出廠值 `PRIVATE` 反推的，
> 而出廠值是最短的那一個。**用出廠資料量欄位大小，量到的是下限不是大小。**

## 5. 記錄的完整佈局（這一輪之後）

| 位移 | 內容 |
|---|---|
| `0x00`–`0x0D` | 名字（NUL 結尾）|
| `0x0E`–`0x31` | 屬性、金錢、性別、國籍、AC、CON、裝備、技能點、經驗值、等級、狀態…… |
| `0x32`–`0x4A` | **階級字串**（NUL 結尾，最長 `Lieutenant Commander` 到 `+0x46`）|
| `0x4B` | 參與過摧毀 Base Cochise（bit0，[`96`](96-ending.md) §5）|
| `0x4C` | 總部已經表揚過（bit0）|
| **`0x4D`–`0x7F`** | **沒有存取點、出廠全零（51 bytes）**。⚠ 階級字串長到 25 個字元以上就會寫進來，出貨的字串表最長 20 |
| `0x80`–`0xBB` | 技能 30 × 2 |
| `0xBD`–`0xF8` | 物品 30 × 2 |
| `0xF9`–`0xFF` | 沒有存取點（7 bytes）|

## 6. 這個結論能用來做什麼

`0x4D`–`0x7F` 那 51 bytes 是重製版**唯一一塊可以寫東西又不影響原版的空間**
（例如放完整的 UTF-8 名字，見 `docs/utf8-experiment.md` §4）。要用它之前
還缺一個驗證：**把資料寫進去、用原版跑一輪、存檔、再讀回來**，
確認原版不會清掉它。原版沒有寫入點，但建立角色時那筆記錄的內容從哪來
（是否沿用被刪角色的殘留）還沒讀。**沒做這個實驗之前，不要把它當成保證。**

## 7. 這一輪學到的（寫成規則）

- **`[bx+di]` 這種慣用法要連基底一起比對，只看位移會把別的結構算進來。**
  第一版盤點在「角色記錄」下報出 `+0x5D`，那其實是敵方記錄的最後一個 byte。
- **跨 `call` 的 `bl` 沒有通則，只能人工判讀。** 擋 `call` 會漏掉真欄位
  （`+0x0E`），不擋會生出假欄位（`+0x4A`、`+0x51`）——而假欄位看起來完全合理：
  落在已知欄位附近、位址也在角色相關的函式裡。**兩個方向都會錯的規則，
  不要挑一個方向寫死**，要把它標成待判讀。
- **要寫掃描工具之前先看 `tools/` 有沒有。** 這一輪的問題早就有工具
  （`summarize_record_fields.py`，`docs/re/15` 那一輪寫的），而且它處理了兩件
  重寫的版本不會想到的事：IDA 沒建成函式的兩成程式碼要以 `retn` 切成偽區塊
  （否則整批漏掃、而且是安靜的），以及**往前跳的目標要清空追蹤狀態**
  （兩條互斥分支的 `inc bl` 會被線性掃描疊加成不存在的位移）。
- **用出廠資料量欄位大小，量到的是下限。** 階級欄被寫成 14 bytes，
  是因為出廠值 `PRIVATE` 只有 7 個字元；真正的上限要看**寫入端有沒有邊界檢查**，
  而這一支沒有。
- **「沒有人用」要三層佐證**：靜態立即數全檔掃描、算出來的位移逐支追來源、
  出廠資料實際內容。少一層都會變成「我沒找到」而不是「沒有」。

## 8. 可重跑的完整指令

```bash
# 欄位盤點（角色記錄的基底是 46b5；不給參數會掃七個基底）
docker run --rm --log-opt max-size=10m --log-opt max-file=3 --network none \
  -v "$PWD:/w" -w /w -u "$(id -u):$(id -g)" python:3.12-slim \
  python3 tools/summarize_record_fields.py workplace/analysis/dumps/listing.json 46b5

# 全檔掃 mov bl, imm 落在 0x4D–0x7F 的（連上下文一起印，逐處判讀）
docker run --rm --log-opt max-size=10m --log-opt max-file=3 --network none \
  -v "$PWD:/w" -w /w -u "$(id -u):$(id -g)" python:3.12-slim python3 -c "
import json,re
ins=json.load(open('workplace/analysis/dumps/listing.json'))['instructions']
pat=re.compile(r'^mov\s+(bl|bx),\s*([0-9A-F]+)h?\b')
for k,i in enumerate(ins):
    m=pat.match(i['disasm'].strip())
    if not m: continue
    v=int(m.group(2),16)
    if not (0x4D<=v<=0x7F): continue
    print('---', i['ea'], i['disasm'].strip(), i.get('func'))
    for j in range(k+1,k+5): print('   ', ins[j]['ea'], ins[j]['disasm'].strip())"

# 階級字串的寫入迴圈
docker run --rm --log-opt max-size=10m --log-opt max-file=3 --network none \
  -v "$PWD:/w" -w /w -u "$(id -u):$(id -g)" python:3.12-slim python3 -c "
import json; d=json.load(open('workplace/analysis/dumps/listing.json'))
for i in d['instructions']:
    ea=int(i['ea'],16)
    if 0x1BB6C<=ea<0x1BB94 or 0x19614<=ea<0x1963A: print(hex(ea), i['disasm'])"
```

出廠資料那一段（解密 ＋ 驗 checksum ＋ 列非零位移）在
`internal/game/recordtail_test.go` 裡跑，不必手動重跑。

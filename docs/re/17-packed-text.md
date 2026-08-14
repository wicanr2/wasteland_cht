# 17：5-bit 打包文字與執行檔的字串表

日期：2026-08-14 ｜ 對應盤點 **E4**（遊戲內短訊息來源）、**D1**（技能表）、**C3** 的數個欄位

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

---

## 1. 結論

執行檔裡的遊戲文字**不是明文 ASCII**，是 **5-bit 打包 ＋ 字元對照表**。
先前在資料段掃 ASCII 只掃得到選單與錯誤訊息，因為真正的文字表根本不長那樣。

執行檔裡有 **九張這樣的表**，合計 **444 個字串槽、426 條非空**——包含技能、物品、階級、
疾病、介面訊息，以及**長篇的結局敘述**。
地圖區塊另有各自的表，共 4,445 個槽、4,401 條非空（[`docs/re/18`](18-block-text.md)）。

位移表一項涵蓋四個字串，所以**最後一組不一定用滿**——沒用到的槽解出來是雜訊。
字串靠編號定址，槽要原樣保留，統計時才把「槽數」與「非空條數」分開講。

## 2. 格式

字串表的基址由 `ds:4692h` 指定（`sub_17ACE` 對執行檔內的表寫死 `0xB270`，
線性 `0x28090`）。佈局：

```
表基址 +0x00 … +0x3B     60 bytes 字元對照表（符號 → ASCII，0 ＝ 字串結束）
表基址 +0x3C …           16-bit 位移表，每 4 個字串一項，位移相對於 +0x3C
位移指到的地方            5-bit 符號流
```

位移表沒有長度欄位，但**第一項的值就是資料起點**，所以「第一項 ÷ 2」＝ 表的項數
（這裡是 44）。最後一項是沒用到的填充，值落在區塊外。

### 2.1 位元讀取（`sub_17BC7`）

```
0x17BC7  b305        mov  bl, 5              ; 一個符號 5 個位元
0x17BCB  8a0ebb46    mov  cl, ds:46BBh       ; 剩餘位元數
0x17BCF  84c9        test cl, cl
0x17BD1  791d        jns  short loc_17BF0
0x17BD3  b207        mov  dl, 7              ; ┐ 補一個 byte
0x17BDF  8b3e8046    mov  di, ds:4680h       ; │
0x17BE3  8a01        mov  al, [bx+di]        ; │
0x17BE5  ff068046    inc  word ptr ds:4680h  ; ┘
0x17BF0  d02ebd46    shr  byte ptr ds:46BDh, 1  ; 取最低位
0x17BF4  d0d8        rcr  al, 1
0x17BFA  fecb        dec  bl
0x17BFC  75cd        jnz  short loc_17BCB
0x17BFE  d0e8        shr  al, 1  ×3
```

**每個 byte 由最低位往最高位取**，先讀到的位元是符號的第 0 位。
`ds:46BBh` 初值 `0xFF`（負數）＝「手上沒有位元」，所以第一次就會補 byte。

### 2.2 符號到字元（`sub_17B8F`）

```
0x17B8F  f8          clc
0x17B90  d01ebc46    rcr  byte ptr ds:46BCh, 1   ; 大寫旗標移位
0x17B94  e83000      call sub_17BC7
0x17B97  3c1e        cmp  al, 1Eh
0x17B99  f9          stc
0x17B9A  74f4        jz   short loc_17B90        ; 0x1E → 設旗標、重取
0x17B9C  3c1f        cmp  al, 1Fh
0x17B9E  7505        jnz  short loc_17BA5
0x17BA0  e82400      call sub_17BC7              ; 0x1F → escape
0x17BA3  041e        add  al, 1Eh                ;        值 ＋0x1E
0x17BA5  8ad8        mov  bl, al
0x17BA7  8b3ea146    mov  di, ds:46A1h           ; 字元對照表
0x17BAB  8a01        mov  al, [bx+di]
0x17BB8  740c        jz   short locret_17BC6
0x17BBA  3c61        cmp  al, 61h                ; ┐ 旗標成立且是小寫
0x17BC4  2c20        sub  al, 20h                ; ┘ → 轉大寫
```

- 符號 `0x1E` ＝**下一個字元轉大寫**（不是字元本身）
- 符號 `0x1F` ＝ escape，再取一個符號後 `+0x1E` → 字元集共 **60 個**
- 對照表值為 0 ＝ 字串結束

### 2.3 取第 N 個字串（`sub_178B9`）

```
0x178D5  d0e8        shr  al, 1
0x178D7  24fe        and  al, 0FEh       ; ＝ (N >> 2) × 2
0x178DB  8b3e8046    mov  di, ds:4680h   ; ＝ 表基址 + 0x3C
0x178DF  8a01        mov  al, [bx+di]    ; 讀位移（低位）
0x178E1  02068046    add  al, ds:4680h   ; ＋ 基址
…
0x178F6  2403        and  al, 3
0x178F8  7410        jz   short locret_1790A
0x178FD  e88f02      call sub_17B8F      ; 跳掉 N & 3 個字串
```

位移表是**每 4 個字串一項**，餘數靠實際解碼跳過。

推論等級：**已確認**（三支函式讀完，實作在九張表上解出 442 個可讀字串）。

## 3. 九張表

表基址存在 `ds:4692h`。全檔 13 個寫入點裡，9 個寫的是常數（下表），
另外 4 個不是常數：`sub_1790B` 寫的是 MSQ 區塊標頭 `+0x00`（隨地圖而變，見 §6），其餘三處待追。
**找表不要掃資料，要問「誰寫 `ds:4692h`」**——掃資料完全掃不到，因為文字是打包的。

| `ds:` 基址 | 設定者 | 可用字串 | 內容 |
|---|---|---:|---|
| `0xA703` | `sub_16390` | 20 | 開場字幕與製作名單（`Written by Alan Pavlish`、`IBM version by Michael Quarles`） |
| `0xAB3E` | `sub_16CBD` | 108 | 無線電、隊伍操作、戰鬥訊息 |
| `0xB270` | `sub_17ACE`（18 個呼叫端） | 170 | 技能、物品、角色名片欄位 |
| `0xCE4B` | `sub_1A3E1` | 12 | 角色建立 |
| `0xD18E` | `0x1B7C2` | 16 | **結局敘述**（見 §4.6） |
| `0xD622` | `sub_1BB5D` | 64 | 階級（`Private`、`Private 1st class`、`Specialist`…） |
| `0xDACC` | `sub_1BE31` | 12 | 技能學習 |
| `0xDBF8` | `sub_1C213` | 12 | 商店 |
| `0xDCED` | `sub_1C561` | 28 | 疾病與狀態（`Radiation poisoning`、`Wasteland Herpes`、`Bug byte`） |

`0x1B7BF`（載入 `0xD18E` 的那條指令）正是 `docs/re/12` §5 列為「未追」的兩個
引用點之一——它不是頻率表的引用，是結局敘述表的基址。

## 4. `0xB270` 表的內容

`python3 tools/decode_text.py workplace/analysis/unpacked/wl.merged.exe`

| 索引 | 內容 |
|---|---|
| 1–35 | **技能名**：`Brawling`、`Climb`、`Clip pistol`、`Knife fight`、`Pugilism`、`Rifle`、`Swim`、`Knife throw`、`Perception`、`Assault rifle`、`AT weapon`、`SMG`、`Acrobat`、`Gamble`、`Picklock`、`Silent move`、`Combat shooting`、`Confidence`、`Sleight of hand`、`Demolitions`、`Forgery`、`Alarm disarm`、`Bureaucracy`、`Bomb disarm`、`Medic`、`Safecrack`、`Cryptology`、`Metallurgy`、`Helicopter pilot`、`Electronics`、`Toaster repair`、`Doctor`、`Clone tech`、`Energy weapon`、`Cyborg tech` |
| 36–130 | **物品名**（武器、彈藥、護甲、雜物、任務道具） |
| 131–170 | 介面訊息、狀態縮寫、性別與國籍 |

### 4.1 單複數是資料裡的機制

物品名長這樣：

```
'Ax\n\nes\n'          → Ax   / Axes
'Kni\nfe\nves\n'      → Knife / Knives
'Plastic explosive\n\ns\n'
```

`\n`（`0x0A`）在字串裡是分隔控制碼：**字根 ／ 單數字尾 ／ 複數字尾**。
中文化時這個機制可以直接拿掉（中文沒有單複數），但**翻譯工具必須認得它**，
否則會把 `\n` 當成換行塞進譯文。

### 4.2 名片行的欄位名，直接對上 `docs/re/15`

索引 136：

```
'   NAME        AC AMM MAX CON WEAPON \x02'
```

`sub_1708B` 依序在欄位座標 `0x11`、`0x15`、`0x18`、`0x1C`、`0x20` 印五個值，
順序與這個表頭一致，所以角色記錄的欄位是：

| 欄 | 角色記錄位移 | 語意 |
|---|---|---|
| AC | `+0x1A` | 護甲等級 |
| AMM | `+0x1F` 指到的項目 | 彈藥 |
| MAX | `+0x1B`–`+0x1C` | 最大體力 |
| CON | `+0x1D`–`+0x1E` | 目前體力 |
| WEAPON | `sub_196C9` 取回 | 裝備中的武器 |

### 4.3 性別與國籍：算術完全吻合

`sub_19362` 印角色資料時：

```
0x19362  mov  al, 89h   ; 字串 137 ＝ 'NAME: '
0x1936D  mov  al, 8Ah   ; 字串 138 ＝ 'SEX: '
0x19372  mov  bl, 18h   ; 讀角色記錄 +0x18
0x1937A  add  al, 0A2h  ; ＋162 → 字串 162/163 ＝ 'Male'/'Female'
0x1937F  mov  al, 8Bh   ; 字串 139 ＝ 'NATIONALITY: '
0x19384  mov  bl, 19h   ; 讀角色記錄 +0x19
0x1938C  add  al, 91h   ; ＋145 → 字串 145–149
```

字串 145–149 是 `U.S.`、`Russian`、`Mexican`、`Indian`、`Chinese`。

**角色記錄 `+0x18` ＝ 性別（0／1）、`+0x19` ＝ 國籍（0–4）**，
標籤字串的編號與加法常數逐一對上，沒有例外。推論等級：**已確認**。

### 4.4 傷勢等級的字串

`docs/re/15` §3 解出的訊息碼表 `85 9A 9B 9C 9D 84` 現在有內容了：

| 等級 | 訊息碼 | 字串 |
|---:|---|---|
| 0 | `0x85` | `UNC`（unconscious） |
| 1 | `0x9A` | `SER`（serious） |
| 2 | `0x9B` | `CRT`（critical） |
| 3 | `0x9C` | `MRT`（mortal） |
| 4 | `0x9D` | `COM`（coma） |
| 5 | `0x84` | `0x7F` 那個骷髏字模（死亡） |

CON 越負等級越高（門檻 −11／−20／−30／−40），CON 恰為 0 走等級 5。

### 4.5 文字控制碼在字串裡看得到

`docs/re/14` §4.1 解出的控制碼，在這批字串裡直接現身：

```
'\x07NAME: \x0b'                      ; 0x0B ＝ 插入角色名字
'\r\x0b doesn\'t want to trade.\r\x06'
'\r\x0b raised h\x0cis\x0cer\x0c '    ; 0x0C ＝ 依性別選 his/her
'\x01SER\x02'                          ; 0x01/0x02 ＝ 反白開／關
```

`0x0C` 的用法（`his` ／ `her` 兩段用 `0x0C` 夾起來）是先前只從程式碼看不出來的，
**這是中文化必須處理的一類**：中文沒有 his/her 的區別，但格式必須 round-trip。

### 4.6 結局敘述：原版**確實有**長篇文字

`0xD18E` 表的內容是完整的段落，不是短訊息：

> `Shuddering explosions rock the base,\rfire blossoms throughout every\rdoorway.\r\r`
>
> `\r  "But certainly most important of all were the rangers who gave their lives to
> destroy Base Cochise, the greatest threat that man has ever known."\r`
>
> `\rThe following is an excerpt from the dedication to The History of the Rangers,
> Vol. II, by Karl Allard, 2095, Allard Press, Desert Center, Hardbound pp ii, $10 gold.\r`

`\r`（`0x0D`）是段內換行。表的前面還有明文的地名 `Base Cochise`。

**這直接推翻 `docs/re/12` §4 的「原版遊戲檔案裡沒有長篇敘述文字」。**
那條結論的證據是「掃不到 ASCII」，而執行檔的文字本來就掃不到。

## 5. 對中文化的意義

- 執行檔內的遊戲文字要先解包才看得到，**不能用 hex editor 直接改**。
- 字元對照表只有 60 個符號，中文一定塞不進去；重製版走自己的文字管線，
  但**抽取原文要用這支解碼器**。
- 單複數（`\n`）與性別（`0x0C`）是資料裡的機制，翻譯格式要保留它們的位置。

## 6. 還沒解的

- ~~MSQ 區塊的文字~~ **已解**：見 [`docs/re/18`](18-block-text.md)。
  42 個區塊各有一張自己的字串表，共 4,445 個槽。之前解不出來是因為
  **把整個區塊都拿去 XOR**，而原版只解到標頭第一個 word 那個長度為止。
  區塊標頭 `+0x02` 指到的是**明文** NUL 分隔的敵人名表（`Biker Scum`、
  `Rib Cracker` …），`+0x04` 指到 8 bytes 一筆的記錄表。
- 段落書（`paragraphs.txt`）的 162 段與遊戲的關係：既然執行檔裡有長篇敘述，
  「遊戲只給編號、內容全在紙本」這個模型要重新檢查。

## 7. 可重跑的完整指令

```bash
python3 tools/decode_text.py workplace/analysis/unpacked/wl.merged.exe \
  docs/re/generated/ida94/exe-strings.json

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/text-codec.json \
  0x178A3 0x178B9 0x17B8F 0x17BC7 0x17B80 0x17ACE 0x19362 --callers
```

## 8. 這一輪學到的（寫成規則）

- **「掃不到 ASCII」只證明文字不是 ASCII 存的。** 1980 年代的遊戲把文字打包是常態
  （這裡是 5-bit ＋ 對照表，省下約 37%）。要下「沒有這類文字」的結論，
  得先找到印字串的程式碼、讀出它怎麼取字元，而不是掃資料。
- **入口在「印」那一端，不在資料那一端。** 這次是從
  「`sub_178A0` 印訊息」→ `sub_178A3` → `sub_17B8F` 逐跳追下來的，
  總共三跳。先前在資料段掃了很久都沒結果。
- **要列舉同一類資源，去找「指向它的那個變數是誰寫的」。** 用資料特徵盲掃
  60 bytes 的字元對照表，12 個命中裡全是重複字元造成的假陽性；
  改成掃「誰寫 `ds:4692h`」，一次拿到九張表，沒有一張是猜的。
